package convert

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 类型转换工具包 - 用户ID解析、参数提取
// 职责：提供HTTP上下文中的数据类型转换
// ============================================================

// ParseUserID 从interface{}解析用户ID
func ParseUserID(value interface{}) (uint, bool) {
	switch v := value.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64:
		return uint(v), true
	case string:
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(id), true
		}
	}
	return 0, false
}

// GetUserID 从gin.Context中获取用户ID（从中间件设置的上下文）
func GetUserID(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return ParseUserID(userIDValue)
}

// GetUserIDFromParam 从URL路径参数获取用户ID
// 示例：/api/user/:id -> GetUserIDFromParam(c, "id")
func GetUserIDFromParam(c *gin.Context, key string) (uint, bool) {
	idStr := c.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// GetUserIDFromQuery 从查询参数获取用户ID
// 示例：/api/user?id=123 -> GetUserIDFromQuery(c, "id")
func GetUserIDFromQuery(c *gin.Context, key string) (uint, bool) {
	idStr := c.Query(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// GetUintFromQuery 从查询参数获取uint值（通用）
func GetUintFromQuery(c *gin.Context, key string, defaultValue uint) uint {
	if val, ok := GetUserIDFromQuery(c, key); ok {
		return val
	}
	return defaultValue
}
