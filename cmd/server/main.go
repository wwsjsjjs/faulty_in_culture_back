package main

import (
	"fmt"
	"os"
	"time"

	"faulty_in_culture/go_back/internal/chat"
	"faulty_in_culture/go_back/internal/routes"
	"faulty_in_culture/go_back/internal/savegame"
	"faulty_in_culture/go_back/internal/shared/infra/cache"
	"faulty_in_culture/go_back/internal/shared/infra/config"
	"faulty_in_culture/go_back/internal/shared/infra/db"
	"faulty_in_culture/go_back/internal/shared/infra/logger"
	"faulty_in_culture/go_back/internal/shared/infra/ws"
	"faulty_in_culture/go_back/internal/shared/middleware"
	"faulty_in_culture/go_back/internal/user"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	_ "faulty_in_culture/go_back/docs"
)

// @title faulty_in_culture API
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	// 初始化日志
	logMode := getEnv("LOG_MODE", "dev")
	if err := logger.InitLogger(logMode); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("应用启动（简化MVC架构）")

	// 加载配置
	config.LoadConfig("config.yaml")

	// 初始化共享基础设施（会自动检查连接）
	if err := db.InitDatabase(); err != nil {
		logger.Error("数据库初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// 获取配置
	cfg := config.AppConfig
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	if err := cache.InitCache(); err != nil {
		logger.Warn("Redis缓存初始化失败", zap.Error(err))
	}

	if err := middleware.InitLimiters(redisAddr, cfg.Redis.Password, cfg.Redis.DB); err != nil {
		logger.Error("限流器初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// WebSocket管理器
	wsManager := ws.NewManager()
	wsManager.StartHeartbeat(10*time.Second, 30*time.Second)

	// ============================================================
	// 依赖注入 - 构建各领域模块（简化MVC）
	// 架构层次：Entity -> Repository -> Service -> Handler
	// ============================================================

	// 获取数据库连接
	database := db.GetDB()

	// User模块（完整依赖注入示例）
	_ = user.NewRepository(database) // TODO: 实现完整依赖注入
	// userService := user.NewService(userRepo, passwordHasher, tokenGen, cacheImpl)
	// userHandler := user.NewHandler(userService)

	// Chat模块
	_ = chat.NewRepository(database) // TODO: 实现完整依赖注入
	// chatService := chat.NewService(chatRepo, aiClient, wsManager, cacheImpl)
	// chatHandler := chat.NewHandler(chatService)

	// SaveGame模块
	saveGameRepo := savegame.NewRepository(database)
	saveGameService := savegame.NewService(saveGameRepo)
	saveGameHandler := savegame.NewHandler(saveGameService)

	// 临时：使用占位符（待实现完整的依赖注入）
	handlers := &routes.Handlers{
		// User:     userHandler,
		// Chat:     chatHandler,
		SaveGame: saveGameHandler,
	}

	// 启动HTTP服务
	gin.SetMode(getEnv("GIN_MODE", gin.DebugMode))
	router := gin.Default()
	router.Use(corsMiddleware())

	// 设置路由（传入handlers）
	routes.SetupRoutes(router, handlers)

	port := getEnv("PORT", "8080")
	logger.Info("服务启动（简化MVC）",
		zap.String("port", port),
		zap.String("swagger", fmt.Sprintf("http://localhost:%s/swagger/index.html", port)),
		zap.Strings("modules", []string{"user", "chat", "savegame"}))

	if err := router.Run(":" + port); err != nil {
		logger.Error("服务启动失败", zap.Error(err))
		os.Exit(1)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
