package routes

import (
	"faulty_in_culture/go_back/internal/handlers"
	"faulty_in_culture/go_back/internal/logger"
	"faulty_in_culture/go_back/internal/middleware"
	ws "faulty_in_culture/go_back/internal/websocket"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置所有路由
// 类型：Gin 路由注册函数
// 功能：注册所有 API 路由、Swagger 文档路由和健康检查路由，将 HTTP 路径与对应的 handler 绑定。
func SetupRoutes(router *gin.Engine, wsManager *ws.Manager) {
	logger.Info("routes.SetupRoutes: 开始设置路由")

	// 创建处理器实例
	// 创建排名业务处理器（Gin handler，业务逻辑层）
	rankingHandler := handlers.NewRankingHandler()
	saveGameHandler := handlers.NewSaveGameHandler()
	chatHandler := handlers.NewChatHandler(wsManager)

	// API路由组
	api := router.Group("/api")
	{
		// 用户相关路由（限流：60次/分钟）
		api.POST("/register", middleware.LimiterGlobal, handlers.Register)
		api.POST("/login", middleware.LimiterGlobal, handlers.Login)

		// 消息相关路由（统一限流）
		api.POST("/send-message", middleware.LimiterGlobal, handlers.SendMessage)
		api.GET("/query-result", middleware.LimiterGlobal, handlers.QueryResult)
		api.GET("/messages", middleware.LimiterGlobal, handlers.GetMessages)

		// 排名相关路由
		rankings := api.Group("/rankings")
		{
			rankings.GET("", middleware.LimiterGlobal, rankingHandler.GetRankings)
			rankings.GET("/top", middleware.LimiterGlobal, rankingHandler.GetTopRankings)
			rankings.GET("/:id", middleware.LimiterGlobal, rankingHandler.GetRanking)

			// 需要认证的接口（统一限流）
			rankingsAuth := rankings.Group("")
			rankingsAuth.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
			rankingsAuth.POST("", rankingHandler.CreateRanking)
			rankingsAuth.PUT("/:id", rankingHandler.UpdateRanking)
			rankingsAuth.DELETE("/:id", rankingHandler.DeleteRanking)
		}

		// 存档相关路由（需要认证，统一限流）
		savegames := api.Group("/savegames")
		savegames.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			savegames.GET("", saveGameHandler.GetSaveGames)
			savegames.GET("/:slot", saveGameHandler.GetSaveGame)
			savegames.PUT("/:slot", saveGameHandler.CreateOrUpdateSaveGame)
			savegames.DELETE("/:slot", saveGameHandler.DeleteSaveGame)
		}

		// AI聊天相关路由（需要认证，统一限流）
		chat := api.Group("/chat")
		chat.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			chat.POST("/start", chatHandler.StartChat)
			chat.POST("/send", chatHandler.SendMessage)
			chat.GET("/sessions", chatHandler.GetChatSessions)
			chat.GET("/:session_id", chatHandler.GetChatHistory)
		}
	}

	// WebSocket 路由
	router.GET("/ws", handlers.HandleWebSocket)

	// Swagger文档路由
	// 注册 Swagger 文档路由（第三方中间件，API 文档展示）
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	// 注册健康检查路由（Gin handler，基础运维功能）
	// HealthCheck 健康检查接口，返回服务状态
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Ranking API is running",
		})
	})

	logger.Info("routes.SetupRoutes: 路由设置完成",
		zap.Int("total_handlers", 4),
		zap.Strings("groups", []string{"user", "message", "ranking", "savegame", "chat"}),
	)
}
