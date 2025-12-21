package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatSession struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Type      int            `gorm:"not null;default:1" json:"type"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ChatMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	SessionID uint           `gorm:"index:idx_session_created;not null" json:"session_id"`
	Session   ChatSession    `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Role      int            `gorm:"not null;default:1" json:"role"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `gorm:"index:idx_session_created;autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 自定义表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// TableName 自定义表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}
