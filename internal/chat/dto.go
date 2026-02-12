package chat

import "time"

// ============ 请求DTO ============

// StartRequest 开始聊天请求
type StartRequest struct {
	Title string `json:"title" example:"新对话"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	SessionID uint   `json:"session_id" binding:"required" example:"1"`
	Content   string `json:"content" binding:"required" example:"你好"`
}

// ============ 响应VO ============

// SessionVO 会话值对象
type SessionVO struct {
	ID        uint      `json:"id" example:"1"`
	UserID    uint      `json:"user_id" example:"1"`
	Title     string    `json:"title" example:"我的对话"`
	Type      int       `json:"type" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-12-20T10:00:00Z"`
}

// MessageVO 消息值对象
type MessageVO struct {
	ID           uint      `json:"id" example:"1"`
	SessionID    uint      `json:"session_id" example:"1"`
	MessageIndex int       `json:"message_index" example:"1"` // 1=用户, 2=AI, 3=用户...
	Content      string    `json:"content" example:"你好"`
	CreatedAt    time.Time `json:"created_at" example:"2023-12-20T10:00:00Z"`
}

// HistoryResponse 聊天历史响应
type HistoryResponse struct {
	Session  SessionVO   `json:"session"`
	Messages []MessageVO `json:"messages"`
}

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type         string `json:"type"` // "ai_message"
	SessionID    uint   `json:"session_id"`
	Content      string `json:"content"`
	MessageIndex int    `json:"message_index"` // AI消息序号（偶数）
}
