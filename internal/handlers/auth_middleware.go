package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 用户认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未提供 token"})
			c.Abort()
			return
		}

		// 解析 token（格式：userID:username:timestamp）
		// 实际应用中应使用 JWT 等更安全的方式
		parts := strings.Split(token, ":")
		if len(parts) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误"})
			c.Abort()
			return
		}

		// 提取用户ID
		userID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 中的用户ID无效"})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", uint(userID))
		c.Set("username", parts[1])

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
