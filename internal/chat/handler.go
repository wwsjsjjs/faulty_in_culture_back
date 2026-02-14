// Package chat - AI聊天模块HTTP处理层
// 功能：处理聊天相关的HTTP请求
// 架构：MVC中的Controller层，提供RESTful API
package chat

import (
	errcode "faulty_in_culture/go_back/internal/shared/errors"
	"faulty_in_culture/go_back/internal/shared/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 聊天处理器
type Handler struct {
	service *Service
}

// NewHandler 创建聊天处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// toSessionVO 转换Session为SessionVO
func toSessionVO(session *Session) SessionVO {
	return SessionVO{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		Type:      session.Type,
		CreatedAt: session.CreatedAt,
	}
}

// handleSessionError 统一处理会话错误
func handleSessionError(c *gin.Context, err error) {
	if strings.Contains(err.Error(), "会话不存在") {
		response.Error(c, http.StatusNotFound, errcode.SessionNotFound)
	} else if strings.Contains(err.Error(), "未授权") {
		response.Error(c, http.StatusForbidden, errcode.Unauthorized)
	} else {
		response.Error(c, http.StatusInternalServerError, errcode.ServerError)
	}
}

// StartChat 创建新会话
// @Summary 创建新会话
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body StartRequest true "请求体"
// @Success 200 {object} response.Response{data=SessionVO}
// @Router /api/chat/sessions [post]
func (h *Handler) StartChat(c *gin.Context) {
	userID := c.GetUint("user_id") // 从中间件获取

	var req StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	session, err := h.service.StartChat(userID, req.Title)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		return
	}

	response.Success(c, toSessionVO(session))
}

// ListSessions 获取会话列表
// @Summary 获取会话列表
// @Tags 聊天
// @Produce json
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=[]SessionVO}
// @Router /api/chat/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	userID := c.GetUint("user_id")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, err := h.service.ListSessions(userID, offset, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		return
	}

	vos := make([]SessionVO, len(sessions))
	for i, s := range sessions {
		vos[i] = toSessionVO(s)
	}

	response.Success(c, vos)
}

// GetSession 获取会话详情
// @Summary 获取会话详情
// @Tags 聊天
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response{data=SessionVO}
// @Router /api/chat/sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if sessionID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	session, err := h.service.GetSession(userID, uint(sessionID))
	if err != nil {
		handleSessionError(c, err)
		return
	}

	response.Success(c, toSessionVO(session))
}

// UpdateSession 更新会话
// @Summary 更新会话
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Param request body UpdateSessionRequest true "请求体"
// @Success 200 {object} response.Response{data=SessionVO}
// @Router /api/chat/sessions/{id} [put]
func (h *Handler) UpdateSession(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if sessionID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	session, err := h.service.UpdateSession(userID, uint(sessionID), req.Title)
	if err != nil {
		handleSessionError(c, err)
		return
	}

	response.Success(c, toSessionVO(session))
}

// DeleteSession 删除会话
// @Summary 删除会话
// @Tags 聊天
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response
// @Router /api/chat/sessions/{id} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if sessionID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	err := h.service.DeleteSession(userID, uint(sessionID))
	if err != nil {
		handleSessionError(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// SendMessage 发送消息
// @Summary 发送消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Param request body SendMessageRequest true "请求体"
// @Success 200 {object} response.Response{data=MessageVO}
// @Router /api/chat/sessions/{id}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if sessionID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	message, err := h.service.SendMessage(userID, uint(sessionID), req.Content)
	if err != nil {
		if strings.Contains(err.Error(), "未授权") {
			response.Error(c, http.StatusForbidden, errcode.Unauthorized)
		} else if strings.Contains(err.Error(), "消息") {
			response.Error(c, http.StatusBadRequest, errcode.MessageTooLong)
		} else {
			response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		}
		return
	}

	vo := MessageVO{
		ID:        message.ID,
		SessionID: message.SessionID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	response.SuccessWithMessage(c, "已发送，AI正在思考..", vo)
}

// GetHistory 获取消息历史
// @Summary 获取消息历史
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "每页数量" default(50)
// @Success 200 {object} response.Response{data=HistoryResponse}
// @Router /api/chat/sessions/{id}/messages [get]
func (h *Handler) GetHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	sessionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if sessionID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	history, err := h.service.GetHistory(userID, uint(sessionID), offset, limit)
	if err != nil {
		handleSessionError(c, err)
		return
	}

	response.Success(c, history)
}

// RecallMessages 撤回消息
// @Summary 撤回消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} response.Response
// @Router /api/chat/messages/{id} [delete]
func (h *Handler) RecallMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	messageID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if messageID == 0 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	// 注意：这里简化处理，实际应该从消息获取session_id
	// 为了保持简单，我们修改为通过query传递session_id
	sessionID, _ := strconv.ParseUint(c.Query("session_id"), 10, 64)
	if sessionID == 0 {
		response.ErrorWithMessage(c, http.StatusBadRequest, errcode.InvalidParams, "缺少session_id参数")
		return
	}

	messageIDs := []uint{uint(messageID)}
	err := h.service.RecallMessages(userID, uint(sessionID), messageIDs)
	if err != nil {
		if strings.Contains(err.Error(), "未授权") {
			response.Error(c, http.StatusForbidden, errcode.Unauthorized)
		} else {
			response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		}
		return
	}

	response.SuccessWithMessage(c, "撤回成功", nil)
}
