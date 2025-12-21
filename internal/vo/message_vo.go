package vo

import "time"

type MessageResponse struct {
	ID        uint      `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageListResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Messages []MessageResponse `json:"messages"`
}
