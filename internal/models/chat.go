package models

import (
	"time"

	"gorm.io/gorm"
)

// ChatSession 聊天会话模型
type ChatSession struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`     // 用户ID
	Title     string         `gorm:"type:varchar(200)" json:"title"`    // 会话标题
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`  // 创建时间
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`  // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // 软删除
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	SessionID uint           `gorm:"index;not null" json:"session_id"`      // 会话ID
	Role      string         `gorm:"type:varchar(20);not null" json:"role"` // user/assistant
	Content   string         `gorm:"type:text;not null" json:"content"`     // 消息内容
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`      // 创建时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`     // 软删除
}

// TableName 自定义表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// TableName 自定义表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}
