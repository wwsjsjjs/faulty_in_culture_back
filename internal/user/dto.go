package user

// ============================================================
// 请求DTO (Data Transfer Objects)
// 设计模式：DTO模式 - 用于在不同层之间传输数据
// ============================================================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"player1"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"player1"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// UpdateScoreRequest 更新分数请求
type UpdateScoreRequest struct {
	RankType int `json:"rank_type" binding:"required,min=1,max=9" example:"1"` // 排行榜类型1-9
	Score    int `json:"score" binding:"min=0" example:"100"`                  // 分数
}

// ============================================================
// 响应VO (Value Objects)
// 设计模式：值对象模式 - 不可变的数据对象，用于API响应
// ============================================================

// UserVO 用户信息值对象
type UserVO struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"player1"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"player1"`
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserVO `json:"user"`
}

// RankingItem 排名项
type RankingItem struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"player1"`
	Score    int    `json:"score" example:"1000"`
	Rank     int    `json:"rank" example:"1"`
}

// RankingListResponse 排名列表响应
type RankingListResponse struct {
	Page     int           `json:"page" example:"1"`
	Limit    int           `json:"limit" example:"10"`
	Rankings []RankingItem `json:"rankings"`
}

// SuccessResponse 通用成功响应
type SuccessResponse struct {
	Message string `json:"message" example:"success"`
}

// ErrorResponse 通用错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
