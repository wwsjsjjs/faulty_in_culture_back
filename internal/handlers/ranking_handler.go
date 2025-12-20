package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ranking-api/internal/cache"
	"github.com/yourusername/ranking-api/internal/database"
	"github.com/yourusername/ranking-api/internal/dto"
	"github.com/yourusername/ranking-api/internal/models"
	"github.com/yourusername/ranking-api/internal/vo"
)

// RankingHandler 排名处理器
// 类型：Gin handler 结构体
// 功能：封装所有与排名相关的 HTTP 业务处理方法
type RankingHandler struct{}

// NewRankingHandler 创建排名处理器实例
// 类型：构造函数
// 功能：返回 RankingHandler 实例，供路由注册使用
func NewRankingHandler() *RankingHandler {
	return &RankingHandler{}
}

// CreateRanking 创建新排名
// 类型：Gin handler 方法
// 功能：处理 POST /api/rankings 请求，创建新的用户排名
// @Summary 创建新的排名记录
// @Description 创建一个新的用户排名记录
// @Tags rankings
// @Accept json
// @Produce json
// @Param ranking body dto.CreateRankingRequest true "排名信息"
// @Success 201 {object} models.Ranking
// @Failure 400 {object} vo.ErrorResponse
// @Failure 409 {object} vo.ErrorResponse
// @Router /api/rankings [post]
func (h *RankingHandler) CreateRanking(c *gin.Context) {
	var req dto.CreateRankingRequest

	// 绑定并验证请求
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existing models.Ranking
	if err := database.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, vo.ErrorResponse{Error: "username already exists"})
		return
	}

	// 创建新记录
	ranking := models.Ranking{
		Username: req.Username,
		Score:    req.Score,
	}

	if err := database.DB.Create(&ranking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 清除相关缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete("rankings:all")
		_ = cacheClient.Delete("rankings:top")
	}

	c.JSON(http.StatusCreated, ranking)
}

// GetRankings 获取所有排名（按分数降序）
// 类型：Gin handler 方法
// 功能：处理 GET /api/rankings 请求，分页获取所有用户排名
// @Summary 获取排名列表
// @Description 获取所有排名记录，按分数降序排列
// @Tags rankings
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {array} vo.RankingResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/rankings [get]
func (h *RankingHandler) GetRankings(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 构建缓存键
	cacheKey := fmt.Sprintf("rankings:page:%d:limit:%d", page, limit)
	var response []vo.RankingResponse

	// 尝试从缓存获取
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &response)
		if err == nil && len(response) > 0 {
			// 缓存命中
			c.JSON(http.StatusOK, response)
			return
		}
	}

	var rankings []models.Ranking

	// 查询排名，按分数降序
	if err := database.DB.Order("score DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rankings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 计算排名并转换为响应格式
	response = make([]vo.RankingResponse, len(rankings))
	for i, r := range rankings {
		response[i] = vo.RankingResponse{
			ID:        r.ID,
			Username:  r.Username,
			Score:     r.Score,
			Rank:      offset + i + 1,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}

	// 缓存结果（5分钟过期）
	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, response, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}

// GetRanking 根据ID获取单个排名
// 类型：Gin handler 方法
// 功能：处理 GET /api/rankings/:id 请求，获取指定 ID 的排名
// @Summary 获取单个排名
// @Description 根据ID获取单个排名记录
// @Tags rankings
// @Produce json
// @Param id path int true "排名ID"
// @Success 200 {object} models.Ranking
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/rankings/{id} [get]
func (h *RankingHandler) GetRanking(c *gin.Context) {
	id := c.Param("id")

	var ranking models.Ranking
	if err := database.DB.First(&ranking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "ranking not found"})
		return
	}

	c.JSON(http.StatusOK, ranking)
}

