// Package ranking - 排行榜模块
// 功能：管理用户在不同类型排行榜中的分数记录
// 特点：独立排行榜表，同用户同类型自动保留最高分
package ranking

import (
	"time"
)

// Entity 排行榜实体
// 设计：独立排行榜表，一个用户可以有多个排行榜类型的记录
// 特点：同一用户同一类型只保存最高分
type Entity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_rank_type;not null" json:"user_id"`
	RankType  int       `gorm:"uniqueIndex:idx_user_rank_type;not null" json:"rank_type"` // 1-9种排行榜类型
	Score     int       `gorm:"not null;default:0;index:idx_rank_score" json:"score"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;index:idx_rank_score" json:"updated_at"` // 用于相同分数时的排序
}

// TableName 指定数据库表名
func (Entity) TableName() string {
	return "rankings"
}

// ValidateRankType 验证排行榜类型（1-9）
func ValidateRankType(rankType int) bool {
	return rankType >= 1 && rankType <= 9
}
