// Package user 用户领域模块
// 包含用户实体、业务逻辑、数据访问和HTTP处理
package user

import (
	"time"

	"gorm.io/gorm"
)

// Entity 用户实体
// 设计模式：实体模式（Entity Pattern） - DDD中的核心领域对象
// 职责：表示用户领域模型，封装用户的属性和行为
type Entity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"` // 密码不返回给前端
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`

	// 9种排行榜分数（策略模式支持）
	Score1 int `gorm:"not null;default:0;index" json:"score1"`
	Score2 int `gorm:"not null;default:0;index" json:"score2"`
	Score3 int `gorm:"not null;default:0;index" json:"score3"`
	Score4 int `gorm:"not null;default:0;index" json:"score4"`
	Score5 int `gorm:"not null;default:0;index" json:"score5"`
	Score6 int `gorm:"not null;default:0;index" json:"score6"`
	Score7 int `gorm:"not null;default:0;index" json:"score7"`
	Score8 int `gorm:"not null;default:0;index" json:"score8"`
	Score9 int `gorm:"not null;default:0;index" json:"score9"`

	Score          int            `gorm:"not null;default:0;index" json:"score"` // 保留字段
	ScoreUpdatedAt time.Time      `json:"score_updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据库表名
// GORM约定：实现TableName()接口自定义表名
func (Entity) TableName() string {
	return "users"
}

// GetScoreByType 根据排行榜类型获取对应分数
// 设计模式：工厂方法模式 - 根据类型返回不同的分数
func (e *Entity) GetScoreByType(rankType int) int {
	switch rankType {
	case 1:
		return e.Score1
	case 2:
		return e.Score2
	case 3:
		return e.Score3
	case 4:
		return e.Score4
	case 5:
		return e.Score5
	case 6:
		return e.Score6
	case 7:
		return e.Score7
	case 8:
		return e.Score8
	case 9:
		return e.Score9
	default:
		return 0
	}
}

// SetScoreByType 根据排行榜类型设置对应分数
// 设计模式：工厂方法模式 - 根据类型设置不同的分数
func (e *Entity) SetScoreByType(rankType, score int) {
	switch rankType {
	case 1:
		e.Score1 = score
	case 2:
		e.Score2 = score
	case 3:
		e.Score3 = score
	case 4:
		e.Score4 = score
	case 5:
		e.Score5 = score
	case 6:
		e.Score6 = score
	case 7:
		e.Score7 = score
	case 8:
		e.Score8 = score
	case 9:
		e.Score9 = score
	}
	e.ScoreUpdatedAt = time.Now()
}
