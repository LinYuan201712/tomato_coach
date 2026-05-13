package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// FriendHandler 好友处理器
type FriendHandler struct {
	*BaseHandler
	friendService service.FriendService
}

// NewFriendHandler 创建新的好友处理器
func NewFriendHandler(friendService service.FriendService, logger *logger.Logger) *FriendHandler {
	return &FriendHandler{
		BaseHandler:   NewBaseHandler(logger),
		friendService: friendService,
	}
}

// SendFriendRequest 发送好友申请
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req model.FriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.friendService.SendFriendRequest(c.Request.Context(), userID, &req); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "好友申请已发送"})
}

// GetFriendRequests 获取好友申请列表
func (h *FriendHandler) GetFriendRequests(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	status := c.DefaultQuery("status", "pending")

	requests, err := h.friendService.GetFriendRequests(c.Request.Context(), userID, status)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, requests)
}

// ProcessFriendRequest 处理好友申请
func (h *FriendHandler) ProcessFriendRequest(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req model.ProcessFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.friendService.ProcessFriendRequest(c.Request.Context(), userID, &req); err != nil {
		h.Error(c, err)
		return
	}

	message := "已拒绝"
	if req.Action == "accept" {
		message = "已同意"
	}
	h.Success(c, map[string]string{"message": message})
}

// GetFriendList 获取好友列表
func (h *FriendHandler) GetFriendList(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	friends, err := h.friendService.GetFriendList(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, friends)
}

// RemoveFriend 删除好友
func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req model.DeleteFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.friendService.RemoveFriend(c.Request.Context(), userID, &req); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "好友已删除"})
}
