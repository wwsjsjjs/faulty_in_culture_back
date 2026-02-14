// Package user - 用户模块适配器
// 功能：提供密码加密器、Token生成器等基础设施适配器
package user

import (
	"faulty_in_culture/go_back/internal/shared/security"
)

// ============================================================
// 安全适配器 - 实现密码加密和Token生成接口
// ============================================================

// PasswordHasherAdapter 密码哈希器适配器
type PasswordHasherAdapter struct{}

// NewPasswordHasher 创建密码哈希器实例
func NewPasswordHasher() PasswordHasher {
	return &PasswordHasherAdapter{}
}

// Hash 加密密码
func (p *PasswordHasherAdapter) Hash(password string) (string, error) {
	return security.HashPassword(password)
}

// Check 验证密码
func (p *PasswordHasherAdapter) Check(password, hash string) bool {
	return security.CheckPassword(password, hash)
}

// TokenGeneratorAdapter Token生成器适配器
type TokenGeneratorAdapter struct{}

// NewTokenGenerator 创建Token生成器实例
func NewTokenGenerator() TokenGenerator {
	return &TokenGeneratorAdapter{}
}

// Generate 生成JWT Token
func (t *TokenGeneratorAdapter) Generate(userID uint, username string) (string, error) {
	return security.GenerateToken(userID, username)
}
