package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// StudyReportHandler 学习报告处理器
type StudyReportHandler struct {
	*BaseHandler
	reportService service.StudyReportService
}

// NewStudyReportHandler 创建新的学习报告处理器
func NewStudyReportHandler(reportService service.StudyReportService, logger *logger.Logger) *StudyReportHandler {
	return &StudyReportHandler{
		BaseHandler:   NewBaseHandler(logger),
		reportService: reportService,
	}
}

// GetDailyReport 获取昨日报告
func (h *StudyReportHandler) GetDailyReport(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	report, err := h.reportService.GetLatestReport(c.Request.Context(), userID, constants.ReportTypeDaily)
	if err != nil {
		// 如果找不到昨天的报告，尝试实时生成一个
		yesterday := time.Now().AddDate(0, 0, -1)
		report, err = h.reportService.GenerateDailyReport(c.Request.Context(), userID, yesterday)
		if err != nil {
			h.Error(c, err)
			return
		}
	}

	h.Success(c, report)
}

// RegenerateDailyReport 重新生成昨日报告
func (h *StudyReportHandler) RegenerateDailyReport(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	report, err := h.reportService.RegenerateDailyReport(c.Request.Context(), userID, yesterday)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, report)
}
