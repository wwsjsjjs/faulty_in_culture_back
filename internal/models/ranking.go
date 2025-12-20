package models

import (
	"time"

	"gorm.io/gorm"
)

// Ranking 排名表模型
// @Description 用户排名信息
type Ranking struct {
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`                                         // 主键ID
	Username  string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username" example:"player1"` // 用户名（唯一）
	Score     int            `gorm:"not null;default:0;index" json:"score" example:"1000"`                     // 分数
	CreatedAt time.Time      `json:"created_at" example:"2023-12-20T10:00:00Z"`                                // 创建时间
	UpdatedAt time.Time      `json:"updated_at" example:"2023-12-20T10:00:00Z"`                                // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                                           // 软删除
}

// TableName 指定表名
func (Ranking) TableName() string {
	return "rankings"
}

// CreateRankingRequest 创建排名请求
type CreateRankingRequest struct {
	Username string `json:"username" binding:"required,min=1,max=100" example:"player1"` // 用户名
	Score    int    `json:"score" binding:"required,min=0" example:"1000"`               // 分数
}

// UpdateRankingRequest 更新排名请求
type UpdateRankingRequest struct {
	Username string `json:"username" binding:"omitempty,min=1,max=100" example:"player1"` // 用户名（可选）
	Score    *int   `json:"score" binding:"omitempty,min=0" example:"1500"`               // 分数（可选）
}

// RankingResponse 排名响应
type RankingResponse struct {
	ID        uint      `json:"id" example:"1"`                            // 主键ID
	Username  string    `json:"username" example:"player1"`                // 用户名
	Score     int       `json:"score" example:"1000"`                      // 分数
	Rank      int       `json:"rank" example:"1"`                          // 排名
	CreatedAt time.Time `json:"created_at" example:"2023-12-20T10:00:00Z"` // 创建时间
	UpdatedAt time.Time `json:"updated_at" example:"2023-12-20T10:00:00Z"` // 更新时间
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"` // 错误信息
}

// MessageResponse 消息响应
type MessageResponse struct {
	Message string `json:"message" example:"success"` // 消息
}
