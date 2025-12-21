package dto

type SaveGameRequest struct {
	GameData string `json:"game_data" binding:"required"`
}

type ChatStartRequest struct {
	Title string `json:"title"`
}

type ChatMessageRequest struct {
	SessionID uint   `json:"session_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
}
