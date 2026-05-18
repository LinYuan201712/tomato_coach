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

func TestBuildContext_UsesUploadedFileCitationFormat(t *testing.T) {
	docs := []*Document{
		{
			ID:      "1",
			Content: "区块链是一种分布式账本技术。",
			MetaData: map[string]any{
				MetaFileName: "区块链发展史：从比特币到模块化.md",
			},
		},
	}

	contextStr := BuildContext(docs)

	assert.Contains(t, contextStr, "来源：区块链发展史：从比特币到模块化.md")
	assert.Contains(t, contextStr, "[1] 来源：[[区块链发展史：从比特币到模块化.md]]")
	assert.Contains(t, contextStr, "禁止生成 Markdown 链接")
}

func TestGetFullDocument_FallsBackToListWhenHybridExactQueryIsEmpty(t *testing.T) {
	mockStore := new(MockStore)
	ragSvc := NewSimpleRAG(mockStore, nil, nil, 3, 0.7)
	ctx := context.Background()

	docs := []*Document{
		{
			ID:      "chunk-2",
			Content: "第二段",
			MetaData: map[string]any{
				MetaFileName:   "区块链发展史与关键技术演进.pdf",
				MetaUserID:     float64(2047527547603783680),
				MetaStartIndex: 10,
			},
		},
		{
			ID:      "chunk-1",
			Content: "第一段",
			MetaData: map[string]any{
				MetaFileName:   "区块链发展史与关键技术演进.pdf",
				MetaUserID:     float64(2047527547603783680),
				MetaStartIndex: 0,
			},
		},
	}

	mockStore.On("GetFullDocument", mock.Anything, "区块链发展史与关键技术演进.pdf", int64(2047527547603783680)).Return("", nil)
	mockStore.On("List", mock.Anything).Return(docs, nil)

	content, err := ragSvc.GetFullDocument(ctx, "区块链发展史与关键技术演进.pdf", 2047527547603783680)

	assert.NoError(t, err)
	assert.Contains(t, content, "第一段\n第二段")
}

func TestDeleteFile_RemovesAllChunksForUserFile(t *testing.T) {
	mockStore := new(MockStore)
	ragSvc := NewSimpleRAG(mockStore, nil, nil, 3, 0.7)
	ctx := context.Background()

	docs := []*Document{
		{ID: "chunk-1", MetaData: map[string]any{MetaFileName: "区块链的发展.md", MetaUserID: int64(1)}},
		{ID: "chunk-2", MetaData: map[string]any{MetaFileName: "区块链的发展.md", MetaUserID: float64(1)}},
		{ID: "other-user", MetaData: map[string]any{MetaFileName: "区块链的发展.md", MetaUserID: int64(2)}},
		{ID: "other-file", MetaData: map[string]any{MetaFileName: "区块链发展史.pdf", MetaUserID: int64(1)}},
	}

	mockStore.On("List", mock.Anything).Return(docs, nil)
	mockStore.On("Delete", mock.Anything, "chunk-1").Return(nil).Once()
	mockStore.On("Delete", mock.Anything, "chunk-2").Return(nil).Once()

	err := ragSvc.DeleteFile(ctx, "区块链的发展.md", 1)

	assert.NoError(t, err)
	mockStore.AssertNotCalled(t, "Delete", mock.Anything, "other-user")
	mockStore.AssertNotCalled(t, "Delete", mock.Anything, "other-file")
	mockStore.AssertExpectations(t)
}

func TestFilterAllowedKnowledgeDocs_DropsSourcesNotInDatabase(t *testing.T) {
	ctx := context.WithValue(context.Background(), AllowedKnowledgeFilesContextKey, map[string]bool{
		"区块链的发展.md": true,
	})
	docs := []*Document{
		{ID: "active", MetaData: map[string]any{MetaFileName: "区块链的发展.md"}},
		{ID: "stale", MetaData: map[string]any{MetaFileName: "ZK技术原理与工程实践指南.md"}},
	}

	filtered := filterAllowedKnowledgeDocs(ctx, docs)

	assert.Len(t, filtered, 1)
	assert.Equal(t, "active", filtered[0].ID)
}

func TestNormalizeSearchQuery_FallsBackWhenModelReturnsAnswer(t *testing.T) {
	original := "讲讲区块链的发展"
	candidate := "### 区块链的发展历程\n| 阶段 | 说明 |\n| --- | --- |\n这是一段完整回答，而不是检索词。"

	query := normalizeSearchQuery(original, candidate)

	assert.Equal(t, original, query)
}

func TestNormalizeSearchQuery_KeepsShortKeywords(t *testing.T) {
	query := normalizeSearchQuery("讲讲区块链的发展", "区块链 发展历程")

	assert.Equal(t, "区块链 发展历程", query)
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
