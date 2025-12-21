package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表模型
// @Description 用户信息
// @Tags User
// @TableName users
// @Param username string 用户名
// @Param password string 密码（加密存储）
type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Password       string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	LastLoginAt    time.Time      `json:"last_login_at"`
	Score          int            `gorm:"not null;default:0;index" json:"score"`
	ScoreUpdatedAt time.Time      `json:"score_updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 User 结构体对应的数据库表名
func (User) TableName() string {
	return "users"
}
