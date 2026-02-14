// Package chat - AI聊天模块数据访问层
// 功能：封装会话和消息的CRUD操作
// 设计模式：Repository模式
package chat

import (
	"fmt"

	"gorm.io/gorm"
)

// Repository 聊天仓储接口
type Repository interface {
	// Session相关
	CreateSession(session *Session) error
	FindSessionByID(id uint) (*Session, error)
	ListSessionsByUserID(userID uint, offset, limit int) ([]*Session, error)
	UpdateSession(session *Session) error
	DeleteSession(id uint) error

	// Message相关
	CreateMessage(message *Message) error
	FindMessagesBySessionID(sessionID uint, offset, limit int) ([]*Message, error)
	CountMessagesBySessionID(sessionID uint) (int64, error)
	DeleteMessages(sessionID uint, messageIDs []uint) error
}

// repositoryImpl Repository的GORM实现
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository 创建聊天仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// CreateSession 创建会话
func (r *repositoryImpl) CreateSession(session *Session) error {
	return r.db.Create(session).Error
}

// FindSessionByID 根据ID查找会话
func (r *repositoryImpl) FindSessionByID(id uint) (*Session, error) {
	var session Session
	err := r.db.First(&session, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, err
	}
	return &session, nil
}

// ListSessionsByUserID 根据用户ID获取会话列表
func (r *repositoryImpl) ListSessionsByUserID(userID uint, offset, limit int) ([]*Session, error) {
	var sessions []*Session
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新会话
func (r *repositoryImpl) UpdateSession(session *Session) error {
	return r.db.Save(session).Error
}

// DeleteSession 删除会话（软删除）
func (r *repositoryImpl) DeleteSession(id uint) error {
	return r.db.Delete(&Session{}, id).Error
}

// CreateMessage 创建消息
func (r *repositoryImpl) CreateMessage(message *Message) error {
	return r.db.Create(message).Error
}

// FindMessagesBySessionID 根据会话ID查找消息
func (r *repositoryImpl) FindMessagesBySessionID(sessionID uint, offset, limit int) ([]*Message, error) {
	var messages []*Message
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// CountMessagesBySessionID 统计会话的消息数量
func (r *repositoryImpl) CountMessagesBySessionID(sessionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Message{}).Where("session_id = ?", sessionID).Count(&count).Error
	return count, err
}

// DeleteMessages 批量删除消息（软删除）
func (r *repositoryImpl) DeleteMessages(sessionID uint, messageIDs []uint) error {
	return r.db.Where("session_id = ? AND id IN ?", sessionID, messageIDs).Delete(&Message{}).Error
}
