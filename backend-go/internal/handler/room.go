package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// RoomHandler 房间处理器
type RoomHandler struct {
	*BaseHandler
	roomService service.RoomService
}

// NewRoomHandler 创建新的房间处理器
func NewRoomHandler(roomService service.RoomService, logger *logger.Logger) *RoomHandler {
	return &RoomHandler{
		BaseHandler: NewBaseHandler(logger),
		roomService: roomService,
	}
}

// CreateRoom 创建房间
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	var req model.RoomCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	room, err := h.roomService.CreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, room)
}

// GetRoomList 获取房间列表
func (h *RoomHandler) GetRoomList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	rooms, _, err := h.roomService.GetRoomList(c.Request.Context(), page, pageSize)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, rooms)
}

// GetRoomByID 获取房间详情
func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	room, err := h.roomService.GetRoomByID(c.Request.Context(), roomID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, room)
}

// UpdateRoom 更新房间
func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	var req model.RoomUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	room, err := h.roomService.UpdateRoom(c.Request.Context(), userID, roomID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, room)
}

// DeleteRoom 删除房间
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	if err := h.roomService.DeleteRoom(c.Request.Context(), userID, roomID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "房间已解散"})
}

// JoinRoom 加入房间
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	if err := h.roomService.JoinRoom(c.Request.Context(), userID, roomID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "加入房间成功"})
}

// LeaveRoom 离开房间
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	if err := h.roomService.LeaveRoom(c.Request.Context(), userID, roomID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "已离开房间"})
}

// GetRoomMembers 获取房间成员
func (h *RoomHandler) GetRoomMembers(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	members, err := h.roomService.GetRoomMembers(c.Request.Context(), roomID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, members)
}

// TransferOwner 转移房主
func (h *RoomHandler) TransferOwner(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	var req map[string]int64
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	newOwnerID, ok := req["newOwnerId"]
	if !ok {
		h.BadRequest(c, "新房主ID不能为空")
		return
	}

	if err := h.roomService.TransferOwner(c.Request.Context(), userID, roomID, newOwnerID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "房主转移成功"})
}

// KickMember 踢出成员
func (h *RoomHandler) KickMember(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	var req map[string]int64
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	targetUserID, ok := req["userId"]
	if !ok {
		h.BadRequest(c, "成员ID不能为空")
		return
	}

	if err := h.roomService.KickMember(c.Request.Context(), userID, roomID, targetUserID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "成员已踢出"})
}

// UpdateMemberStatus 更新成员状态
func (h *RoomHandler) UpdateMemberStatus(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil {
		h.BadRequest(c, "房间ID无效")
		return
	}

	var req model.RoomMemberStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.roomService.UpdateMemberStatus(c.Request.Context(), userID, roomID, &req); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, map[string]string{"message": "状态更新成功"})
}

// GetPersonalRoom 获取个人自习室
func (h *RoomHandler) GetPersonalRoom(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.BadRequest(c, "用户ID无效")
		return
	}

	room, err := h.roomService.GetOrCreatePersonalRoom(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, room)
}
