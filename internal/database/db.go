package database

import (
	"fmt"
	"log"

	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
// 类型：*gorm.DB（GORM 框架数据库连接对象）
var DB *gorm.DB

// InitDatabase 初始化数据库连接
// 类型：初始化函数
// 功能：从配置文件读取数据库参数，连接 MySQL，并自动迁移表结构和插入测试数据
func InitDatabase() error {
	var err error

	// 从配置读取数据库信息
	dbConf := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name, dbConf.Charset)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	log.Println("Database connection established (MySQL)")

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(
		&models.Ranking{},
		&models.User{},
		&models.Message{},
		&models.SaveGame{},
		&models.ChatSession{},
		&models.ChatMessage{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")

	// 插入一些初始测试数据（可选）
	seedData()

	return nil
}

// seedData 插入初始测试数据
// 类型：私有辅助函数
// 功能：如表为空则插入默认测试数据，便于开发和演示
// seedData 插入初始测试数据（如表为空则插入默认测试数据，便于开发和演示）
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
// 类型：Getter
// 功能：返回全局数据库连接对象，供业务逻辑调用
// GetDB 获取数据库实例（返回全局数据库连接对象，供业务逻辑调用）
func GetDB() *gorm.DB {
	return DB
}
