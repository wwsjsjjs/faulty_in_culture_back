package vo

import "time"

// RankingResponse 排名响应VO
// 用于返回给前端的排名信息

type RankingResponse struct {
	ID        uint      `json:"id" example:"1"`
	Username  string    `json:"username" example:"player1"`
	Score     int       `json:"score" example:"1000"`
	Rank      int       `json:"rank" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
}

// ErrorResponse 错误响应VO
// 用于 API 错误返回的统一格式

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

// MessageResponse 消息响应VO
// 用于 API 成功或通用消息返回

type MessageResponse struct {
	Message string `json:"message" example:"success"`
}
