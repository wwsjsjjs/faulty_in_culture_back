package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ranking-api/internal/handlers"
	ws "github.com/yourusername/ranking-api/internal/websocket"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置所有路由
// 类型：Gin 路由注册函数
// 功能：注册所有 API 路由、Swagger 文档路由和健康检查路由，将 HTTP 路径与对应的 handler 绑定。
func SetupRoutes(router *gin.Engine, wsManager *ws.Manager) {
	// 创建处理器实例
	// 创建排名业务处理器（Gin handler，业务逻辑层）
	rankingHandler := handlers.NewRankingHandler()
	saveGameHandler := handlers.NewSaveGameHandler()
	chatHandler := handlers.NewChatHandler(wsManager)

	// API路由组
	api := router.Group("/api")
	{
		// 用户相关路由（限流：5次/分钟）
		api.POST("/register", handlers.RateLimitMiddleware("register", 5, time.Minute), handlers.Register)
		api.POST("/login", handlers.RateLimitMiddleware("login", 10, time.Minute), handlers.Login)

		// 消息相关路由
		api.POST("/send-message", handlers.SendMessage) // 发送延迟消息
		api.GET("/query-result", handlers.QueryResult)  // 查询消息结果
		api.GET("/messages", handlers.GetMessages)      // 获取历史消息列表

		// 排名相关路由
		rankings := api.Group("/rankings")
		{
			rankings.GET("", rankingHandler.GetRankings)        // 获取所有排名（分页）
			rankings.GET("/top", rankingHandler.GetTopRankings) // 获取前N名
			rankings.GET("/:id", rankingHandler.GetRanking)     // 获取单个排名

			// 需要认证的接口（限流：20次/分钟）
			rankingsAuth := rankings.Group("")
			rankingsAuth.Use(handlers.AuthMiddleware())
			rankingsAuth.POST("", rankingHandler.CreateRanking)                                                                            // 创建排名
			rankingsAuth.PUT("/:id", handlers.PerUserRateLimitMiddleware("ranking_update", 20, time.Minute), rankingHandler.UpdateRanking) // 更新排名
			rankingsAuth.DELETE("/:id", rankingHandler.DeleteRanking)                                                                      // 删除排名
		}

		// 存档相关路由（需要认证，限流：30次/分钟）
		savegames := api.Group("/savegames")
		savegames.Use(handlers.AuthMiddleware())
		{
			savegames.GET("", saveGameHandler.GetSaveGames)                                                                                         // 获取所有存档
			savegames.GET("/:slot", saveGameHandler.GetSaveGame)                                                                                    // 获取指定槽位存档
			savegames.PUT("/:slot", handlers.PerUserRateLimitMiddleware("savegame_write", 30, time.Minute), saveGameHandler.CreateOrUpdateSaveGame) // 创建或更新存档
			savegames.DELETE("/:slot", handlers.PerUserRateLimitMiddleware("savegame_delete", 10, time.Minute), saveGameHandler.DeleteSaveGame)     // 删除存档
		}

		// AI聊天相关路由（需要认证）
		chat := api.Group("/chat")
		chat.Use(handlers.AuthMiddleware())
		{
			chat.POST("/start", chatHandler.StartChat)           // 开始新对话
			chat.POST("/send", chatHandler.SendMessage)          // 发送消息
			chat.GET("/sessions", chatHandler.GetChatSessions)   // 获取所有会话
			chat.GET("/:session_id", chatHandler.GetChatHistory) // 获取聊天历史
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
}
