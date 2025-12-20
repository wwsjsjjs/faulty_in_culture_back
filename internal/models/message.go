package models

import (
	"time"

	"gorm.io/gorm"
)

// Message 消息历史记录模型
type Message struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TaskID      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"task_id"` // 任务ID（唯一标识）
	UserID      string         `gorm:"type:varchar(100);index;not null" json:"user_id"`       // 用户ID
	Content     string         `gorm:"type:text;not null" json:"content"`                     // 消息内容
	Status      string         `gorm:"type:varchar(20);default:'pending'" json:"status"`      // 状态：pending, completed, failed
	ProcessedAt *time.Time     `gorm:"index" json:"processed_at"`                             // 消息处理完成时间（准备好的时间）
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`                      // 创建时间
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                      // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                     // 软删除
}

// TableName 自定义表名
func (Message) TableName() string {
	return "messages"
}
