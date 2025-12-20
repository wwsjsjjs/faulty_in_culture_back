package dto

// SaveGameRequest 存档请求
type SaveGameRequest struct {
	Data string `json:"data" binding:"required"` // 存档数据（JSON字符串）
}

// ChatStartRequest 开始聊天请求
type ChatStartRequest struct {
	Title string `json:"title"` // 聊天标题（可选）
}

// ChatMessageRequest 发送消息请求
type ChatMessageRequest struct {
	SessionID uint   `json:"session_id" binding:"required"` // 聊天会话ID
	Content   string `json:"content" binding:"required"`    // 消息内容
}
