package dto

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// GetMessagesRequest 获取消息列表请求
type GetMessagesRequest struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Status string `form:"status" binding:"omitempty,oneof=pending completed failed"`
}
