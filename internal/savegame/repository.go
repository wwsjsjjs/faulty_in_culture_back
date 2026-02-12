package savegame

import (
	"gorm.io/gorm"
)

// Repository 存档仓储接口
type Repository interface {
	// Create 创建或更新存档（Upsert）
	CreateOrUpdate(save *Entity) error
	// FindByUserIDAndSlot 根据用户ID和槽位号查找
	FindByUserIDAndSlot(userID uint, slotNumber int) (*Entity, error)
	// FindAllByUserID 获取用户的所有存档
	FindAllByUserID(userID uint) ([]*Entity, error)
	// Delete 删除存档
	Delete(userID uint, slotNumber int) error
}

// repositoryImpl Repository的GORM实现
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository 创建存档仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// CreateOrUpdate 创建或更新存档（使用Upsert）
func (r *repositoryImpl) CreateOrUpdate(save *Entity) error {
	// GORM的Save方法：如果主键存在则更新，否则创建
	return r.db.Save(save).Error
}

// FindByUserIDAndSlot 根据用户ID和槽位号查找
func (r *repositoryImpl) FindByUserIDAndSlot(userID uint, slotNumber int) (*Entity, error) {
	var save Entity
	err := r.db.Where("user_id = ? AND slot_number = ?", userID, slotNumber).First(&save).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSaveGameNotFound
		}
		return nil, err
	}
	return &save, nil
}

// FindAllByUserID 获取用户的所有存档
func (r *repositoryImpl) FindAllByUserID(userID uint) ([]*Entity, error) {
	var saves []*Entity
	err := r.db.Where("user_id = ?", userID).
		Order("slot_number ASC").
		Find(&saves).Error
	return saves, err
}

// Delete 删除存档（软删除）
func (r *repositoryImpl) Delete(userID uint, slotNumber int) error {
	result := r.db.Where("user_id = ? AND slot_number = ?", userID, slotNumber).Delete(&Entity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSaveGameNotFound
	}
	return nil
}
