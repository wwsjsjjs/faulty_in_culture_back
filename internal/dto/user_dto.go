// Package dto 定义了用户相关的数据传输对象（DTO），用于接口请求和响应参数。
package dto

// UserRegisterRequest 用户注册请求体
type UserRegisterRequest struct {
	// Username 用户名
	Username string `json:"username" binding:"required"`
	// Password 密码
	Password string `json:"password" binding:"required"`
}

// UserLoginRequest 用户登录请求体
type UserLoginRequest struct {
	// Username 用户名
	Username string `json:"username" binding:"required"`
	// Password 密码
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户响应体
type UserResponse struct {
	// ID 用户ID
	ID uint `json:"id"`
	// Username 用户名
	Username string `json:"username"`
}
