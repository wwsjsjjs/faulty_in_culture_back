package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/yourusername/ranking-api/internal/cache"
	"github.com/yourusername/ranking-api/internal/database"
	"github.com/yourusername/ranking-api/internal/dto"
	"github.com/yourusername/ranking-api/internal/models"
	"github.com/yourusername/ranking-api/internal/vo"
	ws "github.com/yourusername/ranking-api/internal/websocket"
)

// ChatHandler AI聊天处理器
type ChatHandler struct {
	wsManager *ws.Manager
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(manager *ws.Manager) *ChatHandler {
	return &ChatHandler{
		wsManager: manager,
	}
}

// StartChat 开始新的聊天会话
// @Summary 开始新的聊天会话
// @Description 创建一个新的AI聊天会话
// @Tags chat
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.ChatStartRequest false "聊天标题"
// @Success 200 {object} vo.ChatSessionResponse
// @Failure 401 {object} vo.ErrorResponse
// @Router /api/chat/start [post]
func (h *ChatHandler) StartChat(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	var req dto.ChatStartRequest
	_ = c.ShouldBindJSON(&req)

	title := req.Title
	if title == "" {
		title = "新对话"
	}

	session := models.ChatSession{
		UserID: userID,
		Title:  title,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "创建会话失败"})
		return
	}

	// 清除会话列表缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete(fmt.Sprintf("chat:sessions:user:%d", userID))
	}

	c.JSON(http.StatusOK, vo.ChatSessionResponse{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	})
}

// SendMessage 发送消息给AI
// @Summary 发送消息给AI
// @Description 向指定会话发送消息，异步调用AI并通过WebSocket返回
// @Tags chat
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body dto.ChatMessageRequest true "消息内容"
// @Success 200 {object} map[string]string
// @Failure 401 {object} vo.ErrorResponse
// @Failure 400 {object} vo.ErrorResponse
// @Router /api/chat/send [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	var req dto.ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "参数错误"})
		return
	}

	// 验证会话是否属于当前用户
	var session models.ChatSession
	if err := database.DB.Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "会话不存在或无权访问"})
		return
	}

	// 保存用户消息
	userMsg := models.ChatMessage{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Content,
	}
	if err := database.DB.Create(&userMsg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "保存消息失败"})
		return
	}

	// 异步调用AI
	go h.callAIAndRespond(userID, req.SessionID, req.Content)

	c.JSON(http.StatusOK, gin.H{
		"message":    "消息已发送，AI正在思考中...",
		"session_id": req.SessionID,
	})
}

// callAIAndRespond 调用AI并通过WebSocket推送响应
func (h *ChatHandler) callAIAndRespond(userID uint, sessionID uint, userContent string) {
	// 获取历史消息
	var messages []models.ChatMessage
	database.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages)

	// 构建消息列表
	aiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "user" {
			aiMessages = append(aiMessages, openai.UserMessage(msg.Content))
		} else {
			aiMessages = append(aiMessages, openai.AssistantMessage(msg.Content))
		}
	}

	// 调用混元AI
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("HUNYUAN_API_KEY")),
		option.WithBaseURL("https://api.hunyuan.cloud.tencent.com/v1/"),
	)

	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: aiMessages,
			Model:    "hunyuan-turbo",
		},
		option.WithJSONSet("enable_enhancement", true),
	)

	if err != nil {
		log.Printf("AI调用失败: %v", err)
		h.sendErrorToUser(userID, sessionID, "AI调用失败")
		return
	}

	aiResponse := chatCompletion.Choices[0].Message.Content

	// 保存AI回复
	assistantMsg := models.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   aiResponse,
	}
	if err := database.DB.Create(&assistantMsg).Error; err != nil {
		log.Printf("保存AI回复失败: %v", err)
	}

	// 通过WebSocket推送给用户
	h.sendMessageToUser(userID, sessionID, aiResponse)
}

// sendMessageToUser 通过WebSocket推送消息给用户
func (h *ChatHandler) sendMessageToUser(userID uint, sessionID uint, content string) {
	userIDStr := fmt.Sprintf("%d", userID)

	msg := map[string]interface{}{
		"type":       "chat_response",
		"session_id": sessionID,
		"content":    content,
	}

	data, _ := json.Marshal(msg)

	if err := h.wsManager.SendMessage(userIDStr, data); err != nil {
		log.Printf("WebSocket推送失败: %v", err)
	}
}

// sendErrorToUser 发送错误消息给用户
func (h *ChatHandler) sendErrorToUser(userID uint, sessionID uint, errMsg string) {
	userIDStr := fmt.Sprintf("%d", userID)

	msg := map[string]interface{}{
		"type":       "chat_error",
		"session_id": sessionID,
		"error":      errMsg,
	}

	data, _ := json.Marshal(msg)
	h.wsManager.SendMessage(userIDStr, data)
}

// GetChatHistory 获取聊天历史
// @Summary 获取聊天历史
// @Description 获取指定会话的聊天历史
// @Tags chat
// @Security ApiKeyAuth
// @Produce json
// @Param session_id path int true "会话ID"
// @Success 200 {object} vo.ChatHistoryResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/chat/{session_id} [get]
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	sessionID := c.Param("session_id")

	// 验证会话是否属于当前用户
	var session models.ChatSession
	if err := database.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "会话不存在或无权访问"})
		return
	}

	// 获取消息
	var messages []models.ChatMessage
	database.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages)

	messageResponses := make([]vo.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = vo.ChatMessageResponse{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, vo.ChatHistoryResponse{
		Session: vo.ChatSessionResponse{
			ID:        session.ID,
			UserID:    session.UserID,
			Title:     session.Title,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		},
		Messages: messageResponses,
	})
}

// GetChatSessions 获取用户所有聊天会话
// @Summary 获取用户所有聊天会话
// @Description 获取当前用户的所有聊天会话列表
// @Tags chat
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} vo.ChatSessionResponse
// @Failure 401 {object} vo.ErrorResponse
// @Router /api/chat/sessions [get]
func (h *ChatHandler) GetChatSessions(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("chat:sessions:user:%d", userID)
	var responses []vo.ChatSessionResponse

	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &responses)
		if err == nil && len(responses) >= 0 {
			c.JSON(http.StatusOK, responses)
			return
		}
	}

	var sessions []models.ChatSession
	database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&sessions)

	responses = make([]vo.ChatSessionResponse, len(sessions))
	for i, s := range sessions {
		responses[i] = vo.ChatSessionResponse{
			ID:        s.ID,
			UserID:    s.UserID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}

	// 缓存结果（3分钟）
	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, responses, 3*time.Minute)
	}

	c.JSON(http.StatusOK, responses)
}
