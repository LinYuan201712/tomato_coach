package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/internal/repository"
)

// ReportParams 学习报告查询参数
type ReportParams struct {
	Date string `json:"date" desc:"查询日期，格式为 YYYY-MM-DD，留空则默认为昨天"`
}

// NewReportTool 创建一个用于查询学习报告的工具
func NewReportTool(repo repository.StudyReportRepository) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "get_study_report",
			Desc: "获取用户特定日期的学习日报/总结。当用户询问“我昨天学了什么”或“帮我回顾报告”时使用。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"date": {
					Type: "string",
					Desc: "查询日期 (YYYY-MM-DD)，如 '2026-05-10'。默认为昨天。",
				},
			}),
		},
		func(ctx context.Context, params *ReportParams) (string, error) {
			if repo == nil {
				return "", fmt.Errorf("report repository not initialized")
			}

			userID, ok := ctx.Value("user_id").(int64)
			if !ok {
				return "", fmt.Errorf("鉴权失败：未在上下文中找到用户 ID")
			}

			var targetDate time.Time
			if params.Date == "" {
				targetDate = time.Now().AddDate(0, 0, -1)
			} else {
				parsed, err := time.Parse("2006-01-02", params.Date)
				if err != nil {
					return "", fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD")
				}
				targetDate = parsed
			}

			report, err := repo.FindByUserIDAndDate(ctx, userID, targetDate, "daily")
			if err != nil {
				return "未找到该日期的学习报告。请确认昨天是否有进行过学习并生成了报告。", nil
			}

			if report == nil {
				return "未找到该日期的学习报告。", nil
			}

			return fmt.Sprintf("【%s 学习日报】\n\n- 专注总时长: %d 分钟\n- 已完成任务: %d\n\n报告详情:\n%s",
				report.ReportDate.Format("2006-01-02"),
				report.TotalFocusTime,
				report.CompletedTasks,
				report.Content,
			), nil
		},
	)
}
