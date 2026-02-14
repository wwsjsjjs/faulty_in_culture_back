// Package savegame - 存档模块
// 功能：管理用户的游戏存档数据
// 特点：每个用户支持6个存档槽位
package savegame

import (
	"time"
)

// Entity 存档实体
type Entity struct {
	UserID     uint      `gorm:"primaryKey;not null" json:"user_id"`
	SlotNumber int       `gorm:"primaryKey;not null;check:slot_number >= 1 AND slot_number <= 6" json:"slot_number"`
	GameData   string    `gorm:"type:text" json:"game_data"`
	SavedAt    time.Time `gorm:"autoCreateTime" json:"saved_at"`
}

func (Entity) TableName() string {
	return "save_games"
}
