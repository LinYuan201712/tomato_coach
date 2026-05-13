package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
)

// SleepTool 模拟一个执行缓慢的工具
type SleepTool struct {
	SleepDuration time.Duration
}

func (t *SleepTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "sleep_tool", Desc: "执行耗时操作"}, nil
}

func (t *SleepTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	select {
	case <-time.After(t.SleepDuration):
		return "Success", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// AuthCheckTool 验证上下文中的 UserID
type AuthCheckTool struct {
	CapturedUserID int64
}

func (t *AuthCheckTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "auth_tool", Desc: "验证鉴权信息"}, nil
}

func (t *AuthCheckTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	uid, ok := ctx.Value("user_id").(int64)
	if !ok {
		return "FAILURE: No user_id in context", nil
	}
	t.CapturedUserID = uid
	return fmt.Sprintf("SUCCESS: UserID is %d", uid), nil
}

func TestToolHardening_Timeout(t *testing.T) {
	// 1. 设置一个执行 30s 的工具，预期会在 20s 被切断
	sleepTool := &SleepTool{SleepDuration: 30 * time.Second}
	
	// Mock 模型：第一次调用返回工具调用，第二次调用返回结束
	mockInner := &model.MockChatModel{}
	chatModel := model.NewChatModelFromInner(mockInner)
	
	count := 0
	mockInner.GenerateFunc = func(ctx context.Context, messages []*schema.Message, opts ...einomodel.Option) (*schema.Message, error) {
		count++
		if count == 1 {
			return &schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{ID: "call_1", Function: schema.FunctionCall{Name: "sleep_tool", Arguments: "{}"}},
				},
			}, nil
		}
		return &schema.Message{Role: schema.Assistant, Content: "完成"}, nil
	}

	msgs := []*schema.Message{schema.UserMessage("Test timeout")}
	
	start := time.Now()
	res, err := RunToolLoop(context.Background(), chatModel, msgs, []tool.BaseTool{sleepTool})
	duration := time.Since(start)

	t.Logf("RunToolLoop finished. Duration: %v, Error: %v, Response: %+v", duration, err, res)
	for i, m := range mockInner.LastMessages {
		t.Logf("LastMessage[%d]: Role=%s, Content=%s", i, m.Role, m.Content)
	}

	if err != nil {
		t.Fatalf("RunToolLoop failed: %v", err)
	}

	// 验证耗时是否在 20s 左右（允许少量误差）
	if duration < 19*time.Second || duration > 22*time.Second {
		t.Errorf("Expected timeout around 20s, got %v", duration)
	}

	// 验证是否收到了超时提示
	foundTimeout := false
	for _, m := range mockInner.LastMessages {
		if m.Role == schema.Tool && strings.Contains(m.Content, "超时") {
			foundTimeout = true
			break
		}
	}
	if !foundTimeout {
		t.Error("Did not find timeout error message in tool results")
	}
}

func TestToolHardening_ZeroTrust(t *testing.T) {
	authTool := &AuthCheckTool{}
	
	mockInner := &model.MockChatModel{}
	chatModel := model.NewChatModelFromInner(mockInner)
	
	mockInner.GenerateFunc = func(ctx context.Context, messages []*schema.Message, opts ...einomodel.Option) (*schema.Message, error) {
		if len(messages) == 1 {
			return &schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{ID: "call_auth", Function: schema.FunctionCall{Name: "auth_tool", Arguments: "{\"user_id\": 999}"}},
				},
			}, nil
		}
		return &schema.Message{Role: schema.Assistant, Content: "完成"}, nil
	}

	// 注入真实的 UserID 为 888
	realUserID := int64(888)
	ctx := context.WithValue(context.Background(), "user_id", realUserID)
	
	msgs := []*schema.Message{schema.UserMessage("Check auth")}
	_, err := RunToolLoop(ctx, chatModel, msgs, []tool.BaseTool{authTool})
	if err != nil {
		t.Fatalf("RunToolLoop failed: %v", err)
	}

	// 验证工具实际接收到的是 888，而不是 LLM 伪造的 999
	if authTool.CapturedUserID != realUserID {
		t.Errorf("Zero Trust Failed: Expected UserID %d, got %d", realUserID, authTool.CapturedUserID)
	}
}
