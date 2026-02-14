// Package response 提供统一的HTTP响应格式
// 功能：封装标准化的JSON响应结构，与errcode包配合使用
package response

import (
	errcode "faulty_in_culture/go_back/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`           // 业务错误码
	Message string      `json:"message"`        // 错误消息
	Data    interface{} `json:"data,omitempty"` // 响应数据（可选）
}

// Success 成功响应（带数据）
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    errcode.Success,
		Message: errcode.GetMessage(errcode.Success),
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    errcode.Success,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应（使用错误码）
func Error(c *gin.Context, httpCode int, errCode int) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: errcode.GetMessage(errCode),
	})
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: message,
	})
}

// ErrorWithData 错误响应（带数据）
func ErrorWithData(c *gin.Context, httpCode int, errCode int, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: errcode.GetMessage(errCode),
		Data:    data,
	})
}

// BadRequest 400参数错误
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, Response{
		Code:    errcode.InvalidParams,
		Message: message,
	})
}

// Unauthorized 401未授权
func Unauthorized(c *gin.Context) {
	Error(c, 401, errcode.Unauthorized)
}

// NotFound 404资源不存在
func NotFound(c *gin.Context) {
	Error(c, 404, errcode.NotFound)
}

// ServerError 500服务器错误
func ServerError(c *gin.Context, message string) {
	c.JSON(500, Response{
		Code:    errcode.ServerError,
		Message: message,
	})
}
