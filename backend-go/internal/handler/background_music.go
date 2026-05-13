package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// BackgroundMusicHandler 背景音乐处理器
type BackgroundMusicHandler struct {
	*BaseHandler
	musicService service.BackgroundMusicService
}

// NewBackgroundMusicHandler 创建新的背景音乐处理器
func NewBackgroundMusicHandler(musicService service.BackgroundMusicService, logger *logger.Logger) *BackgroundMusicHandler {
	return &BackgroundMusicHandler{
		BaseHandler:  NewBaseHandler(logger),
		musicService: musicService,
	}
}

// GetMusicList 获取音乐列表
func (h *BackgroundMusicHandler) GetMusicList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	musics, total, err := h.musicService.GetMusicList(c.Request.Context(), page, pageSize)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]interface{}{
		"musics": musics,
		"total":  total,
	})
}

// SearchMusic 搜索音乐
func (h *BackgroundMusicHandler) SearchMusic(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		h.BadRequest(c, "搜索关键词不能为空")
		return
	}

	musics, err := h.musicService.SearchMusic(c.Request.Context(), keyword)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]interface{}{
		"musics": musics,
		"count":  len(musics),
	})
}

// GetMusicByID 获取音乐详情
func (h *BackgroundMusicHandler) GetMusicByID(c *gin.Context) {
	musicID, err := strconv.ParseInt(c.Param("musicId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "音乐ID无效")
		return
	}

	music, err := h.musicService.GetMusicByID(c.Request.Context(), musicID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, music)
}

// SystemConfigHandler 系统配置处理器
type SystemConfigHandler struct {
	*BaseHandler
	configService service.SystemConfigService
}

// NewSystemConfigHandler 创建新的系统配置处理器
func NewSystemConfigHandler(configService service.SystemConfigService, logger *logger.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{
		BaseHandler:   NewBaseHandler(logger),
		configService: configService,
	}
}

// GetConfigs 获取所有配置
func (h *SystemConfigHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configService.GetConfigs(c.Request.Context())
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, configs)
}

// GetConfig 获取指定配置
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	configKey := c.Param("key")
	if configKey == "" {
		h.BadRequest(c, "配置键不能为空")
		return
	}

	value, err := h.configService.GetConfig(c.Request.Context(), configKey)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]interface{}{
		"key":   configKey,
		"value": value,
	})
}
