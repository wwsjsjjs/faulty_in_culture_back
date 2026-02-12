package middleware

import (
	"net/http"
	"strings"

	"faulty_in_culture/go_back/internal/shared/infra/logger"
	"faulty_in_culture/go_back/internal/shared/pkg/security"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware 用户认证中间件（使用 JWT 框架验证）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("middleware.AuthMiddleware",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		// 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("middleware.AuthMiddleware: 未提供 token",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未提供 token"})
			c.Abort()
			return
		}

		// 提取 Bearer Token
		// 格式：Authorization: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("middleware.AuthMiddleware: token 格式错误，应为 'Bearer <token>'",
				zap.String("authHeader", authHeader),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 使用 JWT 框架解析和验证 Token
		claims, err := security.ParseToken(tokenString)
		if err != nil {
			logger.Error("middleware.AuthMiddleware: token 验证失败",
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		logger.Info("middleware.AuthMiddleware: 认证成功（JWT）",
			zap.Uint("user_id", claims.UserID),
			zap.String("username", claims.Username),
		)

		c.Next()
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}
