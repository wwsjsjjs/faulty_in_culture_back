package vo

import "time"

// SaveGameResponse 存档响应
type SaveGameResponse struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	SlotNumber int       `json:"slot_number"`
	Data       string    `json:"data"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ChatSessionResponse 聊天会话响应
type ChatSessionResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatMessageResponse 聊天消息响应
type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	SessionID uint      `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatHistoryResponse 聊天历史响应
type ChatHistoryResponse struct {
	Session  ChatSessionResponse   `json:"session"`
	Messages []ChatMessageResponse `json:"messages"`
}
