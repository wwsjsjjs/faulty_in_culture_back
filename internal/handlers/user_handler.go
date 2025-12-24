package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"faulty_in_culture/go_back/internal/cache"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/dto"
	"faulty_in_culture/go_back/internal/logger"
	"faulty_in_culture/go_back/internal/models"
	"faulty_in_culture/go_back/internal/vo"
)

// Register 用户注册
// @Summary 用户注册
// @Description 用户注册，传入用户名和密码
// @Tags user
// @Accept json
// @Produce json
// @Param data body dto.UserRegisterRequest true "注册信息"
// @Success 200 {object} vo.UserVO
// @Failure 400 {object} map[string]string
// @Router /api/register [post]
func Register(c *gin.Context) {
	logger.Info("handlers.Register: 开始处理用户注册请求")

	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handlers.Register: 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	logger.Info("handlers.Register: 注册请求参数", zap.String("username", req.Username))

	var user models.User
	db := database.GetDB()
	db.Where("username = ?", req.Username).First(&user)
	if user.ID != 0 {
		logger.Warn("handlers.Register: 用户名已存在", zap.String("username", req.Username))
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("handlers.Register: 密码加密失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user = models.User{
		Username:       req.Username,
		Password:       string(hash),
		CreatedAt:      time.Now(),
		LastLoginAt:    time.Now(),
		Score:          0,
		ScoreUpdatedAt: time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		logger.Error("handlers.Register: 创建用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	token := fmt.Sprintf("%d:%s:%d", user.ID, user.Username, time.Now().Unix())

	logger.Info("handlers.Register: 用户注册成功",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username))

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"token":    token,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录，传入用户名和密码
// @Tags user
// @Accept json
// @Produce json
// @Param data body dto.UserLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/login [post]
func Login(c *gin.Context) {
	logger.Info("handlers.Login: 开始处理用户登录请求")

	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handlers.Login: 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	logger.Info("handlers.Login: 登录请求", zap.String("username", req.Username))

	var user models.User
	cacheKey := fmt.Sprintf("user:username:%s", req.Username)

	// 尝试从缓存获取用户信息
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &user)
		if err == nil && user.ID != 0 {
			logger.Info("handlers.Login: 缓存命中", zap.String("username", req.Username))
			// 缓存命中，验证密码
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
				logger.Warn("handlers.Login: 密码错误", zap.String("username", req.Username))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
				return
			}

			database.GetDB().Model(&user).Update("last_login_at", time.Now())

			token := generateToken(user.ID, user.Username)
			logger.Info("handlers.Login: 登录成功(缓存)", zap.Uint("user_id", user.ID))
			c.JSON(http.StatusOK, gin.H{"token": token, "user": vo.UserVO{ID: user.ID, Username: user.Username}})
			return
		}
	}

	// 缓存未命中，从数据库查询
	logger.Info("handlers.Login: 从数据库查询用户", zap.String("username", req.Username))
	db := database.GetDB()
	db.Where("username = ?", req.Username).First(&user)
	if user.ID == 0 {
		logger.Warn("handlers.Login: 用户不存在", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn("handlers.Login: 密码错误", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	db.Model(&user).Update("last_login_at", time.Now())

	if cacheClient != nil {
		if err := cacheClient.Set(cacheKey, user, 24*time.Hour); err != nil {
			logger.Warn("handlers.Login: 缓存用户信息失败", zap.Error(err))
		}
	}

	// 生成简单 token（示例，实际应用请用 JWT）
	token := generateToken(user.ID, user.Username)
	logger.Info("handlers.Login: 登录成功", zap.Uint("user_id", user.ID), zap.String("username", user.Username))
	c.JSON(http.StatusOK, gin.H{"token": token, "user": vo.UserVO{ID: user.ID, Username: user.Username}})
}

// generateToken 生成简单 token（仅示例，建议用 JWT）
func generateToken(userID uint, username string) string {
	return fmt.Sprintf("%d:%s:%d", userID, username, time.Now().Unix())
}

// GetRankings 获取排行榜
// @Summary 获取排行榜
// @Description 获取用户排行榜，按分数降序排列
// @Tags user
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {array} vo.RankingResponse
// @Failure 500 {object} vo.ErrorResponse
// @Router /api/rankings [get]
func GetRankings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	cacheKey := fmt.Sprintf("rankings:page:%d:limit:%d", page, limit)
	var response []vo.RankingResponse

	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &response)
		if err == nil && len(response) > 0 {
			c.JSON(http.StatusOK, response)
			return
		}
	}

	var users []models.User

	if err := database.DB.Order("score DESC, score_updated_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		logger.Error("获取排名失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}

	response = make([]vo.RankingResponse, len(users))
	for i, u := range users {
		response[i] = vo.RankingResponse{
			ID:        u.ID,
			Username:  u.Username,
			Score:     u.Score,
			Rank:      offset + i + 1,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.ScoreUpdatedAt,
		}
	}

	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, response, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserScore 更新用户分数
// @Summary 更新用户分数
// @Description 更新当前用户的分数
// @Tags user
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.UpdateScoreRequest true "分数信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} vo.ErrorResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/user/score [put]
func UpdateUserScore(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	var req dto.UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateUserScore: 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "参数错误"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "用户不存在"})
		return
	}

	user.Score = req.Score
	user.ScoreUpdatedAt = time.Now()

	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("更新分数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "更新分数失败"})
		return
	}

	cacheClient := cache.GetCache()
	if cacheClient != nil {
		for i := 1; i <= 10; i++ {
			for j := 10; j <= 100; j += 10 {
				_ = cacheClient.Delete(fmt.Sprintf("rankings:page:%d:limit:%d", i, j))
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"score":            user.Score,
		"score_updated_at": user.ScoreUpdatedAt,
	})
}
