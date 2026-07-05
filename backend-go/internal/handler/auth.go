package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	*BaseHandler
	authService service.AuthService
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(authService service.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandler(logger),
		authService: authService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册，返回JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "注册请求"
// @Success 200 {object} Response{data=model.AuthResponse}
// @Failure 400 {object} Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	// 参数绑定和验证
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	// 调用Service
	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// SendVerificationCode 发送邮箱验证码
func (h *AuthHandler) SendVerificationCode(c *gin.Context) {
	var req model.SendVerificationCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.authService.SendVerificationCode(c.Request.Context(), &req); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录，支持用户名/邮箱/电话登录，返回JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "登录请求"
// @Success 200 {object} Response{data=model.AuthResponse}
// @Failure 400 {object} Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	// 参数绑定和验证
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	// 调用Service
	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，需要Bearer token认证
// @Tags 认证
// @Security Bearer
// @Produce json
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从Context获取用户ID
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	// 调用Service
	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}
