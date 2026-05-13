package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/rag"
)

// KnowledgeSearchParams 知识库检索参数
type KnowledgeSearchParams struct {
	Query string `json:"query" desc:"要在知识库中搜索的关键词或问题"`
}

// NewKnowledgeSearchTool 创建一个用于检索个人知识库的工具
func NewKnowledgeSearchTool(r *rag.SimpleRAG) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "search_knowledge",
			Desc: "检索用户的个人知识库（文档、笔记等），当你无法回答用户关于特定知识点的问题时使用",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"query": {
					Type: "string",
					Desc: "搜索关键词",
				},
			}),
		},
		func(ctx context.Context, params *KnowledgeSearchParams) (string, error) {
			if r == nil {
				return "", fmt.Errorf("RAG 引擎未初始化")
			}

			// 从 Context 中提取 UserID
			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return "", fmt.Errorf("鉴权失败：未在上下文中找到用户 ID")
			}

			fmt.Printf("[TOOL] Knowledge Search: %s (User: %d)\n", params.Query, userID)

			contextStr, _, err := r.Query(ctx, params.Query)
			if err != nil {
				return "", fmt.Errorf("检索失败: %w", err)
			}

			if contextStr == "" || contextStr == "(未找到相关知识)" {
				return "在知识库中未找到相关内容。", nil
			}

			return contextStr, nil
		},
	)
}
