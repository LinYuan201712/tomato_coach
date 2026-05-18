package rag

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"os"
)

const (
	MetaFileName   = "file_name"
	MetaUpdateTime = "update_time"
	MetaTitleChain = "title_chain"
	MetaUserID     = "user_id"
	MetaStartIndex = "start_idx"
	MetaEndIndex   = "end_idx"
)

type contextKey string

const AllowedKnowledgeFilesContextKey contextKey = "allowed_knowledge_files"

// Document 文档结构
type Document struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Vector   []float64      `json:"vector,omitempty"`
	MetaData map[string]any `json:"metadata,omitempty"`
	Score    float64        `json:"score,omitempty"`
}

// VectorStore 向量数据库接口
type VectorStore interface {
	Add(ctx context.Context, docs []*Document) error
	Search(ctx context.Context, vector []float64, topK int, threshold float64) ([]*Document, error)
	List(ctx context.Context) ([]*Document, error)
	Delete(ctx context.Context, id string) error
}

// HybridStore 混合检索数据库接口
type HybridStore interface {
	VectorStore
	// HybridSearch 混合搜索：结合 BM25 和 向量搜索
	HybridSearch(ctx context.Context, query string, vector []float64, topK int, threshold float64, weight float64) ([]*Document, error)
	// GetFullDocument 获取源文件的全文
	GetFullDocument(ctx context.Context, fileName string, userID int64) (string, error)
}

// Embedder 文本嵌入接口
type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}

// DocumentSplitter 文档切分接口
type DocumentSplitter interface {
	Split(ctx context.Context, docs []*Document) ([]*Document, error)
}

// MemoryVectorStore 内存向量数据库实现
type MemoryVectorStore struct {
	docs []*Document
}

func NewMemoryVectorStore() *MemoryVectorStore {
	store := &MemoryVectorStore{
		docs: make([]*Document, 0),
	}
	// 确保目录存在
	_ = os.MkdirAll("data", 0755)
	// 自动加载
	_ = store.Load("data/rag_store.json")
	return store
}

func (s *MemoryVectorStore) Save(path string) error {
	data, err := json.Marshal(s.docs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *MemoryVectorStore) Load(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.docs)
}

func (s *MemoryVectorStore) Add(ctx context.Context, docs []*Document) error {
	s.docs = append(s.docs, docs...)
	// 自动保存
	go s.Save("data/rag_store.json")
	return nil
}

func (s *MemoryVectorStore) Search(ctx context.Context, vector []float64, topK int, threshold float64) ([]*Document, error) {
	results := make([]*Document, 0)
	for _, doc := range s.docs {
		score := cosineSimilarity(vector, doc.Vector)
		if score >= threshold {
			d := *doc
			d.Score = score
			results = append(results, &d)
		}
	}
	return results, nil
}

func (s *MemoryVectorStore) List(ctx context.Context) ([]*Document, error) {
	return s.docs, nil
}

func (s *MemoryVectorStore) Delete(ctx context.Context, id string) error {
	for i, doc := range s.docs {
		if doc.ID == id {
			s.docs = append(s.docs[:i], s.docs[i+1:]...)
			return nil
		}
	}
	return nil
}

// RecursiveSplitter 递归文档切分实现
type RecursiveSplitter struct {
	ChunkSize    int
	ChunkOverlap int
}

func NewRecursiveSplitter(size, overlap int) *RecursiveSplitter {
	return &RecursiveSplitter{size, overlap}
}

func (s *RecursiveSplitter) Split(ctx context.Context, docs []*Document) ([]*Document, error) {
	results := make([]*Document, 0)
	for _, doc := range docs {
		content := doc.Content
		if len(content) <= s.ChunkSize {
			results = append(results, doc)
			continue
		}

		// 简单的按长度切分，实际中可以按段落或 Tokenizer 切分
		// 这里假设 1 token ≈ 1.5 - 2 字符，200-300 token 约 400-600 字符
		// 如果用户明确要求 200-300 token，我们在这里调整默认值

		runes := []rune(content)
		for i := 0; i < len(runes); i += s.ChunkSize - s.ChunkOverlap {
			end := i + s.ChunkSize
			if end > len(runes) {
				end = len(runes)
			}

			chunkContent := string(runes[i:end])

			// 继承元数据
			newMeta := make(map[string]any)
			for k, v := range doc.MetaData {
				newMeta[k] = v
			}

			// 可以在这里根据内容提取标题链（如果内容中有 # 等标识）
			// 暂时简单透传，由外部注入
			newMeta[MetaStartIndex] = i
			newMeta[MetaEndIndex] = end

			results = append(results, &Document{
				ID:       fmt.Sprintf("%s_%d", doc.ID, i),
				Content:  chunkContent,
				MetaData: newMeta,
			})

			if end == len(runes) {
				break
			}
		}
	}
	return results, nil
}

func GenerateID(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

func cosineSimilarity(v1, v2 []float64) float64 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0
	}
	var dot, n1, n2 float64
	for i := range v1 {
		dot += v1[i] * v2[i]
		n1 += v1[i] * v1[i]
		n2 += v2[i] * v2[i]
	}
	if n1 == 0 || n2 == 0 {
		return 0
	}
	return dot / (math.Sqrt(n1) * math.Sqrt(n2))
}
