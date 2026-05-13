package rag

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore 模拟向量数据库
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Add(ctx context.Context, docs []*Document) error {
	args := m.Called(ctx, docs)
	return args.Error(0)
}

func (m *MockStore) Search(ctx context.Context, vector []float64, topK int, threshold float64) ([]*Document, error) {
	args := m.Called(ctx, vector, topK, threshold)
	return args.Get(0).([]*Document), args.Error(1)
}

func (m *MockStore) List(ctx context.Context) ([]*Document, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Document), args.Error(1)
}

func (m *MockStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) HybridSearch(ctx context.Context, query string, vector []float64, topK int, threshold float64, weight float64) ([]*Document, error) {
	args := m.Called(ctx, query, vector, topK, threshold, weight)
	return args.Get(0).([]*Document), args.Error(1)
}

func (m *MockStore) GetFullDocument(ctx context.Context, fileName string, userID int64) (string, error) {
	args := m.Called(ctx, fileName, userID)
	return args.String(0), args.Error(1)
}

// MockModel 模拟 LLM
type MockModel struct {
	mock.Mock
}

func (m *MockModel) Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
	args := m.Called(ctx, msgs)
	return args.Get(0).(*schema.Message), args.Error(1)
}

// MockEmbedder 模拟嵌入模型
type MockEmbedder struct {
	mock.Mock
}

func (m *MockEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).([][]float64), args.Error(1)
}

func TestRecursiveSplitter(t *testing.T) {
	splitter := NewRecursiveSplitter(10, 2)
	doc := &Document{
		ID:      "test_doc",
		Content: "abcdefghijklmn", // 14 chars
		MetaData: map[string]any{
			"file_name": "test.txt",
		},
	}

	docs, err := splitter.Split(context.Background(), []*Document{doc})
	assert.NoError(t, err)
	assert.True(t, len(docs) > 1)
	
	// 验证元数据是否继承
	for _, d := range docs {
		assert.Equal(t, "test.txt", d.MetaData["file_name"])
	}
}

func TestProfessionalQuery_HighFrequency(t *testing.T) {
	mockStore := new(MockStore)
	mockModel := new(MockModel)
	mockEmbedder := new(MockEmbedder)
	
	ragSvc := NewSimpleRAG(mockStore, mockEmbedder, mockModel, 3, 0.7)
	
	ctx := context.Background()
	query := "hello"
	
	// 1. 模拟核心要点提取 (以及可能的后续重写)
	mockModel.On("Generate", mock.Anything, mock.Anything).Return(&schema.Message{Content: "hello"}, nil)
	
	// 2. 模拟嵌入
	mockEmbedder.On("EmbedStrings", mock.Anything, mock.Anything).Return([][]float64{{0.1, 0.2}}, nil)
	
	// 3. 模拟检索结果 (命中 3 个同源文档)
	docs := []*Document{
		{ID: "1", Content: "c1", MetaData: map[string]any{MetaFileName: "doc1.txt"}},
		{ID: "2", Content: "c2", MetaData: map[string]any{MetaFileName: "doc1.txt"}},
		{ID: "3", Content: "c3", MetaData: map[string]any{MetaFileName: "doc1.txt"}},
	}
	mockStore.On("HybridSearch", mock.Anything, "hello", []float64{0.1, 0.2}, 3, 0.3, 0.7).Return(docs, nil)
	
	// 4. 模拟全文提取
	mockStore.On("GetFullDocument", mock.Anything, "doc1.txt", int64(123)).Return("full content of doc1", nil)
	
	contextStr, _, err := ragSvc.ProfessionalQuery(ctx, query, "now", "kb", 123)
	
	assert.NoError(t, err)
	// 验证不再自动替换为全文
	assert.NotContains(t, contextStr, "full content of doc1")
	assert.NotContains(t, contextStr, "--- 文档《doc1.txt》全文开始 ---")
	// 验证保留了切片内容
	assert.Contains(t, contextStr, "c1")
	assert.Contains(t, contextStr, "c2")
	assert.Contains(t, contextStr, "c3")
}
