package routes

import (
	"faulty_in_culture/go_back/internal/chat"
	"faulty_in_culture/go_back/internal/savegame"
	"faulty_in_culture/go_back/internal/shared/infra/logger"
	"faulty_in_culture/go_back/internal/shared/middleware"
	"faulty_in_culture/go_back/internal/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Handlers 路由处理器集合（依赖注入）
type Handlers struct {
	User     *user.Handler
	Chat     *chat.Handler
	SaveGame *savegame.Handler
}

// SetupRoutes 设置所有路由（简化MVC架构）
func SetupRoutes(router *gin.Engine, h *Handlers) {
	logger.Info("开始设置路由（简化MVC架构）")

	api := router.Group("/api")
	{
		// 用户模块（公开接口）
		api.POST("/register", middleware.LimiterGlobal, h.User.Register)
		api.POST("/login", middleware.LimiterGlobal, h.User.Login)

		// 排行榜（公开接口）
		api.GET("/rankings/:rank_type", middleware.LimiterGlobal, h.User.GetRankings)

		// 用户模块（需要认证）
		userGroup := api.Group("/user")
		userGroup.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			userGroup.PUT("/score", h.User.UpdateScore)
		}

		// 聊天模块（需要认证）
		chatGroup := api.Group("/chat")
		chatGroup.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			chatGroup.POST("/start", h.Chat.StartChat)
			chatGroup.POST("/send", h.Chat.SendMessage)
			chatGroup.GET("/sessions", h.Chat.ListSessions)
			chatGroup.GET("/history", h.Chat.GetHistory)
			chatGroup.DELETE("/recall", h.Chat.RecallMessages)
		}

		// 存档模块（需要认证）
		saveGameGroup := api.Group("/savegame")
		saveGameGroup.Use(middleware.LimiterGlobal, middleware.AuthMiddleware())
		{
			saveGameGroup.GET("", h.SaveGame.QueryBySlot) // ?slot_number=1
			saveGameGroup.GET("/all", h.SaveGame.QueryAll)
			saveGameGroup.POST("", h.SaveGame.CreateOrUpdate)
			saveGameGroup.DELETE("", h.SaveGame.Delete) // ?slot_number=1
		}
	}

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "简化MVC架构运行中"})
	})

	logger.Info("路由设置完成（简化MVC）",
		zap.Strings("modules", []string{"user", "chat", "savegame"}),
	)
}
