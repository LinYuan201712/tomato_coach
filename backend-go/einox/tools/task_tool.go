package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/model"
)

// TaskProvider 定义了工具所需的任务操作接口，用于解除循环依赖
type TaskProvider interface {
	CreateTask(ctx context.Context, userID int64, req *model.TaskCreateRequest) (*model.TaskResponse, error)
	GetTaskList(ctx context.Context, userID int64) ([]*model.TaskResponse, error)
	CompleteTask(ctx context.Context, userID int64, taskID int64) error
}

// TaskParams 任务管理参数
type TaskParams struct {
	Action   string `json:"action" desc:"操作类型: create (创建), list (列表), complete (完成)"`
	Title    string `json:"title,omitempty" desc:"任务标题"`
	TaskName string `json:"task_name,omitempty" desc:"任务标题 (别名)"`
	Duration int    `json:"duration,omitempty" desc:"持续时间 (分钟)"`
}

// NewTaskTool 创建一个用于管理学习任务的工具
func NewTaskTool(taskSvc TaskProvider) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "manage_task",
			Desc: "管理用户的学习任务，支持创建、列出和完成操作",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"action": {
					Type: "string",
					Desc: "操作: create, list, complete",
				},
				"title": {
					Type: "string",
					Desc: "任务的动作标题 (例如：'写英语作业')。严禁在此字段包含时间信息。",
				},
				"task_name": {
					Type: "string",
					Desc: "同 title，任务标题。",
				},
				"duration": {
					Type: "integer",
					Desc: "任务持续时间 (分钟)，默认为 25",
				},
			}),
		},
		func(ctx context.Context, params *TaskParams) (string, error) {
			if taskSvc == nil {
				return "", fmt.Errorf("task service not initialized")
			}

			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return "", fmt.Errorf("鉴权失败：未在上下文中找到用户 ID")
			}

			switch params.Action {
			case "create":
				// 合并可能存在的两个标题参数
				finalTitle := params.Title
				if finalTitle == "" {
					finalTitle = params.TaskName
				}

				// 防呆校验
				title := strings.TrimSpace(finalTitle)
				if title == "" || strings.Contains(title, "分钟") || strings.Contains(title, "min") {
					return "创建失败：任务标题不能包含时间信息（如'25分钟'），请提供一个具体的任务动作名称（如'写程序设计作业'）。", nil
				}

				duration := params.Duration
				if duration <= 0 {
					duration = 25
				}

				req := &model.TaskCreateRequest{
					TaskName: title,
					Duration: duration,
				}
				resp, err := taskSvc.CreateTask(ctx, userID, req)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("成功创建任务: %s", resp.TaskName), nil

			case "list":
				tasks, err := taskSvc.GetTaskList(ctx, userID)
				if err != nil {
					return "", err
				}
				var sb strings.Builder
				sb.WriteString("以下是你目前的未完成任务：\n")
				found := false
				for _, t := range tasks {
					if t.Status == constants.TaskStatusCompleted {
						continue
					}
					sb.WriteString(fmt.Sprintf("- %s\n", t.TaskName))
					found = true
				}
				if !found {
					return "你目前没有任何未完成的任务。", nil
				}
				return sb.String(), nil

			case "complete":
				// 合并可能存在的两个标题参数
				finalTitle := params.Title
				if finalTitle == "" {
					finalTitle = params.TaskName
				}

				// 注意：这里需要 ID，但 Agent 可能只提供 Title。
				tasks, _ := taskSvc.GetTaskList(ctx, userID)
				var targetID int64
				for _, t := range tasks {
					if t.TaskName == finalTitle {
						targetID = t.TaskID
						break
					}
				}
				if targetID == 0 {
					return "找不到指定标题的任务，请先查看任务列表确认 ID。", nil
				}

				err := taskSvc.CompleteTask(ctx, userID, targetID)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("任务已标记为完成: %s", finalTitle), nil

			default:
				return "未知操作", nil
			}
		},
	)
}

// PomodoroParams 番茄钟参数
type PomodoroParams struct {
	Duration  int    `json:"duration" desc:"持续时间 (分钟)"`
	TaskTitle string `json:"task_title" desc:"关联的任务标题"`
}

// NewPomodoroTool 创建一个开启番茄钟的工具
func NewPomodoroTool() tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "start_pomodoro",
			Desc: "针对特定任务开启番茄钟专注环节",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"duration": {
					Type: "integer",
					Desc: "持续时间 (分钟), 默认为 25",
				},
				"task_title": {
					Type: "string",
					Desc: "要专注的任务标题",
				},
			}),
		},
		func(ctx context.Context, params *PomodoroParams) (string, error) {
			duration := params.Duration
			if duration <= 0 {
				duration = 25
			}
			return fmt.Sprintf("番茄钟已开启！任务: %s, 持续时间: %d 分钟。祝你专注愉快！", params.TaskTitle, duration), nil
		},
	)
}
