package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/internal/domain/entity"
)

// UserProvider 定义了工具所需的用户操作接口
type UserProvider interface {
	FindByUserID(ctx context.Context, userID int64) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID int64, goals string, style string) error
}

// UserInfoResult 用户信息结果
type UserInfoResult struct {
	Goals       []string `json:"goals"`
	Style       string   `json:"style"`
	TomatoCount int      `json:"tomato_count"`
}

// NewUserProfilingTool 创建一个获取用户画像的工具
func NewUserProfilingTool(userRepo UserProvider) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "get_user_profile",
			Desc: "获取当前用户的画像、目标和偏好的教练风格",
		},
		func(ctx context.Context, params map[string]any) (*UserInfoResult, error) {
			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return nil, fmt.Errorf("user_id not found in context")
			}

			user, err := userRepo.FindByUserID(ctx, userID)
			if err != nil {
				return nil, err
			}

			if user == nil {
				return nil, fmt.Errorf("user not found")
			}

			goals := make([]string, 0)
			if user.Goals != "" {
				// 简单处理：逗号分隔
				goals = strings.Split(user.Goals, ",")
			}

			return &UserInfoResult{
				Goals:       goals,
				Style:       user.PreferredStyle,
				TomatoCount: user.Tomato,
			}, nil
		},
	)
}

// UpdateProfileParams 更新参数
type UpdateProfileParams struct {
	Goal  string `json:"goal"`
	Style string `json:"style"`
}

// NewUpdateProfileTool 创建一个更新用户画像的工具
func NewUpdateProfileTool(userRepo UserProvider) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "update_user_profile",
			Desc: "更新用户的学习目标或偏好的教练风格",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"goal": {
					Type: "string",
					Desc: "新的学习目标",
				},
				"style": {
					Type: "string",
					Desc: "新的教练风格",
				},
			}),
		},
		func(ctx context.Context, params *UpdateProfileParams) (string, error) {
			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return "", fmt.Errorf("user_id not found in context")
			}

			err := userRepo.UpdateProfile(ctx, userID, params.Goal, params.Style)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("个人资料已更新: 目标 [%s], 风格 [%s]", params.Goal, params.Style), nil
		},
	)
}
