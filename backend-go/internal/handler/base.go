package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data"`
}

// BaseHandler 基础处理器
type BaseHandler struct {
	logger *logger.Logger
}

// NewBaseHandler 创建新的基础处理器
func NewBaseHandler(logger *logger.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

// Success 返回成功响应
func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "操作成功",
		Data:    data,
	})
}

// Error 返回错误响应
func (h *BaseHandler) Error(c *gin.Context, err error) {
	h.logger.Error("请求处理失败", zap.Error(err))
	be := errors.AsBusinessError(err)
	if be != nil {
		// 业务错误
		statusCode := h.getStatusCodeByErrorCode(be.Code)
		c.JSON(statusCode, Response{
			Success: false,
			Code:    int(be.Code),
			Message: be.Message,
			Data:    nil,
		})
		h.logger.Warn(be.Error())
		return
	}

	// 未知错误
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    errors.CodeInternalError,
		"message": "服务器内部错误",
	})
	h.logger.Error("未知错误", zap.Error(err))
}

// BadRequest 返回参数错误
func (h *BaseHandler) BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Code:    int(errors.CodeValidationError),
		Message: message,
		Data:    nil,
	})
}

// Unauthorized 返回未授权
func (h *BaseHandler) Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Code:    int(errors.CodeUnauthorized),
		Message: errors.GetErrorMessage(errors.CodeUnauthorized),
		Data:    nil,
	})
}

// Forbidden 返回禁止访问
func (h *BaseHandler) Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Code:    int(errors.CodeForbidden),
		Message: errors.GetErrorMessage(errors.CodeForbidden),
		Data:    nil,
	})
}

// getStatusCodeByErrorCode 根据错误码获取HTTP状态码
func (h *BaseHandler) getStatusCodeByErrorCode(code errors.ErrorCode) int {
	switch code {
	case errors.CodeValidationError, errors.CodeInvalidRequest, errors.CodeMissingParam:
		return http.StatusBadRequest
	case errors.CodeUnauthorized, errors.CodeInvalidToken, errors.CodeTokenExpired:
		return http.StatusUnauthorized
	case errors.CodeForbidden, errors.CodeNotRoomOwner:
		return http.StatusForbidden
	case errors.CodeUserNotFound, errors.CodeTaskNotFound, errors.CodeRoomNotFound, errors.CodeFriendNotFound:
		return http.StatusNotFound
	case errors.CodeInternalError, errors.CodeDatabaseError, errors.CodeBusinessError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// GetUserIDFromContext 从Context获取用户ID
func GetUserIDFromContext(c *gin.Context) (int64, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New(errors.CodeUnauthorized, "用户未认证")
	}

	id, ok := userID.(int64)
	if !ok {
		return 0, errors.New(errors.CodeUnauthorized, "用户ID格式错误")
	}

	return id, nil
}

// GetUsernameFromContext 从Context获取用户名
func GetUsernameFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", errors.New(errors.CodeUnauthorized, "用户未认证")
	}

	name, ok := username.(string)
	if !ok {
		return "", errors.New(errors.CodeUnauthorized, "用户名格式错误")
	}

	return name, nil
}
