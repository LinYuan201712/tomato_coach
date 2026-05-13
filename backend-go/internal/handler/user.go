package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	*BaseHandler
	userService service.UserService
}

// NewUserHandler 创建新的用户处理器
func NewUserHandler(userService service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(logger),
		userService: userService,
	}
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取登录用户的详细信息
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} Response{data=model.UserInfoResponse}
// @Failure 401 {object} Response
// @Router /user/me [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	resp, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// GetPublicUserInfo 获取公开用户信息
// @Summary 获取公开用户信息
// @Description 根据用户名获取用户信息，遵守隐私设置
// @Tags 用户
// @Security Bearer
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {object} Response{data=model.PublicUserResponse}
// @Failure 401 {object} Response
// @Router /users/{username} [get]
func (h *UserHandler) GetPublicUserInfo(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		h.BadRequest(c, "用户名不能为空")
		return
	}

	viewerID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	resp, err := h.userService.GetPublicUserInfo(c.Request.Context(), username, viewerID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// UpdateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body model.UpdateUserRequest true "更新请求"
// @Success 200 {object} Response{data=model.UserInfoResponse}
// @Failure 401 {object} Response
// @Router /user/me [put]
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	resp, err := h.userService.UpdateUserInfo(c.Request.Context(), userID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// GetUserPrivacy 获取用户隐私设置
// @Summary 获取用户隐私设置
// @Description 获取当前用户的隐私设置
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} Response{data=model.UserPrivacyResponse}
// @Failure 401 {object} Response
// @Router /user/me/privacy [get]
func (h *UserHandler) GetUserPrivacy(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	resp, err := h.userService.GetUserPrivacy(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// UpdateUserPrivacy 更新用户隐私设置
// @Summary 更新用户隐私设置
// @Description 更新当前用户的隐私设置
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body model.UpdateUserPrivacyRequest true "隐私设置请求"
// @Success 200 {object} Response{data=model.UserPrivacyResponse}
// @Failure 401 {object} Response
// @Router /user/me/privacy [put]
func (h *UserHandler) UpdateUserPrivacy(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req model.UpdateUserPrivacyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	resp, err := h.userService.UpdateUserPrivacy(c.Request.Context(), userID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// GetUserCurrency 获取用户货币信息
// @Summary 获取用户货币信息
// @Description 获取当前用户的硬币和签到信息
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} Response{data=model.CurrencyResponse}
// @Failure 401 {object} Response
// @Router /user/me/currency [get]
func (h *UserHandler) GetUserCurrency(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	resp, err := h.userService.GetUserCurrency(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// Checkin 用户签到
// @Summary 用户签到
// @Description 用户每日签到，获得签到奖励
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /user/checkin [post]
func (h *UserHandler) Checkin(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	if err := h.userService.Checkin(c.Request.Context(), userID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

// GetCheckinDates 获取签到日期
// @Summary 获取签到日期
// @Description 获取本月所有签到日期
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} Response{data=[]string}
// @Failure 401 {object} Response
// @Router /user/me/checkin/dates [get]
func (h *UserHandler) GetCheckinDates(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	dates, err := h.userService.GetCheckinDates(c.Request.Context(), userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, dates)
}
