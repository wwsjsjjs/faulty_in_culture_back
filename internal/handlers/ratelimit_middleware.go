package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ranking-api/internal/cache"
)

// RateLimitMiddleware 频率限制中间件
// limitKey: 限制的键名（如 "register", "login", "savegame"）
// maxRequests: 允许的最大请求数
// window: 时间窗口
func RateLimitMiddleware(limitKey string, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheClient := cache.GetCache()
		if cacheClient == nil {
			// 缓存不可用，跳过限流
			c.Next()
			return
		}

		// 获取用户标识（可以是 IP 或用户ID）
		identifier := c.ClientIP()

		// 如果已登录，使用用户ID
		if userID, exists := GetUserID(c); exists {
			identifier = fmt.Sprintf("user_%d", userID)
		}

		// 构建 Redis key
		key := fmt.Sprintf("ratelimit:%s:%s", limitKey, identifier)

		// 获取当前计数
		var count int
		err := cacheClient.Get(key, &count)
		if err != nil {
			// 第一次请求，设置计数为1
			count = 1
			cacheClient.Set(key, count, window)
		} else {
			// 检查是否超过限制
			if count >= maxRequests {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": fmt.Sprintf("操作过于频繁，请%v后再试", window),
				})
				c.Abort()
				return
			}
			// 增加计数
			count++
			cacheClient.Set(key, count, window)
		}

		c.Next()
	}
}

// PerUserRateLimitMiddleware 针对单个用户的频率限制（必须先认证）
func PerUserRateLimitMiddleware(limitKey string, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		cacheClient := cache.GetCache()
		if cacheClient == nil {
			// 缓存不可用，跳过限流
			c.Next()
			return
		}

		// 构建 Redis key
		key := fmt.Sprintf("ratelimit:%s:user_%d", limitKey, userID)

		// 获取当前计数
		var count int
		err := cacheClient.Get(key, &count)
		if err != nil {
			// 第一次请求
			count = 1
			cacheClient.Set(key, count, window)
		} else {
			if count >= maxRequests {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": fmt.Sprintf("操作过于频繁，请%v后再试", window),
				})
				c.Abort()
				return
			}
			count++
			cacheClient.Set(key, count, window)
		}

		c.Next()
	}
}
