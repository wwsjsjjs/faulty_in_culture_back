// Package user - 用户模块数据访问层
// 功能：封装用户数据的CRUD操作
// 设计模式：Repository模式
package user

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// Repository模式 (Repository Pattern)
// 设计模式：仓储模式 - 封装数据访问逻辑，提供集合式的接口
// 职责：
// 1. 封装所有数据库操作
// 2. 提供领域对象的持久化和查询
// 3. 隐藏底层数据存储细节
// 优点：
// 1. 分离业务逻辑和数据访问逻辑
// 2. 便于单元测试（可Mock）
// 3. 集中管理数据访问代码
// ============================================================

// Repository 用户仓储接口
type Repository interface {
	// Create 创建用户
	Create(user *Entity) error
	// FindByID 根据ID查找用户
	FindByID(id uint) (*Entity, error)
	// FindByUsername 根据用户名查找用户
	FindByUsername(username string) (*Entity, error)
	// Update 更新用户信息
	Update(user *Entity) error
	// UpdateLastLogin 更新最后登录时间
	UpdateLastLogin(userID uint) error
	// FindByIDs 批量查询用户（优化N+1查询）
	FindByIDs(userIDs []uint) ([]*Entity, error)
}

// repositoryImpl Repository的GORM实现
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository 创建用户仓储实例
// 设计模式：工厂方法模式 - 创建Repository实例
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// Create 创建用户
func (r *repositoryImpl) Create(user *Entity) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *repositoryImpl) FindByID(id uint) (*Entity, error) {
	var user Entity
	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *repositoryImpl) FindByUsername(username string) (*Entity, error) {
	var user Entity
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *repositoryImpl) Update(user *Entity) error {
	return r.db.Save(user).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *repositoryImpl) UpdateLastLogin(userID uint) error {
	return r.db.Model(&Entity{}).Where("id = ?", userID).
		Update("last_login_at", time.Now()).Error
}

// FindByIDs 批量查询用户（优化N+1查询）
func (r *repositoryImpl) FindByIDs(userIDs []uint) ([]*Entity, error) {
	if len(userIDs) == 0 {
		return []*Entity{}, nil
	}

	var users []*Entity
	err := r.db.Where("id IN ?", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
