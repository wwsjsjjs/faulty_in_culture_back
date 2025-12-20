package models

import (
	"time"

	"gorm.io/gorm"
)

// SaveGame 存档模型
type SaveGame struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index:idx_user_slot;not null" json:"user_id"`                                                 // 用户ID
	SlotNumber int            `gorm:"index:idx_user_slot;not null;check:slot_number >= 1 AND slot_number <= 6" json:"slot_number"` // 存档槽位（1-6）
	Data       string         `gorm:"type:longtext;not null" json:"data"`                                                          // 存档数据（JSON字符串）
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`                                                            // 创建时间
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                                                            // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                                                           // 软删除
}

// TableName 自定义表名
func (SaveGame) TableName() string {
	return "save_games"
}
