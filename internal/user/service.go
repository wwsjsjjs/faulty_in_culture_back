// Package user - 用户模块业务逻辑层
// 功能：实现用户注册、登录、认证等业务规则
// 特点：密码加密、JWT Token生成、缓存管理
package user

import (
	"fmt"
	"time"
)

// ============================================================
// Service层（业务逻辑层）
// 设计模式：服务层模式 (Service Layer Pattern)
// 职责：
// 1. 封装业务逻辑和业务规则
// 2. 协调Repository和其他服务
// 3. 处理事务边界
// 4. 不依赖具体的数据访问技术
// ============================================================

// PasswordHasher 密码哈希接口
// 设计模式：依赖倒置原则 - 依赖抽象而不是具体实现
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}

// TokenGenerator Token生成器接口
type TokenGenerator interface {
	Generate(userID uint, username string) (string, error)
}

// Cache 缓存接口
type Cache interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

// Service 用户业务服务
type Service struct {
	repo           Repository     // 用户仓储
	passwordHasher PasswordHasher // 密码哈希器
	tokenGen       TokenGenerator // Token生成器
	cache          Cache          // 缓存
}

// NewService 创建用户服务实例
// 设计模式：依赖注入 (Dependency Injection)
// 通过构造函数注入依赖，便于测试和解耦
func NewService(repo Repository, hasher PasswordHasher, tokenGen TokenGenerator, cache Cache) *Service {
	return &Service{
		repo:           repo,
		passwordHasher: hasher,
		tokenGen:       tokenGen,
		cache:          cache,
	}
}

// Register 用户注册业务逻辑
// 业务规则：
// 1. 用户名不能重复
// 2. 密码必须加密存储
// 3. 自动设置创建时间和登录时间
func (s *Service) Register(username, password string) (*Entity, string, error) {
	// 检查用户名是否存在
	existUser, err := s.repo.FindByUsername(username)
	if err == nil && existUser != nil {
		return nil, "", fmt.Errorf("用户名已存在")
	}
	if err != nil && err.Error() != "用户不存在" {
		return nil, "", err
	}

	// 密码加密
	hash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, "", fmt.Errorf("密码加密失败")
	}

	// 创建用户实体
	now := time.Now()
	user := &Entity{
		Username:    username,
		Password:    hash,
		CreatedAt:   now,
		LastLoginAt: now,
	}

	// 持久化
	if err := s.repo.Create(user); err != nil {
		return nil, "", fmt.Errorf("创建用户失败")
	}

	// 生成Token
	token, err := s.tokenGen.Generate(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// Login 用户登录业务逻辑
// 业务规则：
// 1. 验证用户名和密码
// 2. 更新最后登录时间
// 3. 使用缓存提高性能
func (s *Service) Login(username, password string) (*Entity, string, error) {
	cacheKey := fmt.Sprintf("user:username:%s", username)

	// 尝试从缓存获取用户
	var user *Entity
	if s.cache != nil {
		var cachedUser Entity
		if err := s.cache.Get(cacheKey, &cachedUser); err == nil && cachedUser.ID != 0 {
			user = &cachedUser
		}
	}

	// 缓存未命中，从数据库查询
	if user == nil {
		var err error
		user, err = s.repo.FindByUsername(username)
		if err != nil {
			return nil, "", fmt.Errorf("用户名或密码错误") // 不暴露用户是否存在
		}

		// 写入缓存
		if s.cache != nil {
			s.cache.Set(cacheKey, user, 24*time.Hour)
		}
	}

	// 验证密码
	if !s.passwordHasher.Check(password, user.Password) {
		return nil, "", fmt.Errorf("用户名或密码错误")
	}

	// 更新最后登录时间
	s.repo.UpdateLastLogin(user.ID)

	// 生成Token
	token, err := s.tokenGen.Generate(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// GetUsername 获取用户名（供其他Service调用）
// 用于Service层组合调用模式
func (s *Service) GetUsername(userID uint) (string, error) {
	// 尝试从缓存获取
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:username:%d", userID)
		var username string
		if err := s.cache.Get(cacheKey, &username); err == nil && username != "" {
			return username, nil
		}
	}

	// 从数据库查询
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", err
	}

	// 写入缓存（缓存30分钟）
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:username:%d", userID)
		s.cache.Set(cacheKey, user.Username, 30*time.Minute)
	}

	return user.Username, nil
}

// GetUsernames 批量获取用户名（优化N+1查询）
// 返回 map[userID]username
func (s *Service) GetUsernames(userIDs []uint) (map[uint]string, error) {
	if len(userIDs) == 0 {
		return map[uint]string{}, nil
	}

	// 去重
	uniqueIDs := make([]uint, 0, len(userIDs))
	idSet := make(map[uint]bool)
	for _, id := range userIDs {
		if !idSet[id] {
			uniqueIDs = append(uniqueIDs, id)
			idSet[id] = true
		}
	}

	// 批量查询数据库
	users, err := s.repo.FindByIDs(uniqueIDs)
	if err != nil {
		return nil, err
	}

	// 构建 userID -> username 映射
	result := make(map[uint]string, len(users))
	for _, user := range users {
		result[user.ID] = user.Username

		// 写入缓存
		if s.cache != nil {
			cacheKey := fmt.Sprintf("user:username:%d", user.ID)
			s.cache.Set(cacheKey, user.Username, 30*time.Minute)
		}
	}

	return result, nil
}
