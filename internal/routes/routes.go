package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/ranking-api/internal/handlers"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine) {
	// 创建处理器实例
	rankingHandler := handlers.NewRankingHandler()

	// API路由组
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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Ranking API is running",
		})
	})
}
