package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// Handler层 - MVC的Controller
// 职责：
// 1. 接收HTTP请求
// 2. 参数验证和转换
// 3. 调用Service处理业务逻辑
// 4. 返回HTTP响应
// ============================================================

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Handler 聊天处理器
type Handler struct {
	service *Service
}

// NewHandler 创建聊天处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// StartChat 开始新对话
// @Summary 开始新对话
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body StartRequest true "请求体"
// @Success 200 {object} Response{data=SessionVO}
// @Router /api/chat/start [post]
func (h *Handler) StartChat(c *gin.Context) {
	userID := c.GetUint("user_id") // 从中间件获取

	var req StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "参数错误"})
		return
	}

	session, err := h.service.StartChat(userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
		return
	}

	vo := SessionVO{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		Type:      session.Type,
		CreatedAt: session.CreatedAt,
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: vo})
}

// SendMessage 发送消息（异步）
// @Summary 发送消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "请求体"
// @Success 200 {object} Response{data=MessageVO}
// @Router /api/chat/send [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "参数错误"})
		return
	}

	message, err := h.service.SendMessage(userID, req.SessionID, req.Content)
	if err != nil {
		code := 500
		if err == ErrUnauthorized {
			code = 403
		} else if err == ErrMessageTooMany {
			code = 400
		}
		c.JSON(code, Response{Code: code, Msg: err.Error()})
		return
	}

	vo := MessageVO{
		ID:        message.ID,
		SessionID: message.SessionID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "已发送，AI正在思考...", Data: vo})
}

// GetHistory 获取聊天历史
// @Summary 获取聊天历史
// @Tags 聊天
// @Accept json
// @Produce json
// @Param session_id query int true "会话ID"
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "每页数量" default(50)
// @Success 200 {object} Response{data=HistoryResponse}
// @Router /api/chat/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	sessionID, _ := strconv.ParseUint(c.Query("session_id"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if sessionID == 0 {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "缺少session_id参数"})
		return
	}

	history, err := h.service.GetHistory(userID, uint(sessionID), offset, limit)
	if err != nil {
		code := 500
		if err == ErrSessionNotFound {
			code = 404
		} else if err == ErrUnauthorized {
			code = 403
		}
		c.JSON(code, Response{Code: code, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: history})
}

// RecallMessages 撤回消息
// @Summary 撤回消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param session_id query int true "会话ID"
// @Param message_ids body []uint true "消息ID列表"
// @Success 200 {object} Response
// @Router /api/chat/recall [delete]
func (h *Handler) RecallMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Query("session_id"), 10, 64)

	var messageIDs []uint
	if err := c.ShouldBindJSON(&messageIDs); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "参数错误"})
		return
	}

	if sessionID == 0 || len(messageIDs) == 0 {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "参数错误"})
		return
	}

	err := h.service.RecallMessages(userID, uint(sessionID), messageIDs)
	if err != nil {
		code := 500
		if err == ErrUnauthorized {
			code = 403
		}
		c.JSON(code, Response{Code: code, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "撤回成功"})
}

// ListSessions 获取会话列表
// @Summary 获取会话列表
// @Tags 聊天
// @Produce json
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} Response{data=[]SessionVO}
// @Router /api/chat/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	userID := c.GetUint("user_id")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, err := h.service.ListSessions(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
		return
	}

	vos := make([]SessionVO, len(sessions))
	for i, s := range sessions {
		vos[i] = SessionVO{
			ID:        s.ID,
			UserID:    s.UserID,
			Title:     s.Title,
			Type:      s.Type,
			CreatedAt: s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: vos})
}
