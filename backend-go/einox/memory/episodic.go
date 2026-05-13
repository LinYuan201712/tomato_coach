package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tomato/backend/einox/rag"
)

// EpisodicMemory 情景记忆接口
type EpisodicMemory interface {
	// Store 异步提取对话中的重要情景并向量化存储
	Store(ctx context.Context, userID int64, sessionID string, query, reply string) error
	// Recall 在新对话开始前，召回相关的历史情景片段
	Recall(ctx context.Context, userID int64, currentQuery string) (string, error)
}

// VectorEpisodicMemory 基于向量数据库的情景记忆实现
type VectorEpisodicMemory struct {
	store    rag.VectorStore
	embedder rag.Embedder
}

func NewVectorEpisodicMemory(store rag.VectorStore, embedder rag.Embedder) *VectorEpisodicMemory {
	return &VectorEpisodicMemory{
		store:    store,
		embedder: embedder,
	}
}

func (m *VectorEpisodicMemory) Store(ctx context.Context, userID int64, sessionID string, query, reply string) error {
	// 简单过滤过短的对话，避免噪音
	if len(query) < 10 && len(reply) < 20 {
		return nil
	}

	content := fmt.Sprintf("用户问: %s\nAI答: %s", query, reply)
	doc := &rag.Document{
		ID:      rag.GenerateID(fmt.Sprintf("epi_%d_%d", userID, time.Now().UnixNano())),
		Content: content,
		MetaData: map[string]any{
			"user_id":    userID,
			"session_id": sessionID,
			"type":       "episodic_memory",
			"created_at": time.Now().Format(time.RFC3339),
		},
	}

	vectors, err := m.embedder.EmbedStrings(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("记忆向量化失败: %w", err)
	}
	doc.Vector = vectors[0]

	return m.store.Add(ctx, []*rag.Document{doc})
}

func (m *VectorEpisodicMemory) Recall(ctx context.Context, userID int64, currentQuery string) (string, error) {
	vectors, err := m.embedder.EmbedStrings(ctx, []string{currentQuery})
	if err != nil {
		return "", err
	}

	// 搜索相关历史（Top 10，为时间加权留出空间）
	docs, err := m.store.Search(ctx, vectors[0], 10, 0.4)
	if err != nil {
		return "", err
	}

	type scoredDoc struct {
		doc        *rag.Document
		finalScore float64
	}
	var scoredDocs []scoredDoc

	decayRate := 0.005 // 小时衰减率
	now := time.Now()

	for _, doc := range docs {
		// 验证用户隔离
		uid, ok := doc.MetaData["user_id"].(int64)
		if !ok {
			if fuid, ok := doc.MetaData["user_id"].(float64); ok {
				uid = int64(fuid)
			}
		}
		if uid != userID {
			continue
		}

		// 计算时间衰减
		finalScore := doc.Score
		createdAtStr, ok := doc.MetaData["created_at"].(string)
		if ok {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err == nil {
				hoursPassed := now.Sub(createdAt).Hours()
				// Formula: Score * e^(-decay * hours)
				finalScore = doc.Score * math.Exp(-decayRate*hoursPassed)
			}
		}

		scoredDocs = append(scoredDocs, scoredDoc{doc: doc, finalScore: finalScore})
	}

	// 按最终评分排序
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].finalScore > scoredDocs[j].finalScore
	})

	// 取前 3 名
	var relevantMemories []string
	for i := 0; i < len(scoredDocs) && i < 3; i++ {
		sd := scoredDocs[i]
		createdAt, _ := sd.doc.MetaData["created_at"].(string)
		relevantMemories = append(relevantMemories, fmt.Sprintf("[%s] (相关度:%.2f) %s", createdAt, sd.finalScore, sd.doc.Content))
	}

	if len(relevantMemories) == 0 {
		return "", nil
	}

	return "【召回的相关历史记忆】\n" + strings.Join(relevantMemories, "\n---\n"), nil
}
