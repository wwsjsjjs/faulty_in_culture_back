package handlers

import (
	"fmt"
	"net/http"
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
		Username:  req.Username,
		Password:  string(hash),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		logger.Error("handlers.Register: 创建用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成token
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

			// 生成 token
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

	// 登录成功，缓存用户信息（24小时过期）
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
	// 这里只做简单拼接，生产环境请用 JWT
	return fmt.Sprintf("%d:%s:%d", userID, username, time.Now().Unix())
}
