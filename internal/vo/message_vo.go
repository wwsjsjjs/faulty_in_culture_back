package vo

import "time"

// MessageResponse 消息响应
type MessageResponse struct {
	ID          uint       `json:"id"`
	TaskID      string     `json:"task_id"`
	UserID      string     `json:"user_id"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Messages []MessageResponse `json:"messages"`
}
