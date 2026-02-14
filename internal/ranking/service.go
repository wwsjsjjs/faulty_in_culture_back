// Package ranking - 排行榜模块业务逻辑层
// 功能：实现排行榜相关的业务规则
// 特点：自动保留最高分，支持缓存
package ranking

import (
	"fmt"
	"time"
)

// ============================================================
// Service层（业务逻辑层）
// 设计模式：服务层模式 + Service组合调用
// 架构：ranking依赖user的Service接口（单向依赖）
// ============================================================

// UserService 用户服务接口（Service层组合调用）
// 设计模式：依赖倒置 - ranking依赖user的抽象接口，而非具体实现
// 优点：
// 1. 依赖单向（ranking → user，user不感知ranking）
// 2. 符合单一职责，每个Service只负责自己模块的业务逻辑
// 3. 易于单元测试（可Mock UserService）
type UserService interface {
	GetUsername(userID uint) (string, error)
	GetUsernames(userIDs []uint) (map[uint]string, error)
}

// Cache 缓存接口
type Cache interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

// Service 排行榜业务服务
type Service struct {
	repo        Repository
	userService UserService
	cache       Cache
}

// NewService 创建排行榜服务实例（依赖注入）
func NewService(repo Repository, userService UserService, cache Cache) *Service {
	return &Service{
		repo:        repo,
		userService: userService,
		cache:       cache,
	}
}

// UpdateScore 更新用户分数（只在新分数更高时更新）
func (s *Service) UpdateScore(userID uint, rankType, score int) (*Entity, error) {
	if !ValidateRankType(rankType) {
		return nil, fmt.Errorf("排行榜类型无效")
	}
	if score < 0 {
		return nil, fmt.Errorf("分数无效")
	}

	ranking, err := s.repo.UpsertScore(userID, rankType, score)
	if err != nil {
		return nil, fmt.Errorf("更新分数失败")
	}

	if s.cache != nil {
		s.clearRankingCache(rankType)
	}
	return ranking, nil
}

// GetRankings 获取排行榜
func (s *Service) GetRankings(rankType, page, limit int) ([]RankingItem, error) {
	if !ValidateRankType(rankType) {
		return nil, fmt.Errorf("排行榜类型无效")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	cacheKey := fmt.Sprintf("rankings:type:%d:page:%d:limit:%d", rankType, page, limit)

	// 尝试从缓存获取
	if s.cache != nil {
		var cachedRankings []RankingItem
		if err := s.cache.Get(cacheKey, &cachedRankings); err == nil && len(cachedRankings) > 0 {
			return cachedRankings, nil
		}
	}

	// 从数据库查询
	rankings, err := s.repo.GetRankings(rankType, offset, limit)
	if err != nil {
		return nil, err
	}

	// 收集所有 userID 并批量查询用户名（优化N+1查询）
	userIDs := make([]uint, len(rankings))
	for i, r := range rankings {
		userIDs[i] = r.UserID
	}

	// 批量获取用户名
	usernames := make(map[uint]string)
	if s.userService != nil && len(userIDs) > 0 {
		if usernameMap, err := s.userService.GetUsernames(userIDs); err == nil {
			usernames = usernameMap
		}
	}

	// 转换为VO并计算排名
	items := make([]RankingItem, len(rankings))
	for i, r := range rankings {
		// 从批量查询结果中获取用户名
		username := fmt.Sprintf("user_%d", r.UserID)
		if name, ok := usernames[r.UserID]; ok && name != "" {
			username = name
		}

		items[i] = RankingItem{
			Rank:      offset + i + 1, // 排名 = 偏移量 + 当前索引 + 1
			UserID:    r.UserID,
			Username:  username,
			Score:     r.Score,
			UpdatedAt: r.UpdatedAt,
		}
	}

	// 写入缓存（缓存10分钟）
	if s.cache != nil {
		s.cache.Set(cacheKey, items, 10*time.Minute)
	}

	return items, nil
}

// DeleteRanking 删除指定类型的排行榜记录
func (s *Service) DeleteRanking(userID uint, rankType int) error {
	if !ValidateRankType(rankType) {
		return fmt.Errorf("排行榜类型无效")
	}

	if err := s.repo.DeleteByUserAndType(userID, rankType); err != nil {
		return err
	}

	if s.cache != nil {
		s.clearRankingCache(rankType)
	}
	return nil
}

// DeleteAllRankings 删除用户的所有排行榜记录
func (s *Service) DeleteAllRankings(userID uint) error {
	if err := s.repo.DeleteAllByUser(userID); err != nil {
		return err
	}

	// 清除所有排行榜缓存
	if s.cache != nil {
		for i := 1; i <= 9; i++ {
			s.clearRankingCache(i)
		}
	}

	return nil
}

// clearRankingCache 清除排行榜缓存
func (s *Service) clearRankingCache(rankType int) {
	for page := 1; page <= 10; page++ {
		for limit := 10; limit <= 100; limit += 10 {
			key := fmt.Sprintf("rankings:type:%d:page:%d:limit:%d", rankType, page, limit)
			s.cache.Delete(key)
		}
	}
}
