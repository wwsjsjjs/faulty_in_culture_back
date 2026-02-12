package response

import "github.com/gin-gonic/gin"

// Success 统一成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 200, "data": data})
}

// Error 统一错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"code": code, "error": message})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

// ServerError 500错误
func ServerError(c *gin.Context, message string) {
	Error(c, 500, message)
}
