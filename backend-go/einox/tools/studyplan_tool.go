package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// StudyPlanParams 学习计划管理参数
type StudyPlanParams struct {
	Action  string `json:"action" desc:"操作类型: save (保存), read (读取), delete (删除)"`
	Title   string `json:"title" desc:"计划标题 (文件名)"`
	Content string `json:"content,omitempty" desc:"计划内容 (Markdown格式，仅save操作需要)"`
}

// NewStudyPlanTool 创建学习计划管理工具
func NewStudyPlanTool(baseDir string) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "manage_study_plan",
			Desc: "管理完整的学习计划文档（Markdown格式）。支持保存、读取和删除计划。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"action": {
					Type: "string",
					Desc: "操作: save, read, delete",
				},
				"title": {
					Type: "string",
					Desc: "计划标题，例如 'Go语言学习计划'",
				},
				"content": {
					Type: "string",
					Desc: "计划的内容，仅在 action=save 时提供",
				},
			}),
		},
		func(ctx context.Context, params *StudyPlanParams) (string, error) {
			// 从 Context 中提取 UserID，确保多用户隔离
			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return "", fmt.Errorf("鉴权失败：未在上下文中找到用户 ID")
			}

			// 为每个用户创建独立的存储目录
			userBaseDir := filepath.Join(baseDir, fmt.Sprintf("user_%d", userID))
			if err := os.MkdirAll(userBaseDir, 0755); err != nil {
				return "", fmt.Errorf("创建目录失败: %v", err)
			}

			// 简单的文件名处理
			fileName := strings.ReplaceAll(params.Title, " ", "_") + ".md"
			filePath := filepath.Join(userBaseDir, fileName)

			switch params.Action {
			case "save":
				if params.Content == "" {
					return "", fmt.Errorf("保存计划需要提供内容")
				}
				err := os.WriteFile(filePath, []byte(params.Content), 0644)
				if err != nil {
					return "", fmt.Errorf("保存失败: %v", err)
				}
				return fmt.Sprintf("成功保存学习计划: %s", params.Title), nil

			case "read":
				data, err := os.ReadFile(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Sprintf("找不到计划: %s", params.Title), nil
					}
					return "", fmt.Errorf("读取失败: %v", err)
				}
				return string(data), nil

			case "delete":
				err := os.Remove(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Sprintf("计划不存在: %s", params.Title), nil
					}
					return "", fmt.Errorf("删除失败: %v", err)
				}
				return fmt.Sprintf("成功删除学习计划: %s", params.Title), nil

			default:
				return "未知操作", nil
			}
		},
	)
}
