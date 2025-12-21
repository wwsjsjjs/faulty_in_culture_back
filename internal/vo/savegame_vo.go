package vo

import "time"

type SaveGameResponse struct {
	UserID     uint      `json:"user_id"`
	SlotNumber int       `json:"slot_number"`
	GameData   string    `json:"game_data"`
	SavedAt    time.Time `json:"saved_at"`
}

type ChatSessionResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Type      int       `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	SessionID uint      `json:"session_id"`
	Role      int       `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatHistoryResponse struct {
	Session  ChatSessionResponse   `json:"session"`
	Messages []ChatMessageResponse `json:"messages"`
}
