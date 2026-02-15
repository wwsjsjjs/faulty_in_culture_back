package db

import (
	"faulty_in_culture/go_back/internal/chat"
	"faulty_in_culture/go_back/internal/infra/config"
	"faulty_in_culture/go_back/internal/infra/logger"
	"faulty_in_culture/go_back/internal/ranking"
	"faulty_in_culture/go_back/internal/savegame"
	"faulty_in_culture/go_back/internal/user"
	"fmt"

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
	dbConf := config.GlobalConfig.Database

	// 如果启用自动创建，先创建数据库
	if dbConf.AutoCreate {
		createDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
			dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Charset)
		tempDB, err := gorm.Open(mysql.Open(createDSN), &gorm.Config{})
		if err == nil {
			tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", dbConf.Name, dbConf.Charset))
			logger.Info("Database auto-create executed")
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name, dbConf.Charset)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	logger.Info("Database connection established (MySQL)")

	// 自动迁移（简化MVC架构 - 使用各domain的Entity）
	err = DB.AutoMigrate(
		&user.Entity{},
		&ranking.Entity{},
		&savegame.Entity{},
		&chat.Session{},
		&chat.Message{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	logger.Info("数据库表结构迁移完成")

	return nil
}

// GetDB 获取数据库实例
// 类型：Getter
// 功能：返回全局数据库连接对象，供业务逻辑调用
// GetDB 获取数据库实例（返回全局数据库连接对象，供业务逻辑调用）
func GetDB() *gorm.DB {
	return DB
}
