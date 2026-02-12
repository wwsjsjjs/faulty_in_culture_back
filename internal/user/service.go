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
		return nil, "", ErrUserAlreadyExists
	}
	if err != nil && err != ErrUserNotFound {
		return nil, "", err
	}

	// 密码加密
	hash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, "", ErrHashPassword
	}

	// 创建用户实体
	now := time.Now()
	user := &Entity{
		Username:       username,
		Password:       hash,
		CreatedAt:      now,
		LastLoginAt:    now,
		Score:          0,
		ScoreUpdatedAt: now,
	}

	// 持久化
	if err := s.repo.Create(user); err != nil {
		return nil, "", ErrCreateUser
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
			return nil, "", ErrInvalidPassword // 不暴露用户是否存在
		}

		// 写入缓存
		if s.cache != nil {
			s.cache.Set(cacheKey, user, 24*time.Hour)
		}
	}

	// 验证密码
	if !s.passwordHasher.Check(password, user.Password) {
		return nil, "", ErrInvalidPassword
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

// UpdateScore 更新用户分数
// 业务规则：
// 1. 验证排行榜类型(1-9)
// 2. 分数必须>=0
// 3. 更新后清除相关缓存
func (s *Service) UpdateScore(userID uint, rankType, score int) error {
	// 验证排行榜类型
	if !ValidateRankType(rankType) {
		return ErrInvalidRankType
	}

	// 验证分数
	if score < 0 {
		return ErrInvalidScore
	}

	// 更新分数（Repository内部使用策略模式）
	if err := s.repo.UpdateScore(userID, rankType, score); err != nil {
		return ErrUpdateScore
	}

	// 清除排行榜缓存
	if s.cache != nil {
		s.clearRankingCache(rankType)
	}

	return nil
}

// GetRankings 获取排行榜
// 业务规则：
// 1. 验证排行榜类型
// 2. 支持分页
// 3. 使用缓存提高性能
func (s *Service) GetRankings(rankType, page, limit int) ([]RankingItem, error) {
	// 验证排行榜类型
	if !ValidateRankType(rankType) {
		return nil, ErrInvalidRankType
	}

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	cacheKey := fmt.Sprintf("rankings:type:%d:page:%d:limit:%d", rankType, page, limit)

	// 尝试从缓存获取
	var rankings []RankingItem
	if s.cache != nil {
		if err := s.cache.Get(cacheKey, &rankings); err == nil && len(rankings) > 0 {
			return rankings, nil
		}
	}

	// 从数据库查询
	users, err := s.repo.GetRankings(rankType, offset, limit)
	if err != nil {
		return nil, err
	}

	// 使用策略模式获取对应分数
	strategy := GetScoreStrategy(rankType)
	rankings = make([]RankingItem, len(users))
	for i, u := range users {
		rankings[i] = RankingItem{
			ID:       u.ID,
			Username: u.Username,
			Score:    strategy.GetScore(u),
			Rank:     offset + i + 1,
		}
	}

	// 写入缓存
	if s.cache != nil {
		s.cache.Set(cacheKey, rankings, 5*time.Minute)
	}

	return rankings, nil
}

// clearRankingCache 清除排行榜缓存
// 清除指定排行榜类型的所有分页缓存
func (s *Service) clearRankingCache(rankType int) {
	if s.cache == nil {
		return
	}

	// 清除常见的分页缓存
	for page := 1; page <= 10; page++ {
		for limit := 10; limit <= 100; limit += 10 {
			cacheKey := fmt.Sprintf("rankings:type:%d:page:%d:limit:%d", rankType, page, limit)
			s.cache.Delete(cacheKey)
		}
	}
}
