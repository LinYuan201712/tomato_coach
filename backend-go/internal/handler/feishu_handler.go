package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

type FeishuHandler struct {
	feishuService service.FeishuService
	logger        *logger.Logger
}

func NewFeishuHandler(feishuService service.FeishuService, logger *logger.Logger) *FeishuHandler {
	return &FeishuHandler{
		feishuService: feishuService,
		logger:        logger,
	}
}

func (h *FeishuHandler) GetConfig(c *gin.Context) {
	userID := uint64(c.MustGet("user_id").(int64))

	cfg, err := h.feishuService.GetConfig(c.Request.Context(), userID)
	if err != nil {
		h.logger.Errorf("Failed to get feishu config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get config"})
		return
	}

	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *FeishuHandler) UpdateConfig(c *gin.Context) {
	userID := uint64(c.MustGet("user_id").(int64))

	var req entity.UserFeishuConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.feishuService.SaveConfig(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Errorf("Failed to save feishu config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
