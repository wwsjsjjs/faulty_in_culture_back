package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// Handler层（HTTP处理层/控制器层）
// 设计模式：MVC模式中的Controller
// 职责：
// 1. 处理HTTP请求和响应
// 2. 参数验证和绑定
// 3. 调用Service层处理业务逻辑
// 4. 错误处理和响应格式化
// 不应该：
// - 包含业务逻辑
// - 直接访问数据库
// - 包含复杂的数据转换逻辑
// ============================================================

// Handler 用户HTTP处理器
type Handler struct {
	service *Service
}

// NewHandler 创建用户处理器实例
// 设计模式：依赖注入
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register 用户注册
// @Summary 用户注册
// @Description 用户通过用户名和密码注册新账号，密码将被加密存储
// @Tags user
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "注册信息"
// @Success 200 {object} RegisterResponse "注册成功，返回用户信息和token"
// @Failure 400 {object} ErrorResponse "参数错误或用户名已存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误"})
		return
	}

	// 调用Service层处理业务逻辑
	user, token, err := h.service.Register(req.Username, req.Password)
	if err != nil {
		if err == ErrUserAlreadyExists {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "服务器错误"})
		}
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过用户名和密码登录，返回token用于后续认证
// @Tags user
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse "登录成功，返回token和用户信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 401 {object} ErrorResponse "用户名或密码错误"
// @Router /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误"})
		return
	}

	// 调用Service层处理业务逻辑
	user, token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户名或密码错误"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User: UserVO{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

// GetRankings 获取排行榜
// @Summary 获取排行榜
// @Description 获取指定类型的用户排行榜，按分数降序排列，支持分页查询
// @Tags user
// @Produce json
// @Param rank_type path int true "排行榜类型(1-9)"
// @Param page query int false "页码，默认为1"
// @Param limit query int false "每页数量，默认为10，最大100"
// @Success 200 {object} RankingListResponse "排行榜数据"
// @Failure 400 {object} ErrorResponse "排行榜类型无效"
// @Failure 500 {object} ErrorResponse "服务器错误"
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
		if err == ErrInvalidRankType {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "服务器错误"})
		}
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, RankingListResponse{
		Page:     page,
		Limit:    limit,
		Rankings: rankings,
	})
}

// UpdateScore 更新用户分数
// @Summary 更新用户分数
// @Description 更新当前登录用户的指定排行榜分数，分数更新后排行榜缓存将被清除
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdateScoreRequest true "分数信息"
// @Success 200 {object} SuccessResponse "更新成功"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/user/score [put]
func (h *Handler) UpdateScore(c *gin.Context) {
	// 从上下文获取用户ID（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未授权"})
		return
	}

	var req UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误"})
		return
	}

	// 调用Service层处理业务逻辑
	err := h.service.UpdateScore(userID.(uint), req.RankType, req.Score)
	if err != nil {
		if err == ErrInvalidRankType || err == ErrInvalidScore {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "服务器错误"})
		}
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, SuccessResponse{Message: "分数更新成功"})
}
