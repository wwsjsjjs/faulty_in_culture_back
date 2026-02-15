package user

import (
	"fmt"
	"time"

	"faulty_in_culture/go_back/internal/infra/logger"
	"faulty_in_culture/go_back/internal/shared/errors"

	"go.uber.org/zap"
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
	logger.Info("[user.Register] 开始注册", zap.String("username", username))
	
	// 检查用户名是否存在
	existUser, err := s.repo.FindByUsername(username)
	if err == nil && existUser != nil {
		logger.Warn("[user.Register] 用户已存在", zap.String("username", username))
		return nil, "", errors.New(errors.UserAlreadyExists)
	}
	if err != nil && err.Error() != "用户不存在" {
		logger.Error("[user.Register] 查询用户失败", zap.String("username", username), zap.Error(err))
		return nil, "", err
	}

	// 密码加密
	hash, err := s.passwordHasher.Hash(password)
	if err != nil {
		logger.Error("[user.Register] 密码加密失败", zap.Error(err))
		return nil, "", errors.NewWithMessage(errors.ServerError, "密码加密失败")
	}

	logger.Info("[user.Register] 密码加密成功",
		zap.String("username", username),
		zap.String("password_length", fmt.Sprintf("%d", len(password))))

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
		logger.Error("[user.Register] 创建用户失败", zap.String("username", username), zap.Error(err))
		return nil, "", errors.NewWithMessage(errors.ServerError, "创建用户失败")
	}
	logger.Info("[user.Register] 用户创建成功", zap.String("username", username), zap.Uint("user_id", user.ID))

	// 生成Token
	token, err := s.tokenGen.Generate(user.ID, user.Username)
	if err != nil {
		logger.Error("[user.Register] 生成token失败", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	logger.Info("[user.Register] 注册完成", zap.String("username", username), zap.Uint("user_id", user.ID))
	return user, token, nil
}

// Login 用户登录业务逻辑
// 业务规则：
// 1. 验证用户名和密码
// 2. 更新最后登录时间
// 3. 密码验证不使用缓存，确保安全性
func (s *Service) Login(username, password string) (*Entity, string, error) {
	logger.Info("[user.Login] 开始登录", zap.String("username", username))
	
	// 从数据库查询用户（不使用缓存，因为需要验证密码hash）
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		logger.Warn("[user.Login] 用户不存在", zap.String("username", username))
		return nil, "", errors.New(errors.InvalidPassword) // 不暴露用户是否存在
	}

	// 安全截取密码hash用于日志
	hashPrefix := user.Password
	if len(hashPrefix) > 20 {
		hashPrefix = hashPrefix[:20] + "..."
	}

	logger.Info("[user.Login] 验证密码",
		zap.String("username", username),
		zap.Uint("user_id", user.ID))

	// 验证密码
	if !s.passwordHasher.Check(password, user.Password) {
		logger.Warn("[user.Login] 密码验证失败",
			zap.String("username", username))
		return nil, "", errors.New(errors.InvalidPassword)
	}

	logger.Info("[user.Login] 密码验证成功", zap.String("username", username))

	// 更新最后登录时间
	s.repo.UpdateLastLogin(user.ID)
	logger.Info("[user.Login] 更新登录时间", zap.Uint("user_id", user.ID))

	// 生成Token
	token, err := s.tokenGen.Generate(user.ID, user.Username)
	if err != nil {
		logger.Error("[user.Login] 生成token失败", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	logger.Info("[user.Login] 登录成功", zap.String("username", username), zap.Uint("user_id", user.ID))
	return user, token, nil
}

// GetUsername 根据用户ID获取用户名
func (s *Service) GetUsername(userID uint) (string, error) {
	logger.Debug("[user.GetUsername] 获取用户名", zap.Uint("user_id", userID))
	user, err := s.repo.FindByID(userID)
	if err != nil {
		logger.Error("[user.GetUsername] 查询失败", zap.Uint("user_id", userID), zap.Error(err))
		return "", err
	}
	logger.Debug("[user.GetUsername] 成功", zap.Uint("user_id", userID), zap.String("username", user.Username))
	return user.Username, nil
}

// GetUsernames 批量获取用户名
func (s *Service) GetUsernames(userIDs []uint) (map[uint]string, error) {
	logger.Debug("[user.GetUsernames] 批量获取用户名", zap.Int("count", len(userIDs)))
	users, err := s.repo.FindByIDs(userIDs)
	if err != nil {
		logger.Error("[user.GetUsernames] 批量查询失败", zap.Int("count", len(userIDs)), zap.Error(err))
		return nil, err
	}

	usernameMap := make(map[uint]string, len(users))
	for _, user := range users {
		usernameMap[user.ID] = user.Username
	}
	logger.Debug("[user.GetUsernames] 成功", zap.Int("count", len(users)))
	return usernameMap, nil
}
