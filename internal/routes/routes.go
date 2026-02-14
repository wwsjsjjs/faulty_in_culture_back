// Package routes - 路由配置模块
// 功能：配置所有HTTP路由和中间件
// 架构：RESTful风格API设计
package routes

import (
	"faulty_in_culture/go_back/internal/chat"
	"faulty_in_culture/go_back/internal/infra/logger"
	"faulty_in_culture/go_back/internal/ranking"
	"faulty_in_culture/go_back/internal/savegame"
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
	Ranking  *ranking.Handler
}

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine, h *Handlers) {
	logger.Info("开始设置路由（RESTful风格）")

	api := router.Group("/api")
	{
		// ========== 用户认证（公开接口）==========
		api.POST("/register", h.User.Register)
		api.POST("/login", h.User.Login)

		// ========== 排行榜模块 ==========
		// 查询排行榜（公开接口）
		api.GET("/rankings/:rank_type", h.Ranking.GetRankings)

		// 排行榜管理（需要认证）
		rankingGroup := api.Group("/rankings")
		rankingGroup.Use(middleware.AuthMiddleware())
		{
			rankingGroup.POST("", h.Ranking.UpdateScore)                // 更新分数
			rankingGroup.DELETE("/:rank_type", h.Ranking.DeleteRanking) // 删除指定类型
			rankingGroup.DELETE("", h.Ranking.DeleteAllRankings)        // 删除所有
		}

		// ========== 聊天模块（需要认证）==========
		chatGroup := api.Group("/chat")
		chatGroup.Use(middleware.AuthMiddleware())
		{
			// 会话管理
			chatGroup.GET("/sessions", h.Chat.ListSessions)         // 会话列表
			chatGroup.POST("/sessions", h.Chat.StartChat)           // 创建会话
			chatGroup.GET("/sessions/:id", h.Chat.GetSession)       // 会话详情
			chatGroup.PUT("/sessions/:id", h.Chat.UpdateSession)    // 更新会话
			chatGroup.DELETE("/sessions/:id", h.Chat.DeleteSession) // 删除会话

			// 消息管理
			chatGroup.GET("/sessions/:id/messages", h.Chat.GetHistory)   // 消息历史
			chatGroup.POST("/sessions/:id/messages", h.Chat.SendMessage) // 发送消息
			chatGroup.DELETE("/messages/:id", h.Chat.RecallMessages)     // 撤回消息
		}

		// ========== 存档模块（需要认证）==========
		saveGameGroup := api.Group("/savegame")
		saveGameGroup.Use(middleware.AuthMiddleware())
		{
			saveGameGroup.GET("", h.SaveGame.QueryBySlot)     // 查询指定槽位 ?slot_number=1
			saveGameGroup.GET("/all", h.SaveGame.QueryAll)    // 查询所有存档
			saveGameGroup.POST("", h.SaveGame.CreateOrUpdate) // 创建或更新存档
			saveGameGroup.DELETE("", h.SaveGame.Delete)       // 删除存档 ?slot_number=1
		}
	}

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "RESTful API运行中"})
	})

	logger.Info("路由设置完成",
		zap.Strings("modules", []string{"user", "ranking", "chat", "savegame"}),
	)
}
