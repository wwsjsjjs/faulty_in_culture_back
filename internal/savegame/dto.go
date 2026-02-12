package savegame

import "time"

// ============ 请求DTO ============

// QueryRequest 查询存档请求
type QueryRequest struct {
	SlotNumber int `form:"slot_number" binding:"min=1,max=6" example:"1"`
}

// CreateRequest 创建/更新存档请求
type CreateRequest struct {
	SlotNumber int    `json:"slot_number" binding:"required,min=1,max=6" example:"1"`
	GameData   string `json:"game_data" binding:"required" example:"{\"level\":5}"`
}

// UpdateRequest 更新存档请求
type UpdateRequest struct {
	GameData string `json:"game_data" binding:"required" example:"{\"level\":10}"`
}

// ============ 响应VO ============

// SaveGameVO 存档值对象
type SaveGameVO struct {
	UserID     uint      `json:"user_id" example:"1"`
	SlotNumber int       `json:"slot_number" example:"1"`
	GameData   string    `json:"game_data" example:"{\"level\":5}"`
	SavedAt    time.Time `json:"saved_at" example:"2023-12-20T10:00:00Z"`
}

// SaveGameListResponse 存档列表响应
type SaveGameListResponse struct {
	Total int          `json:"total" example:"3"`
	List  []SaveGameVO `json:"list"`
}
