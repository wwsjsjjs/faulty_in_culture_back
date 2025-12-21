package models

import (
	"time"

	"gorm.io/gorm"
)

// Message 消息历史记录模型
type Message struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TaskID      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"task_id"`
	UserID      uint           `gorm:"index:idx_user_created;not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Status      int            `gorm:"not null;default:0" json:"status"`
	CreatedAt   time.Time      `gorm:"index:idx_user_created;autoCreateTime" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 自定义表名
func (Message) TableName() string {
	return "messages"
}
