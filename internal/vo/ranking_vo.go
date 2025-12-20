// Package vo 定义了用于 API 响应的视图对象（VO）。
package vo

import "time"

// RankingResponse 排名响应 VO
// 用于返回给前端的排名信息
type RankingResponse struct {
	// ID 排名ID
	ID uint `json:"id" example:"1"`
	// Username 用户名
	Username string `json:"username" example:"player1"`
	// Score 分数
	Score int `json:"score" example:"1000"`
	// Rank 排名
	Rank int `json:"rank" example:"1"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
}

// ErrorResponse 错误响应 VO
// 用于 API 错误返回的统一格式
type ErrorResponse struct {
	// Error 错误信息
	Error string `json:"error" example:"invalid request"`
}

// SuccessMessageResponse 成功消息响应 VO
// 用于 API 成功或通用消息返回
type SuccessMessageResponse struct {
	// Message 消息内容
	Message string `json:"message" example:"success"`
}
