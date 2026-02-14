// Package main - 游戏后端API服务主入口
// 功能：用户认证、排行榜管理、AI聊天、存档管理
// 架构：RESTful API + MVC分层架构
package main

import (
	"fmt"
	"os"

	"faulty_in_culture/go_back/internal/chat"
	"faulty_in_culture/go_back/internal/infra/cache"
	"faulty_in_culture/go_back/internal/infra/config"
	"faulty_in_culture/go_back/internal/infra/db"
	"faulty_in_culture/go_back/internal/infra/logger"
	"faulty_in_culture/go_back/internal/ranking"
	"faulty_in_culture/go_back/internal/routes"
	"faulty_in_culture/go_back/internal/savegame"
	"faulty_in_culture/go_back/internal/user"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	_ "faulty_in_culture/go_back/docs"
)

// @title Faulty In Culture API
// @version 1.0
// @description 游戏后端API服务 - 提供用户认证、排行榜、AI聊天、存档管理功能
// @host localhost:8080
// @BasePath /
func main() {
	// 加载配置文件
	config.LoadConfig("config.yaml")
	cfg := &config.GlobalConfig

	// 初始化日志系统
	if err := logger.InitLogger(cfg.App.LogMode); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("应用启动（RESTful架构）", zap.String("environment", cfg.App.Environment))

	// 初始化数据库连接
	if err := db.InitDatabase(); err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// 初始化Redis缓存
	if err := cache.InitCache(); err != nil {
		logger.Error("Redis缓存初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// ============================================================
	// 依赖注入 - 构建各领域模块
	// 架构层次：Entity -> Repository -> Service -> Handler
	// ============================================================

	database := db.GetDB()
	cacheInstance := cache.GetCache()

	// User模块 - 用户认证管理
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo, user.NewPasswordHasher(), user.NewTokenGenerator(), cacheInstance)
	userHandler := user.NewHandler(userService)

	// Ranking模块 - 排行榜管理（依赖UserService，使用批量查询优化）
	rankingRepo := ranking.NewRepository(database)
	rankingService := ranking.NewService(rankingRepo, userService, cacheInstance)
	rankingHandler := ranking.NewHandler(rankingService)

	// Chat模块 - AI聊天管理
	chatRepo := chat.NewRepository(database)
	chatService := chat.NewService(chatRepo, chat.NewAIClient(), nil, cacheInstance)
	chatHandler := chat.NewHandler(chatService)

	// SaveGame模块 - 存档管理
	saveGameRepo := savegame.NewRepository(database)
	saveGameService := savegame.NewService(saveGameRepo)
	saveGameHandler := savegame.NewHandler(saveGameService)

	// 组装所有处理器
	handlers := &routes.Handlers{
		User:     userHandler,
		Chat:     chatHandler,
		SaveGame: saveGameHandler,
		Ranking:  rankingHandler,
	}

	// 启动HTTP服务器
	gin.SetMode(cfg.App.GinMode)
	router := gin.Default()

	// 设置路由
	routes.SetupRoutes(router, handlers)

	logger.Info("服务启动成功",
		zap.String("port", cfg.App.Port),
		zap.String("swagger", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.App.Port)),
		zap.Strings("modules", []string{"user", "ranking", "chat", "savegame"}))

	if err := router.Run(":" + cfg.App.Port); err != nil {
		logger.Error("服务启动失败", zap.Error(err))
		os.Exit(1)
	}
}
