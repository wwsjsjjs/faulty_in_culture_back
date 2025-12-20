package models

// 本文件定义了排名相关的数据结构体，所有结构体均带有 Swagger 注释和字段示例，便于自动生成 API 文档。

import (
	"time"

	"gorm.io/gorm"
)

// Ranking 排名表模型
// 类型：GORM 数据模型结构体
// 功能：描述排名表结构，映射数据库表 rankings
// @Description 用户排名信息
// @Tags Ranking
type Ranking struct {
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`                                         // 主键ID
	Username  string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username" example:"player1"` // 用户名（唯一）
	Score     int            `gorm:"not null;default:0;index" json:"score" example:"1000"`                     // 分数
	CreatedAt time.Time      `json:"created_at" example:"2023-12-20T10:00:00Z" format:"date-time"`             // 创建时间
	UpdatedAt time.Time      `json:"updated_at" example:"2023-12-20T10:00:00Z" format:"date-time"`             // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                                           // 软删除
}

// TableName 指定表名
// 类型：GORM 接口实现
// 功能：指定该模型对应的数据库表名为 rankings
// TableName 指定 Ranking 结构体对应的数据库表名
func (Ranking) TableName() string {
	return "rankings"
}
