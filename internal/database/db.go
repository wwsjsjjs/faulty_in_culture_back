package database

import (
	"fmt"
	"log"

	"github.com/yourusername/ranking-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	var err error

	// 使用SQLite数据库（生产环境可替换为MySQL/PostgreSQL）
	// SQLite文件会自动创建在项目根目录
	DB, err = gorm.Open(sqlite.Open("ranking.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开启SQL日志
	})

	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	log.Println("Database connection established")

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&models.Ranking{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")

	// 插入一些初始测试数据（可选）
	seedData()

	return nil
}

// seedData 插入初始测试数据
func seedData() {
	var count int64
	DB.Model(&models.Ranking{}).Count(&count)

	// 如果表中没有数据，插入测试数据
	if count == 0 {
		testData := []models.Ranking{
			{Username: "Alice", Score: 1500},
			{Username: "Bob", Score: 2000},
			{Username: "Charlie", Score: 1200},
			{Username: "David", Score: 1800},
			{Username: "Eve", Score: 2500},
		}

		for _, data := range testData {
			DB.Create(&data)
		}

		log.Println("Test data inserted successfully")
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
