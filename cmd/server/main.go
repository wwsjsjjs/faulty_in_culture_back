package main

import (
	"fmt"
	"os"
	"time"

	"faulty_in_culture/go_back/internal/cache"
	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/handlers"
	"faulty_in_culture/go_back/internal/logger"
	"faulty_in_culture/go_back/internal/middleware"
	"faulty_in_culture/go_back/internal/queue"
	"faulty_in_culture/go_back/internal/routes"
	"faulty_in_culture/go_back/internal/scheduler"
	ws "faulty_in_culture/go_back/internal/websocket"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "faulty_in_culture/go_back/docs" // 导入swagger文档
)

// @title 排名系统API
// @version 1.0
// @description 这是一个用户排名管理系统的API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API支持
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	// 初始化日志系统（开发模式输出到终端）
	logMode := os.Getenv("LOG_MODE")
	if logMode == "" {
		logMode = "dev" // 默认开发模式
	}
	if err := logger.InitLogger(logMode); err != nil {
		logger.Error("main: 初始化 logger 失败", zap.Error(err))
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("main: 应用程序启动")

	// 加载配置
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		logger.Info("main: 加载配置文件", zap.String("path", configPath))
		config.LoadConfig(configPath)
	} else {
		logger.Error("main: 配置文件不存在", zap.String("path", configPath), zap.Error(err))
		os.Exit(1)
	}

	// 初始化数据库
	logger.Info("main: 初始化数据库")
	if err := database.InitDatabase(); err != nil {
		logger.Error("main: 数据库初始化失败", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("main: 数据库初始化成功")

	// 初始化 Redis 缓存
	logger.Info("main: 初始化 Redis 缓存")
	if err := cache.InitCache(); err != nil {
		logger.Warn("main: Redis 缓存初始化失败，服务将继续运行但无缓存功能", zap.Error(err))
	} else {
		logger.Info("main: Redis 缓存初始化成功")
	}

	// 初始化限流器
	redisConf := config.AppConfig.Redis
	redisAddr := fmt.Sprintf("%s:%d", redisConf.Host, redisConf.Port)
	logger.Info("main: 初始化限流器", zap.String("redisAddr", redisAddr))
	if err := middleware.InitLimiters(redisAddr, redisConf.Password, redisConf.DB); err != nil {
		logger.Error("main: 限流器初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// 初始化 WebSocket 管理器
	logger.Info("main: 初始化 WebSocket 管理器")
	wsManager := ws.NewManager()
	// 启动 WebSocket 心跳检测，10秒检测一次，30秒未活跃自动清理
	wsManager.StartHeartbeat(10*time.Second, 30*time.Second)
	handlers.InitMessageHandler(wsManager)
	logger.Info("main: WebSocket 管理器初始化成功")

	// 初始化消息队列（使用 Redis Streams）
	logger.Info("main: 初始化消息队列")
	if err := queue.InitQueue(redisAddr, redisConf.Password, redisConf.DB); err != nil {
		logger.Error("main: 消息队列初始化失败", zap.Error(err))
		os.Exit(1)
	}
	defer queue.Shutdown()
	logger.Info("main: 消息队列初始化成功")

	// 启动 Redis Streams worker
	logger.Info("main: 启动消息队列 worker")
	queue.StartWorker(handlers.ProcessDelayedMessage)

	// 启动定时清理任务
	logger.Info("main: 启动定时清理任务")
	scheduler.StartMessageCleanupScheduler()

	// 设置Gin模式（开发环境使用debug，生产环境使用release）
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	logger.Info("main: 设置 Gin 模式", zap.String("mode", mode))

	// 创建Gin路由器
	router := gin.Default()

	// 配置CORS（允许跨域请求）
	router.Use(corsMiddleware())

	// 设置路由（传递 wsManager）
	logger.Info("main: 设置路由")
	routes.SetupRoutes(router, wsManager)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	logger.Info("main: 启动服务器", zap.String("port", port))
	logger.Info("main: Swagger 文档地址", zap.String("swagger", fmt.Sprintf("http://localhost:%s/swagger/index.html", port)))
	logger.Info("main: 健康检查地址", zap.String("health", fmt.Sprintf("http://localhost:%s/health", port)))

	if err := router.Run(":" + port); err != nil {
		logger.Error("main: 服务器启动失败", zap.Error(err))
		os.Exit(1)
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
