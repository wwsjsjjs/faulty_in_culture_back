// Package ranking - 排行榜模块数据访问层
// 功能：封装排行榜数据的CRUD操作
// 设计模式：Repository模式
package ranking

import (
	"fmt"

	"gorm.io/gorm"
)

// ============================================================
// Repository模式 (Repository Pattern)
// 设计模式：仓储模式 - 封装数据访问逻辑
// ============================================================

// Repository 排行榜仓储接口
type Repository interface {
	// UpsertScore 创建或更新分数（只在新分数更高时更新）
	UpsertScore(userID uint, rankType, score int) (*Entity, error)
	// GetRankings 获取排行榜（分页）
	GetRankings(rankType, offset, limit int) ([]*Entity, error)
	// DeleteByUserAndType 删除指定用户和类型的排行榜记录
	DeleteByUserAndType(userID uint, rankType int) error
	// DeleteAllByUser 删除指定用户的所有排行榜记录
	DeleteAllByUser(userID uint) error
	// FindByUserAndType 查找指定用户和类型的排行榜记录
	FindByUserAndType(userID uint, rankType int) (*Entity, error)
}

// repositoryImpl Repository的GORM实现
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository 创建排行榜仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// UpsertScore 创建或更新分数（只在新分数更高时更新）
func (r *repositoryImpl) UpsertScore(userID uint, rankType, score int) (*Entity, error) {
	var ranking Entity

	// 先查询是否存在
	err := r.db.Where("user_id = ? AND rank_type = ?", userID, rankType).First(&ranking).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		ranking = Entity{
			UserID:   userID,
			RankType: rankType,
			Score:    score,
		}
		if err := r.db.Create(&ranking).Error; err != nil {
			return nil, err
		}
		return &ranking, nil
	}

	if err != nil {
		return nil, err
	}

	// 存在，只在新分数更高时更新
	if score > ranking.Score {
		ranking.Score = score
		if err := r.db.Save(&ranking).Error; err != nil {
			return nil, err
		}
	}

	return &ranking, nil
}

// GetRankings 获取排行榜（按分数降序，分数相同按更新时间升序）
func (r *repositoryImpl) GetRankings(rankType, offset, limit int) ([]*Entity, error) {
	var rankings []*Entity
	err := r.db.Where("rank_type = ?", rankType).
		Order("score DESC, updated_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rankings).Error

	return rankings, err
}

// DeleteByUserAndType 删除指定用户和类型的排行榜记录
func (r *repositoryImpl) DeleteByUserAndType(userID uint, rankType int) error {
	return r.db.Where("user_id = ? AND rank_type = ?", userID, rankType).Delete(&Entity{}).Error
}

// DeleteAllByUser 删除指定用户的所有排行榜记录
func (r *repositoryImpl) DeleteAllByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&Entity{}).Error
}

// FindByUserAndType 查找指定用户和类型的排行榜记录
func (r *repositoryImpl) FindByUserAndType(userID uint, rankType int) (*Entity, error) {
	var ranking Entity
	err := r.db.Where("user_id = ? AND rank_type = ?", userID, rankType).First(&ranking).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("排行榜记录不存在")
		}
		return nil, err
	}
	return &ranking, nil
}
