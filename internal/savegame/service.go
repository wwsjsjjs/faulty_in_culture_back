// Package savegame - 存档模块业务逻辑层
// 功能：实现存档的业务规则
// 特点：支持6个槽位，创建/更新/删除/查询
package savegame

import (
	"fmt"
	"time"

	"faulty_in_culture/go_back/internal/infra/logger"

	"go.uber.org/zap"
)

// ============================================================
// Service层 - 业务逻辑层
// 职责：
// 1. 实现存档的CRUD业务逻辑
// 2. 验证槽位号有效性
// 3. 权限验证
// ============================================================

// Service 存档服务
type Service struct {
	repo Repository
}

// NewService 创建存档服务实例
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// QueryBySlot 查询指定槽位的存档
func (s *Service) QueryBySlot(userID uint, slotNumber int) (*Entity, error) {
	logger.Debug("[savegame.QueryBySlot] 查询存档",
		zap.Uint("user_id", userID),
		zap.Int("slot_number", slotNumber))

	if slotNumber < 1 || slotNumber > 6 {
		logger.Warn("[savegame.QueryBySlot] 槽位号无效",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber))
		return nil, fmt.Errorf("槽位号无效")
	}

	save, err := s.repo.FindByUserIDAndSlot(userID, slotNumber)
	if err != nil {
		logger.Debug("[savegame.QueryBySlot] 查询失败",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber),
			zap.Error(err))
	}
	return save, err
}

// QueryAll 查询用户的所有存档
func (s *Service) QueryAll(userID uint) ([]*Entity, error) {
	logger.Debug("[savegame.QueryAll] 查询所有存档", zap.Uint("user_id", userID))

	saves, err := s.repo.FindAllByUserID(userID)
	if err != nil {
		logger.Error("[savegame.QueryAll] 查询失败", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}

	logger.Debug("[savegame.QueryAll] 成功", zap.Uint("user_id", userID), zap.Int("count", len(saves)))
	return saves, nil
}

// CreateOrUpdate 创建或更新存档
func (s *Service) CreateOrUpdate(userID uint, slotNumber int, gameData string) (*Entity, error) {
	logger.Info("[savegame.CreateOrUpdate] 创建或更新存档",
		zap.Uint("user_id", userID),
		zap.Int("slot_number", slotNumber),
		zap.Int("data_size", len(gameData)))

	if slotNumber < 1 || slotNumber > 6 {
		logger.Warn("[savegame.CreateOrUpdate] 槽位号无效",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber))
		return nil, fmt.Errorf("槽位号无效")
	}

	save := &Entity{
		UserID:     userID,
		SlotNumber: slotNumber,
		GameData:   gameData,
		SavedAt:    time.Now(),
	}

	err := s.repo.CreateOrUpdate(save)
	if err != nil {
		logger.Error("[savegame.CreateOrUpdate] 保存失败",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber),
			zap.Error(err))
		return nil, err
	}

	logger.Info("[savegame.CreateOrUpdate] 保存成功",
		zap.Uint("user_id", userID),
		zap.Int("slot_number", slotNumber),
		zap.Int("data_size", len(gameData)))
	return save, nil
}

// Delete 删除存档
func (s *Service) Delete(userID uint, slotNumber int) error {
	logger.Info("[savegame.Delete] 删除存档",
		zap.Uint("user_id", userID),
		zap.Int("slot_number", slotNumber))

	if slotNumber < 1 || slotNumber > 6 {
		logger.Warn("[savegame.Delete] 槽位号无效",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber))
		return fmt.Errorf("槽位号无效")
	}

	err := s.repo.Delete(userID, slotNumber)
	if err != nil {
		logger.Error("[savegame.Delete] 删除失败",
			zap.Uint("user_id", userID),
			zap.Int("slot_number", slotNumber),
			zap.Error(err))
		return err
	}

	logger.Info("[savegame.Delete] 删除成功",
		zap.Uint("user_id", userID),
		zap.Int("slot_number", slotNumber))
	return nil
}
