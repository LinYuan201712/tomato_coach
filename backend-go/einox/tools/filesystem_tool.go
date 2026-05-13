package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// FileSystemParams 文件系统操作参数
type FileSystemParams struct {
	Action  string `json:"action" desc:"操作类型: read (读取), write (写入)"`
	Path    string `json:"path" desc:"文件相对路径"`
	Content string `json:"content,omitempty" desc:"文件内容 (仅write操作需要)"`
}

// NewFileSystemTool 创建通用文件系统工具
func NewFileSystemTool(baseDir string) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "manage_file",
			Desc: "进行通用的文件读写操作。用于处理笔记、数据文件等。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"action": {
					Type: "string",
					Desc: "操作: read, write",
				},
				"path": {
					Type: "string",
					Desc: "文件路径，相对于工作目录",
				},
				"content": {
					Type: "string",
					Desc: "写入的内容",
				},
			}),
		},
		func(ctx context.Context, params *FileSystemParams) (string, error) {
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

			// 安全检查：防止路径穿越
			safePath := filepath.Join(userBaseDir, filepath.Base(params.Path))

			switch params.Action {
			case "read":
				data, err := os.ReadFile(safePath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Sprintf("文件不存在: %s", params.Path), nil
					}
					return "", fmt.Errorf("读取失败: %v", err)
				}
				return string(data), nil

			case "write":
				err := os.WriteFile(safePath, []byte(params.Content), 0644)
				if err != nil {
					return "", fmt.Errorf("写入失败: %v", err)
				}
				return fmt.Sprintf("成功写入文件: %s", params.Path), nil

			default:
				return "未知操作", nil
			}
		},
	)
}
