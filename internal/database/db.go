package database

import (
	"faulty_in_culture/go_back/internal/logger"
	"fmt"

	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	logger.Info("Database connection established (MySQL)")

	err = DB.AutoMigrate(
		&models.User{},
		&models.SaveGame{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.Message{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	logger.Info("Database migration completed")

	return nil
}

// GetDB 获取数据库实例
// 类型：Getter
// 功能：返回全局数据库连接对象，供业务逻辑调用
// GetDB 获取数据库实例（返回全局数据库连接对象，供业务逻辑调用）
func GetDB() *gorm.DB {
	return DB
}
