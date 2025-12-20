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
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`
	Username  string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username" example:"user1"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-" example:"$2a$10$..."`
	CreatedAt time.Time      `json:"created_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-12-20T10:00:00Z" format:"date-time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 User 结构体对应的数据库表名
func (User) TableName() string {
	return "users"
}
