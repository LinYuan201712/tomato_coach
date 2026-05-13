package prompt

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestTruncateMessages(t *testing.T) {
	msgs := []*schema.Message{
		{Role: schema.System, Content: "You are an assistant."},
		{Role: schema.User, Content: "Hello, this is a very long message that we want to test truncation with."},
		{Role: schema.Assistant, Content: "I understand."},
		{Role: schema.User, Content: "Tell me more."},
	}

	// 1. 测试不截断（预算足够）
	res := TruncateMessages(msgs, 1000)
	if len(res) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(res))
	}

	// 2. 测试截断（预算极低，只保留 System 和最后一条 User）
	// 注意：Tiktoken 编码后 Token 数会比字数少，但我们设置一个极小值
	res = TruncateMessages(msgs, 10) 
	
	// TruncateMessages 的逻辑是：保留 System，然后从后往前加
	if len(res) < 2 {
		t.Fatalf("Expected at least 2 messages (System + Last User), got %d", len(res))
	}
	
	if res[0].Role != schema.System {
		t.Errorf("First message should be System, got %s", res[0].Role)
	}
	
	lastMsg := res[len(res)-1]
	if lastMsg.Content != "Tell me more." {
		t.Errorf("Last message should be 'Tell me more.', got '%s'", lastMsg.Content)
	}
}

func TestCountTokens(t *testing.T) {
	msgs := []*schema.Message{
		{Role: schema.User, Content: "Hello"},
	}
	
	count := CountTokens(msgs)
	if count <= 0 {
		t.Errorf("Token count should be positive, got %d", count)
	}
	
	// "Hello" 应该是 1 个 token，加上消息格式开销
	if count < 5 {
		t.Errorf("Token count seems too low: %d", count)
	}
}
