package models

import (
	"time"

	"gorm.io/gorm"
)

type SaveGame struct {
	UserID     uint           `gorm:"primaryKey;not null" json:"user_id"`
	SlotNumber int            `gorm:"primaryKey;not null;check:slot_number >= 1 AND slot_number <= 6" json:"slot_number"`
	User       User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	GameData   string         `gorm:"type:text" json:"game_data"`
	SavedAt    time.Time      `gorm:"autoCreateTime" json:"saved_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 自定义表名
func (SaveGame) TableName() string {
	return "save_games"
}
