// Package user - 用户模块数据传输对象
// 功能：定义API请求和响应的数据结构
package user

// ============================================================
// 请求DTO (Data Transfer Objects)
// 设计模式：DTO模式 - 用于在不同层之间传输数据
// ============================================================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"player1"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"player1"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// ============================================================
// 响应VO (Value Objects)
// 设计模式：值对象模式 - 不可变的数据对象，用于API响应
// ============================================================

// UserVO 用户信息值对象
type UserVO struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"player1"`
}

// AuthResponse 统一认证响应（注册和登录）
type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserVO `json:"user"`
}
