// Package vo 定义了用于 API 响应的视图对象（VO）。
package vo

// UserVO 用户信息响应 VO
type UserVO struct {
	// ID 用户ID
	ID uint `json:"id"`
	// Username 用户名
	Username string `json:"username"`
}
