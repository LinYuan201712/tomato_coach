package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// FocusHandler 专注处理器
type FocusHandler struct {
	*BaseHandler
	focusService service.FocusService
}

// NewFocusHandler 创建新的专注处理器
func NewFocusHandler(focusService service.FocusService, logger *logger.Logger) *FocusHandler {
	return &FocusHandler{
		BaseHandler:  NewBaseHandler(logger),
		focusService: focusService,
	}
}

// StartFocus 开始专注
func (h *FocusHandler) StartFocus(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req model.StartFocusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	focus, err := h.focusService.StartFocus(c.Request.Context(), userID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, focus)
}

// StopFocus 结束专注
func (h *FocusHandler) StopFocus(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req map[string]interface{}
	var sessionID int64
	
	// 尝试绑定 JSON，但不强制要求有 body
	if err := c.ShouldBindJSON(&req); err == nil {
		if id, ok := req["sessionId"]; ok {
			switch v := id.(type) {
			case float64:
				sessionID = int64(v)
			case string:
				sessionID, _ = strconv.ParseInt(v, 10, 64)
			}
		}
	}

	result, err := h.focusService.StopFocus(c.Request.Context(), userID, sessionID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, result)
}

// GetFocusRecords 获取专注记录
func (h *FocusHandler) GetFocusRecords(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	records, err := h.focusService.GetFocusRecords(c.Request.Context(), userID, days)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]interface{}{
		"records": records,
		"count":   len(records),
	})
}

// GetDailyReport 获取每日报告
func (h *FocusHandler) GetDailyReport(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	report, err := h.focusService.GetDailyReport(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, report)
}

// GetWeeklyReport 获取周报告
func (h *FocusHandler) GetWeeklyReport(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	report, err := h.focusService.GetWeeklyReport(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, report)
}

// GetMonthlyReport 获取月报告
func (h *FocusHandler) GetMonthlyReport(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	report, err := h.focusService.GetMonthlyReport(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, report)
}
