package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/ranking-api/internal/handlers"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置所有路由
// 类型：Gin 路由注册函数
// 功能：注册所有 API 路由、Swagger 文档路由和健康检查路由，将 HTTP 路径与对应的 handler 绑定。
func SetupRoutes(router *gin.Engine) {
	// 创建处理器实例
	// 创建排名业务处理器（Gin handler，业务逻辑层）
	rankingHandler := handlers.NewRankingHandler()

	// API路由组
	// 注册 API 路由组
	api := router.Group("/api")
	{
		// 排名相关路由
		rankings := api.Group("/rankings")
		{
			rankings.POST("", rankingHandler.CreateRanking)       // 创建排名
			rankings.GET("", rankingHandler.GetRankings)          // 获取所有排名（分页）
			rankings.GET("/top", rankingHandler.GetTopRankings)   // 获取前N名
			rankings.GET("/:id", rankingHandler.GetRanking)       // 获取单个排名
			rankings.PUT("/:id", rankingHandler.UpdateRanking)    // 更新排名
			rankings.DELETE("/:id", rankingHandler.DeleteRanking) // 删除排名
		}
	}

	// Swagger文档路由
	// 注册 Swagger 文档路由（第三方中间件，API 文档展示）
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	// 注册健康检查路由（Gin handler，基础运维功能）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Ranking API is running",
		})
	})
}
