package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"faulty_in_culture/go_back/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware 用户认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("middleware.AuthMiddleware",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		token := c.GetHeader("Authorization")
		if token == "" {
			logger.Warn("middleware.AuthMiddleware: 未提供 token",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未提供 token"})
			c.Abort()
			return
		}

		// 解析 token（格式：userID:username:timestamp）
		// 实际应用中应使用 JWT 等更安全的方式
		parts := strings.Split(token, ":")
		if len(parts) < 2 {
			logger.Warn("middleware.AuthMiddleware: token 格式错误",
				zap.String("token", token),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误"})
			c.Abort()
			return
		}

		// 提取用户ID
		userID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			logger.Error("middleware.AuthMiddleware: 解析用户ID失败",
				zap.String("token", token),
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 中的用户ID无效"})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", uint(userID))
		c.Set("username", parts[1])

		logger.Info("middleware.AuthMiddleware: 认证成功",
			zap.Uint("user_id", uint(userID)),
			zap.String("username", parts[1]),
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
