package tools

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/cloudwego/eino/components/tool"
	"github.com/tomato/backend/einox/rag"
)

// ToolRegistry 动态工具注册表接口
type ToolRegistry interface {
	// Register 注册工具并对其描述进行向量化
	Register(ctx context.Context, t tool.BaseTool) error
	// Retrieve 根据查询语义召回 Top-K 最相关的工具
	Retrieve(ctx context.Context, query string, topK int) ([]tool.BaseTool, error)
	// GetAll 获取所有已注册的工具
	GetAll() []tool.BaseTool
}

type toolEntry struct {
	tool   tool.BaseTool
	vector []float64
	desc   string
}

// SemanticToolRegistry 基于语义检索的工具注册表实现
type SemanticToolRegistry struct {
	embedder rag.Embedder
	tools    []toolEntry
}

func NewSemanticToolRegistry(embedder rag.Embedder) *SemanticToolRegistry {
	return &SemanticToolRegistry{
		embedder: embedder,
		tools:    make([]toolEntry, 0),
	}
}

func (r *SemanticToolRegistry) Register(ctx context.Context, t tool.BaseTool) error {
	info, err := t.Info(ctx)
	if err != nil {
		return fmt.Errorf("获取工具信息失败: %w", err)
	}

	desc := fmt.Sprintf("%s: %s", info.Name, info.Desc)
	vectors, err := r.embedder.EmbedStrings(ctx, []string{desc})
	if err != nil {
		return fmt.Errorf("工具描述向量化失败: %w", err)
	}

	if len(vectors) == 0 {
		return fmt.Errorf("向量化返回为空")
	}

	r.tools = append(r.tools, toolEntry{
		tool:   t,
		vector: vectors[0],
		desc:   desc,
	})

	return nil
}

func (r *SemanticToolRegistry) Retrieve(ctx context.Context, query string, topK int) ([]tool.BaseTool, error) {
	if len(r.tools) == 0 {
		return nil, nil
	}

	if topK > len(r.tools) {
		topK = len(r.tools)
	}

	// 对查询进行向量化
	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	queryVec := vectors[0]

	// 计算相似度并排序
	type scoredTool struct {
		t     tool.BaseTool
		score float64
	}
	scored := make([]scoredTool, len(r.tools))
	for i, entry := range r.tools {
		scored[i] = scoredTool{
			t:     entry.tool,
			score: cosineSimilarity(queryVec, entry.vector),
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 返回 Top-K
	results := make([]tool.BaseTool, topK)
	for i := 0; i < topK; i++ {
		results[i] = scored[i].t
		fmt.Printf("[TOOL RAG] Selected tool: %s (Score: %.4f)\n", scored[i].t, scored[i].score)
	}

	return results, nil
}

func (r *SemanticToolRegistry) GetAll() []tool.BaseTool {
	res := make([]tool.BaseTool, len(r.tools))
	for i, entry := range r.tools {
		res[i] = entry.tool
	}
	return res
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
