package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/pkg/auth"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(tokenManager *auth.TokenManager, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, map[string]interface{}{
				"success": false,
				"code":    int(errors.CodeUnauthorized),
				"message": "缺少Authorization header",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, map[string]interface{}{
				"success": false,
				"code":    int(errors.CodeInvalidToken),
				"message": "无效的token格式",
				"data":    nil,
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		claims, err := tokenManager.ValidateToken(token)
		if err != nil {
			c.JSON(401, map[string]interface{}{
				"success": false,
				"code":    int(errors.CodeInvalidToken),
				"message": "token无效或已过期",
				"data":    nil,
			})
			c.Abort()
			logger.Warn("Token验证失败", zap.Error(err))
			return
		}

		// 将用户信息注入到Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// LoggerMiddleware 日志记录中间件
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求信息
		logger.Info("请求开始",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()

		// 记录响应信息
		logger.Info("请求结束",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("请求异常", zap.Any("error", err))
				c.JSON(500, map[string]interface{}{
					"success": false,
					"code":    int(errors.CodeInternalError),
					"message": "服务器内部错误",
					"data":    nil,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
