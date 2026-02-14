// Package chat - AI聊天模块
// 功能：管理用户与AI的对话会话和消息记录
// 架构：会话管理 + 消息管理
package chat

import (
	"time"
)

// Session 聊天会话实体
type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Type      int       `gorm:"not null;default:1" json:"type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Session) TableName() string { return "chat_sessions" }

// Message 聊天消息实体
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID uint      `gorm:"index:idx_session_created;not null" json:"session_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_session_created;autoCreateTime" json:"created_at"`
}

func (Message) TableName() string { return "chat_messages" }

// IsUserMessage 判断是否是用户消息（序号为奇数）
func (m *Message) IsUserMessage(index int) bool {
	return index%2 == 1
}
