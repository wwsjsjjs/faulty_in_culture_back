// Package ranking - 排行榜模块HTTP处理层
// 功能：处理排行榜相关的HTTP请求
// 架构：MVC中的Controller层
package ranking

import (
	errcode "faulty_in_culture/go_back/internal/shared/errors"
	"faulty_in_culture/go_back/internal/shared/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// Handler层（HTTP处理层/控制器层）
// 设计模式：MVC模式中的Controller
// ============================================================

// Handler 排行榜处理器
type Handler struct {
	service *Service
}

// NewHandler 创建排行榜处理器实例（依赖注入）
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetRankings 获取排行榜
// @Summary 获取排行榜
// @Description 获取指定类型的排行榜，按分数降序排列，支持分页查询
// @Tags ranking
// @Produce json
// @Param rank_type path int true "排行榜类型(1-9)"
// @Param page query int false "页码，默认为1"
// @Param limit query int false "每页数量，默认为10，最大100"
// @Success 200 {object} RankingListResponse "排行榜数据"
// @Failure 400 {object} response.Response "排行榜类型无效"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rankings/{rank_type} [get]
func (h *Handler) GetRankings(c *gin.Context) {
	// 解析路径参数
	rankType, _ := strconv.Atoi(c.Param("rank_type"))

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用Service层处理业务逻辑
	rankings, err := h.service.GetRankings(rankType, page, limit)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidRankType)
		return
	}

	// 返回响应
	response.Success(c, RankingListResponse{
		RankType: rankType,
		Page:     page,
		Limit:    limit,
		Rankings: rankings,
	})
}

// UpdateScore 更新排行榜分数（需要认证）
// @Summary 更新排行榜分数
// @Description 更新当前登录用户的指定排行榜分数，只在新分数更高时更新
// @Tags ranking
// @Accept json
// @Produce json
// @Param request body UpdateScoreRequest true "更新分数请求"
// @Success 200 {object} UpdateScoreResponse "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rankings [post]
func (h *Handler) UpdateScore(c *gin.Context) {
	userID := c.GetUint("user_id") // 从中间件获取

	var req UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	// 调用Service层处理业务逻辑
	ranking, err := h.service.UpdateScore(userID, req.RankType, req.Score)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidRankType)
		return
	}

	// 返回响应
	response.Success(c, UpdateScoreResponse{
		UserID:    ranking.UserID,
		RankType:  ranking.RankType,
		Score:     ranking.Score,
		UpdatedAt: ranking.UpdatedAt,
	})
}

// DeleteRanking 删除指定类型的排行榜记录（需要认证）
// @Summary 删除指定类型排行榜记录
// @Description 删除当前登录用户的指定类型排行榜记录
// @Tags ranking
// @Produce json
// @Param rank_type path int true "排行榜类型(1-9)"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rankings/{rank_type} [delete]
func (h *Handler) DeleteRanking(c *gin.Context) {
	userID := c.GetUint("user_id")
	rankType, _ := strconv.Atoi(c.Param("rank_type"))

	if err := h.service.DeleteRanking(userID, rankType); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidRankType)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// DeleteAllRankings 删除所有排行榜记录（需要认证）
// @Summary 删除所有排行榜记录
// @Description 删除当前登录用户的所有排行榜记录
// @Tags ranking
// @Produce json
// @Success 200 {object} map[string]string "删除成功"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rankings [delete]
func (h *Handler) DeleteAllRankings(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := h.service.DeleteAllRankings(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
