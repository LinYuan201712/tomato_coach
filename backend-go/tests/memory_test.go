package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/memory"
	"github.com/tomato/backend/internal/domain/entity"
)

// MockRepo 模拟 Repository
type MockRepo struct {
	session *entity.ChatSession
	history []*entity.ChatMessage
}

func (m *MockRepo) SaveMessage(ctx context.Context, msg *entity.ChatMessage) error {
	m.history = append(m.history, msg)
	return nil
}

func (m *MockRepo) GetHistory(ctx context.Context, sessionID string, limit int) ([]*entity.ChatMessage, error) {
	if len(m.history) > limit {
		return m.history[len(m.history)-limit:], nil
	}
	return m.history, nil
}

func (m *MockRepo) CreateSession(ctx context.Context, session *entity.ChatSession) error {
	m.session = session
	return nil
}

func (m *MockRepo) GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error) {
	return nil, nil
}

func (m *MockRepo) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return nil
}

func (m *MockRepo) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockRepo) GetSession(ctx context.Context, sessionID string) (*entity.ChatSession, error) {
	return m.session, nil
}

func (m *MockRepo) UpdateSessionSummary(ctx context.Context, sessionID string, summary string, msgCount int) error {
	m.session.Summary = summary
	return nil
}

// MockSummarizer 模拟 LLM 摘要生成器
type MockSummarizer struct{}

func (m *MockSummarizer) Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
	// 简单的模拟逻辑：返回一条包含“Summary of”的消息
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "Summarized content",
	}, nil
}

func TestPersistentMemory_SlidingWindow(t *testing.T) {
	repo := &MockRepo{
		session: &entity.ChatSession{SessionID: "test_session", Summary: ""},
		history: make([]*entity.ChatMessage, 0),
	}
	summarizer := &MockSummarizer{}
	m := memory.NewPersistentMemory(repo, summarizer)

	ctx := context.Background()

	// 1. 模拟存入 24 条消息（未触发压缩）
	for i := 0; i < 24; i++ {
		repo.history = append(repo.history, &entity.ChatMessage{
			SessionID: "test_session",
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
		})
	}

	history, err := m.GetHistory(ctx, 1, "test_session")
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 24 {
		t.Errorf("Expected 24 messages, got %d", len(history))
	}

	// 2. 存入第 25 条消息，触发压缩
	repo.history = append(repo.history, &entity.ChatMessage{
		SessionID: "test_session",
		Role:      "user",
		Content:   "Message 25",
	})

	history, err = m.GetHistory(ctx, 1, "test_session")
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	// 压缩逻辑：前 20 条被压缩为 1 条 System 消息，剩下 5 条（21, 22, 23, 24, 25）
	// 总长度应为 1 + 5 = 6
	if len(history) != 6 {
		t.Errorf("Expected 6 messages after compression, got %d", len(history))
	}

	if history[0].Role != schema.System {
		t.Errorf("First message should be System summary, got %v", history[0].Role)
	}

	// 3. 验证摘要是否持久化到 repo
	if repo.session.Summary != "Summarized content" {
		t.Errorf("Summary was not persisted to repository")
	}

	// 4. 再次调用，验证摘要是否被正确读取
	history, err = m.GetHistory(ctx, 1, "test_session")
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 6 {
		t.Errorf("Expected 6 messages in subsequent call, got %d", len(history))
	}
	if history[0].Content != "以下是之前的历史对话摘要：Summarized content" {
		t.Errorf("Summary not correctly prepended, got: %s", history[0].Content)
	}
}
