// Package dto 定义了用户相关的数据传输对象（DTO），用于接口请求和响应参数。
package dto

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type UpdateScoreRequest struct {
	Score int `json:"score" binding:"min=0"`
}
