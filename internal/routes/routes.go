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

	saveGameHandler := handlers.NewSaveGameHandler()
	chatHandler := handlers.NewChatHandler(wsManager)

	api := router.Group("/api")
	{
		api.GET("/config", handlers.GetPublicConfig)

		api.POST("/register", middleware.LimiterGlobal, handlers.Register)
		api.POST("/login", middleware.LimiterGlobal, handlers.Login)

		api.POST("/send-message", middleware.LimiterGlobal, handlers.SendMessage)
		api.GET("/query-result", middleware.LimiterGlobal, handlers.QueryResult)
		api.GET("/messages", middleware.LimiterGlobal, handlers.GetMessages)

		rankings := api.Group("/rankings")
		{
			rankings.GET("", middleware.LimiterGlobal, handlers.GetRankings)
		}

		user := api.Group("/user")
		user.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			user.PUT("/score", handlers.UpdateUserScore)
		}

		savegames := api.Group("/savegames")
		savegames.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			savegames.GET("", saveGameHandler.GetSaveGames)
			savegames.GET("/:slot", saveGameHandler.GetSaveGame)
			savegames.PUT("/:slot", saveGameHandler.CreateOrUpdateSaveGame)
			savegames.DELETE("/:slot", saveGameHandler.DeleteSaveGame)
		}

		chat := api.Group("/chat")
		chat.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			chat.POST("/start", chatHandler.StartChat)
			chat.POST("/send", chatHandler.SendMessage)
			chat.GET("/sessions", chatHandler.GetChatSessions)
			chat.GET("/:session_id", chatHandler.GetChatHistory)
		}
	}

	router.GET("/ws", handlers.HandleWebSocket)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Ranking API is running",
		})
	})

	logger.Info("routes.SetupRoutes: 路由设置完成",
		zap.Int("total_handlers", 3),
		zap.Strings("groups", []string{"user", "message", "savegame", "chat"}),
	)
}
