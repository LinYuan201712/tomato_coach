package workflow

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	einomodel "github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/rag"
	domainmodel "github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/einox/tools"
)

// MockTaskProvider 模拟任务服务
type MockTaskProvider struct{}
func (m *MockTaskProvider) CreateTask(ctx context.Context, userID int64, req *domainmodel.TaskCreateRequest) (*domainmodel.TaskResponse, error) {
	return &domainmodel.TaskResponse{TaskID: 1, TaskName: req.TaskName}, nil
}
func (m *MockTaskProvider) GetTaskList(ctx context.Context, userID int64) ([]*domainmodel.TaskResponse, error) {
	return []*domainmodel.TaskResponse{{TaskID: 1, TaskName: "测试任务"}}, nil
}
func (m *MockTaskProvider) CompleteTask(ctx context.Context, userID int64, taskID int64) error {
	return nil
}

// MockUserProvider 模拟用户画像服务
type MockUserProvider struct{}
func (m *MockUserProvider) FindByUserID(ctx context.Context, userID int64) (*entity.User, error) {
	return &entity.User{
		Goals: "学习 Go",
		PreferredStyle: "硬核",
	}, nil
}
func (m *MockUserProvider) UpdateProfile(ctx context.Context, userID int64, goals string, style string) error {
	return nil
}

// MockEmbedder 模拟向量化模型
type MockEmbedder struct{}
func (m *MockEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	res := make([][]float64, len(texts))
	for i := range texts {
		res[i] = []float64{0.1, 0.2, 0.3}
	}
	return res, nil
}

func TestBuildCoachChatGraph_Routing(t *testing.T) {
	ctx := context.Background()
	
	// 1. 初始化 Mock 环境
	mockInner := &einomodel.MockChatModel{}
	chatModel := einomodel.NewChatModelFromInner(mockInner)
	
	store := rag.NewMemoryVectorStore()
	embedder := &MockEmbedder{}
	r := rag.NewSimpleRAG(store, embedder, chatModel, 5, 0.5)
	
	taskSvc := &MockTaskProvider{}
	userRepo := &MockUserProvider{}
	registry := tools.NewSemanticToolRegistry(embedder)
	
	graph, err := BuildCoachChatGraph(ctx, chatModel, chatModel, r, registry)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	
	// 2. 测试任务路由 (带有“任务”关键词，直接走 detectIntent 的关键词逻辑)
	mockInner.Response = &schema.Message{Role: schema.Assistant, Content: "正在为你创建任务"}
	input := &CoachChatInput{
		UserID:       1,
		Query:        "帮我创建一个学习任务",
		UserProfile:  "学习 Go",
	}
	
	resp, err := graph.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	
	if !strings.Contains(resp.Content, "任务") {
		t.Errorf("Expected task related response, got: %s", resp.Content)
	}
	
	// 3. 测试情感路由 (模拟 LLM 识别意图)
	mockInner.Response = nil
	mockInner.ResponseQueue = []*schema.Message{
		{Role: schema.Assistant, Content: "emotion"}, // 第一次：detectIntent
		{Role: schema.Assistant, Content: "别伤心，有我在呢"}, // 第二次：EmotionAgent.Generate
	}
	
	input = &CoachChatInput{
		UserID:       1,
		Query:        "我有点难过",
		UserProfile:  "学习 Go",
	}
	
	resp, err = graph.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	
	if !strings.Contains(resp.Content, "在呢") {
		t.Errorf("Expected emotion companion response, got: %s", resp.Content)
	}
}
