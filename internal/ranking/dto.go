// Package ranking - 排行榜模块数据传输对象
// 功能：定义API请求和响应的数据结构
package ranking

import "time"

// ============ 请求DTO ============

// UpdateScoreRequest 更新分数请求
type UpdateScoreRequest struct {
	RankType int `json:"rank_type" binding:"required,min=1,max=9" example:"1"` // 排行榜类型1-9
	Score    int `json:"score" binding:"required,min=0" example:"100"`         // 分数
}

// ============ 响应VO ============

// RankingItem 排行榜项
type RankingItem struct {
	Rank      int       `json:"rank" example:"1"`                          // 排名
	UserID    uint      `json:"user_id" example:"1"`                       // 用户ID
	Username  string    `json:"username" example:"player1"`                // 用户名
	Score     int       `json:"score" example:"100"`                       // 分数
	UpdatedAt time.Time `json:"updated_at" example:"2023-12-20T10:00:00Z"` // 更新时间
}

// RankingListResponse 排行榜列表响应
type RankingListResponse struct {
	RankType int           `json:"rank_type" example:"1"` // 排行榜类型
	Page     int           `json:"page" example:"1"`      // 当前页
	Limit    int           `json:"limit" example:"10"`    // 每页数量
	Rankings []RankingItem `json:"rankings"`              // 排行榜数据
}

// UpdateScoreResponse 更新分数响应
type UpdateScoreResponse struct {
	UserID    uint      `json:"user_id" example:"1"`
	RankType  int       `json:"rank_type" example:"1"`
	Score     int       `json:"score" example:"100"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-12-20T10:00:00Z"`
}
