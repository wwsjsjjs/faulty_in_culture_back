package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 用户认证中间件（示例，生产建议用 JWT）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未提供 token"})
			c.Abort()
			return
		}
		// 这里只做简单校验，实际应校验 JWT
		// 示例：token 格式为 userID:username:timestamp
		// 可根据需要解析 token 并设置用户信息到 context
		c.Next()
	}
}
