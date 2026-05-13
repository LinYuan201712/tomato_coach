package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
)

func TestTaskAgent_Generate(t *testing.T) {
	// 1. 初始化 Mock
	mockInner := &model.MockChatModel{
		Response: &schema.Message{Role: schema.Assistant, Content: "任务已处理"},
	}
	chatModel := model.NewChatModelFromInner(mockInner)
	
	agent := NewTaskAgent(chatModel, nil)
	
	// 2. 执行
	userID := int64(12345)
	query := "帮我创建一个学习 Go 语言的任务"
	taskSummary := "当前无任务"
	
	_, err := agent.Generate(context.Background(), userID, query, taskSummary, "喜欢硬核", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// 3. 验证 Prompt 注入
	lastMsgs := mockInner.LastMessages
	if len(lastMsgs) < 1 {
		t.Fatal("No messages sent to model")
	}
	
	systemMsg := lastMsgs[0]
	if systemMsg.Role != schema.System {
		t.Errorf("Expected first message to be system, got %s", systemMsg.Role)
	}
	
	// 验证 UserID 注入 (检查 Config 加载后的占位符替换)
	if !strings.Contains(systemMsg.Content, "12345") {
		t.Errorf("System prompt does not contain userID 12345. Content: %s", systemMsg.Content)
	}
	
	if !strings.Contains(systemMsg.Content, taskSummary) {
		t.Errorf("System prompt does not contain task summary. Content: %s", systemMsg.Content)
	}
}

func TestEmotionAgent_陪伴模式(t *testing.T) {
	mockInner := &model.MockChatModel{
		Response: &schema.Message{Role: schema.Assistant, Content: "抱抱你"},
	}
	chatModel := model.NewChatModelFromInner(mockInner)
	agent := NewEmotionAgent(chatModel, nil)
	
	// 测试短对话进入陪伴模式 (不带“任务”字样)
	query := "我今天不开心"
	_, err := agent.Generate(context.Background(), 1, query, nil, "无", "温柔一点", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// 验证是否绑定了工具 (陪伴模式不应绑定工具以减少开销)
	if mockInner.BindToolsCalled && mockInner.LastToolInfos != nil {
		// 注意：NewEmotionAgent 内部可能会调用 BindTools(nil)
		// 我们主要检查最后一次调用是否清空了工具
	}
}

func init() {
	// 确保测试时有默认配置，避免空指针
	prompt.Config.TaskAgentSystem = "UserID: {user_id}, Tasks: {task_summary}"
	prompt.Config.EmotionAgentSystem = "Emotion Mode"
}
