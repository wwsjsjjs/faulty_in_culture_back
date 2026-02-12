package savegame

import (
	"time"

	"gorm.io/gorm"
)

// Entity 存档实体
type Entity struct {
	UserID     uint           `gorm:"primaryKey;not null" json:"user_id"`
	SlotNumber int            `gorm:"primaryKey;not null;check:slot_number >= 1 AND slot_number <= 6" json:"slot_number"`
	GameData   string         `gorm:"type:text" json:"game_data"`
	SavedAt    time.Time      `gorm:"autoCreateTime" json:"saved_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Entity) TableName() string {
	return "save_games"
}

// IsValid 验证槽位号是否有效
func (e *Entity) IsValid() bool {
	return e.SlotNumber >= 1 && e.SlotNumber <= 6
}