// UpdateRanking 更新排名
// 类型：Gin handler 方法
// 功能：处理 PUT /api/rankings/:id 请求，更新指定 ID 的排名信息
// @Summary 更新排名记录
// @Description 根据ID更新排名记录
// @Tags rankings
// @Accept json
// @Produce json
// @Param id path int true "排名ID"
// @Param ranking body dto.UpdateRankingRequest true "更新信息"
// @Success 200 {object} models.Ranking
// @Failure 400 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/rankings/{id} [put]
func (h *RankingHandler) UpdateRanking(c *gin.Context) {
	id := c.Param("id")

	var ranking models.Ranking
	if err := database.DB.First(&ranking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "ranking not found"})
		return
	}

	var req dto.UpdateRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 更新字段
	if req.Username != "" {
		// 检查新用户名是否已被其他记录使用
		var existing models.Ranking
		if err := database.DB.Where("username = ? AND id != ?", req.Username, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, vo.ErrorResponse{Error: "username already exists"})
			return
		}
		ranking.Username = req.Username
	}

	if req.Score != nil {
		ranking.Score = *req.Score
	}

	if err := database.DB.Save(&ranking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 清除相关缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete("rankings:all")
		_ = cacheClient.Delete("rankings:top")
		// 清除所有分页缓存（简单处理，生产环境可用更精细的方式）
		for i := 1; i <= 10; i++ {
			for j := 10; j <= 100; j += 10 {
				_ = cacheClient.Delete(fmt.Sprintf("rankings:page:%d:limit:%d", i, j))
			}
		}
	}

	c.JSON(http.StatusOK, ranking)
}

// DeleteRanking 删除排名
// 类型：Gin handler 方法
// 功能：处理 DELETE /api/rankings/:id 请求，软删除指定 ID 的排名
// @Summary 删除排名记录
// @Description 根据ID删除排名记录（软删除）
// @Tags rankings
// @Produce json
// @Param id path int true "排名ID"
// @Success 200 {object} vo.SuccessMessageResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/rankings/{id} [delete]
func (h *RankingHandler) DeleteRanking(c *gin.Context) {
	id := c.Param("id")

	var ranking models.Ranking
	if err := database.DB.First(&ranking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "ranking not found"})
		return
	}

	// 软删除
	if err := database.DB.Delete(&ranking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	// 清除相关缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete("rankings:all")
		_ = cacheClient.Delete("rankings:top")
		// 清除所有分页缓存
		for i := 1; i <= 10; i++ {
			for j := 10; j <= 100; j += 10 {
				_ = cacheClient.Delete(fmt.Sprintf("rankings:page:%d:limit:%d", i, j))
			}
		}
	}

	c.JSON(http.StatusOK, vo.SuccessMessageResponse{Message: "ranking deleted successfully"})
}

// GetTopRankings 获取前N名排名
// 类型：Gin handler 方法
// 功能：处理 GET /api/rankings/top 请求，获取分数最高的前 N 名用户
// @Summary 获取排行榜前N名
// @Description 获取分数最高的前N名用户
// @Tags rankings
// @Produce json
// @Param top query int false "前N名" default(10)
// @Success 200 {array} vo.RankingResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/rankings/top [get]
func (h *RankingHandler) GetTopRankings(c *gin.Context) {
	top, _ := strconv.Atoi(c.DefaultQuery("top", "10"))

	if top < 1 || top > 100 {
		top = 10
	}

	// 构建缓存键
	cacheKey := fmt.Sprintf("rankings:top:%d", top)
	var response []vo.RankingResponse

	// 尝试从缓存获取
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &response)
		if err == nil && len(response) > 0 {
			// 缓存命中
			c.JSON(http.StatusOK, response)
			return
		}
	}

	var rankings []models.Ranking

	if err := database.DB.Order("score DESC, created_at ASC").
		Limit(top).
		Find(&rankings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	response = make([]vo.RankingResponse, len(rankings))
	for i, r := range rankings {
		response[i] = vo.RankingResponse{
			ID:        r.ID,
			Username:  r.Username,
			Score:     r.Score,
			Rank:      i + 1,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}

	// 缓存结果（5分钟过期）
	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, response, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}
