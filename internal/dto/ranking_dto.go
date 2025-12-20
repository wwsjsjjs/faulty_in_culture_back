package dto

// CreateRankingRequest 创建排名请求DTO
// 用于 POST /api/rankings 创建排名时的请求体
// 只包含前端提交的参数

type CreateRankingRequest struct {
	Username string `json:"username" binding:"required,min=1,max=100" example:"player1"`
	Score    int    `json:"score" binding:"required,min=0" example:"1000"`
}

// UpdateRankingRequest 更新排名请求DTO
// 用于 PUT /api/rankings/:id 更新排名时的请求体
// 只包含前端可提交的参数

type UpdateRankingRequest struct {
	Username string `json:"username" binding:"omitempty,min=1,max=100" example:"player1"`
	Score    *int   `json:"score" binding:"omitempty,min=0" example:"1500"`
}
