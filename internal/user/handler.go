// Package user - 用户模块HTTP处理层
// 功能：处理用户注册、登录相关的HTTP请求
// 架构：MVC中的Controller层
package user

import (
	errcode "faulty_in_culture/go_back/internal/shared/errors"
	"faulty_in_culture/go_back/internal/shared/response"
	"net/http"

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
// @Success 200 {object} AuthResponse "注册成功，返回用户信息和token"
// @Failure 400 {object} response.Response "参数错误或用户名已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	// 调用Service层处理业务逻辑
	user, token, err := h.service.Register(req.Username, req.Password)
	if err != nil {
		response.ErrorWithMessage(c, http.StatusBadRequest, errcode.UserAlreadyExists, err.Error())
		return
	}

	// 返回响应
	response.Success(c, AuthResponse{
		Token: token,
		User: UserVO{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过用户名和密码登录，返回token用于后续认证
// @Tags user
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录信息"
// @Success 200 {object} AuthResponse "登录成功，返回token和用户信息"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Router /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	// 调用Service层处理业务逻辑
	user, token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errcode.InvalidPassword)
		return
	}

	// 返回响应
	response.Success(c, AuthResponse{
		Token: token,
		User: UserVO{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}
