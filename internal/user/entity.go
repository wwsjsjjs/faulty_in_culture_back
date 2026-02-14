// Package user - 用户领域模块
// 功能：用户注册、登录、认证管理
// 架构：MVC分层架构，包含实体、仓储、服务、处理器
package user

import (
	"time"
)

// ============================================================
// 实体层 (Entity Layer)
// 职责：表示用户领域模型，封装用户的属性和行为
type Entity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"` // 密码不返回给前端
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// TableName 指定数据库表名
// GORM约定：实现TableName()接口自定义表名
func (Entity) TableName() string {
	return "users"
}
