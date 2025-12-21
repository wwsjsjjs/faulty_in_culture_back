package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"faulty_in_culture/go_back/internal/cache"
	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/handlers"
	"faulty_in_culture/go_back/internal/queue"
	"faulty_in_culture/go_back/internal/routes"
	"faulty_in_culture/go_back/internal/scheduler"
	ws "faulty_in_culture/go_back/internal/websocket"

	"github.com/gin-gonic/gin"

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
	// 加载配置
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		// 存在 config.yaml
		importConfig := "faulty_in_culture/go_back/internal/config"
		_ = importConfig // 避免未使用错误
		config.LoadConfig(configPath)
	} else {
		log.Fatalf("Config file not found: %v", err)
	}

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Redis 缓存
	if err := cache.InitCache(); err != nil {
		log.Printf("Warning: Failed to initialize cache: %v", err)
		log.Println("Server will continue without caching")
	} else {
		log.Println("Redis cache initialized successfully")
	}

	// 初始化 WebSocket 管理器
	wsManager := ws.NewManager()
	// 启动 WebSocket 心跳检测，10秒检测一次，30秒未活跃自动清理
	wsManager.StartHeartbeat(10*time.Second, 30*time.Second)
	handlers.InitMessageHandler(wsManager)

	// 初始化消息队列（使用 Redis Streams）
	redisConf := config.AppConfig.Redis
	redisAddr := fmt.Sprintf("%s:%d", redisConf.Host, redisConf.Port)
	if err := queue.InitQueue(redisAddr, redisConf.Password, redisConf.DB); err != nil {
		log.Fatalf("Failed to initialize message queue: %v", err)
	}
	defer queue.Shutdown()

	// 启动 Redis Streams worker
	queue.StartWorker(handlers.ProcessDelayedMessage)

	// 启动定时清理任务
	scheduler.StartMessageCleanupScheduler()

	// 设置Gin模式（开发环境使用debug，生产环境使用release）
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	// 创建Gin路由器
	router := gin.Default()

	// 配置CORS（允许跨域请求）
	router.Use(corsMiddleware())

	// 设置路由（传递 wsManager）
	routes.SetupRoutes(router, wsManager)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	log.Printf("Server starting on http://localhost:%s", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)
	log.Printf("Health check available at http://localhost:%s/health", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
