package savegame

import "time"

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
	if slotNumber < 1 || slotNumber > 6 {
		return nil, ErrInvalidSlotNumber
	}
	return s.repo.FindByUserIDAndSlot(userID, slotNumber)
}

// QueryAll 查询用户的所有存档
func (s *Service) QueryAll(userID uint) ([]*Entity, error) {
	return s.repo.FindAllByUserID(userID)
}

// CreateOrUpdate 创建或更新存档
func (s *Service) CreateOrUpdate(userID uint, slotNumber int, gameData string) (*Entity, error) {
	if slotNumber < 1 || slotNumber > 6 {
		return nil, ErrInvalidSlotNumber
	}

	save := &Entity{
		UserID:     userID,
		SlotNumber: slotNumber,
		GameData:   gameData,
		SavedAt:    time.Now(),
	}

	err := s.repo.CreateOrUpdate(save)
	return save, err
}

// Delete 删除存档
func (s *Service) Delete(userID uint, slotNumber int) error {
	if slotNumber < 1 || slotNumber > 6 {
		return ErrInvalidSlotNumber
	}
	return s.repo.Delete(userID, slotNumber)
}
