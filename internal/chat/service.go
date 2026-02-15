// Package chat - AI聊天模块业务逻辑层
// 功能：实现聊天会话和消息管理的业务规则
// 特点：异步AI调用、消息历史管理、会话CRUD
package chat

import (
	"context"
	"fmt"
	"time"

	"faulty_in_culture/go_back/internal/infra/logger"
	"faulty_in_culture/go_back/internal/infra/ws"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// ============================================================
// Service层 - 业务逻辑层
// 职责：
// 1. 实现核心业务逻辑
// 2. 协调Repository和外部服务
// 3. 处理异步AI调用和WebSocket推送
// ============================================================

// AIClient AI客户端接口（依赖注入）
type AIClient interface {
	Chat(ctx context.Context, messages []map[string]string) (string, error)
	ChatStream(ctx context.Context, messages []map[string]string, callback func(chunk string)) (string, error)
}

// Cache 缓存接口（依赖注入）
type Cache interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

// Service 聊天服务
type Service struct {
	repo      Repository
	aiClient  AIClient
	wsManager *ws.Manager
	cache     Cache
}

// NewService 创建聊天服务实例（依赖注入）
func NewService(repo Repository, aiClient AIClient, wsManager *ws.Manager, cache Cache) *Service {
	return &Service{
		repo:      repo,
		aiClient:  aiClient,
		wsManager: wsManager,
		cache:     cache,
	}
}

// StartChat 开始新对话
func (s *Service) StartChat(userID uint, title string) (*Session, error) {
	logger.Info("[chat.StartChat] 创建新会话", zap.Uint("user_id", userID), zap.String("title", title))

	if title == "" {
		title = fmt.Sprintf("对话-%s", time.Now().Format("0102-1504"))
		logger.Info("[chat.StartChat] 使用自动生成标题", zap.String("title", title))
	}

	session := &Session{
		UserID: userID,
		Title:  title,
		Type:   1,
	}
	err := s.repo.CreateSession(session)
	if err != nil {
		logger.Error("[chat.StartChat] 创建会话失败", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}

	logger.Info("[chat.StartChat] 会话创建成功",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", session.ID),
		zap.String("title", title))
	return session, err
}

// SendMessage 发送消息（异步）
func (s *Service) SendMessage(userID, sessionID uint, content string) (*Message, int, error) {
	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return nil, 0, err
	}
	if session.UserID != userID {
		return nil, 0, fmt.Errorf("未授权")
	}

	// 2. 检查消息数量限制（200条）
	count, err := s.repo.CountMessagesBySessionID(sessionID)
	if err != nil {
		return nil, 0, err
	}
	if count >= 200 {
		return nil, 0, fmt.Errorf("消息数量过多")
	}

	// 3. 保存用户消息
	userMsg := &Message{
		SessionID: sessionID,
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMsg); err != nil {
		return nil, 0, err
	}

	// 4. 异步调用AI并推送
	go s.callAIAndPush(userID, sessionID)

	return userMsg, int(count) + 1, nil
}

// callAIAndPush 异步调用AI并通过WebSocket流式推送
func (s *Service) callAIAndPush(userID, sessionID uint) {
	logger.Info("[chat.callAIAndPush] 开始异步AI处理",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID))

	ctx := context.Background()

	// 1. 获取历史消息
	messages, err := s.repo.FindMessagesBySessionID(sessionID, 0, 50)
	if err != nil {
		logger.Error("[chat.callAIAndPush] 获取历史消息失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return
	}
	logger.Debug("[chat.callAIAndPush] 获取历史消息成功",
		zap.Uint("session_id", sessionID),
		zap.Int("history_count", len(messages)))

	// 2. 构建AI请求（奇数=用户，偶数=AI）
	aiMessages := make([]map[string]string, 0, len(messages))
	for i, msg := range messages {
		role := "user"
		if (i+1)%2 == 0 { // 偶数序号=AI
			role = "assistant"
		}
		aiMessages = append(aiMessages, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}

	// 3. 调用AI
	aiReply, err := s.aiClient.Chat(ctx, aiMessages)
	if err != nil {
		aiReply = "抱歉，AI服务暂时不可用"
	}

	// 4. 保存AI消息
	aiMsg := &Message{
		SessionID: sessionID,
		Content:   aiReply,
	}
	s.repo.CreateMessage(aiMsg)

	// 5. 通过WebSocket推送给用户
	wsMsg := WebSocketMessage{
		Type:         "ai_message",
		SessionID:    sessionID,
		Content:      aiReply,
		MessageIndex: len(messages) + 1, // AI消息序号（偶数）
	}
	s.wsManager.SendToUser(userID, wsMsg)
}

// GetHistory 获取聊天历史
func (s *Service) GetHistory(userID, sessionID uint, offset, limit int) (*HistoryResponse, error) {
	logger.Debug("[chat.GetHistory] 获取聊天历史",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID),
		zap.Int("offset", offset),
		zap.Int("limit", limit))

	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		logger.Error("[chat.GetHistory] 会话查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}
	if session.UserID != userID {
		logger.Warn("[chat.GetHistory] 权限验证失败",
			zap.Uint("user_id", userID),
			zap.Uint("session_id", sessionID))
		return nil, fmt.Errorf("未授权")
	}

	// 2. 获取消息列表
	messages, err := s.repo.FindMessagesBySessionID(sessionID, offset, limit)
	if err != nil {
		logger.Error("[chat.GetHistory] 消息查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}

	// 3. 构建响应
	sessionVO := SessionVO{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		Type:      session.Type,
		CreatedAt: session.CreatedAt,
	}

	messageVOs := make([]MessageVO, len(messages))
	for i, msg := range messages {
		messageVOs[i] = MessageVO{
			ID:           msg.ID,
			SessionID:    msg.SessionID,
			MessageIndex: offset + i + 1,
			Content:      msg.Content,
			CreatedAt:    msg.CreatedAt,
		}
	}

	logger.Debug("[chat.GetHistory] 成功获取历史",
		zap.Uint("session_id", sessionID),
		zap.Int("count", len(messages)))
	return &HistoryResponse{
		Session:  sessionVO,
		Messages: messageVOs,
	}, nil
}

// ListSessions 获取会话列表
func (s *Service) ListSessions(userID uint, offset, limit int) ([]*Session, error) {
	logger.Debug("[chat.ListSessions] 获取会话列表",
		zap.Uint("user_id", userID),
		zap.Int("offset", offset),
		zap.Int("limit", limit))

	sessions, err := s.repo.ListSessionsByUserID(userID, offset, limit)
	if err != nil {
		logger.Error("[chat.ListSessions] 查询失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("[chat.ListSessions] 成功",
		zap.Uint("user_id", userID),
		zap.Int("count", len(sessions)))
	return sessions, nil
}

// GetSession 获取会话详情
func (s *Service) GetSession(userID, sessionID uint) (*Session, error) {
	logger.Debug("[chat.GetSession] 获取会话详情",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID))

	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		logger.Error("[chat.GetSession] 查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}
	if session.UserID != userID {
		logger.Warn("[chat.GetSession] 权限验证失败",
			zap.Uint("user_id", userID),
			zap.Uint("session_id", sessionID))
		return nil, fmt.Errorf("未授权")
	}
	return session, nil
}

// UpdateSession 更新会话
func (s *Service) UpdateSession(userID, sessionID uint, title string) (*Session, error) {
	logger.Info("[chat.UpdateSession] 更新会话",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID),
		zap.String("title", title))

	// 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		logger.Error("[chat.UpdateSession] 会话查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}
	if session.UserID != userID {
		logger.Warn("[chat.UpdateSession] 权限验证失败",
			zap.Uint("user_id", userID),
			zap.Uint("session_id", sessionID))
		return nil, fmt.Errorf("未授权")
	}

	// 更新标题
	session.Title = title
	if err := s.repo.UpdateSession(session); err != nil {
		logger.Error("[chat.UpdateSession] 更新失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}

	logger.Info("[chat.UpdateSession] 更新成功",
		zap.Uint("session_id", sessionID),
		zap.String("title", title))
	return session, nil
}

// DeleteSession 删除会话
func (s *Service) DeleteSession(userID, sessionID uint) error {
	logger.Info("[chat.DeleteSession] 删除会话",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID))

	// 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		logger.Error("[chat.DeleteSession] 会话查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return err
	}
	if session.UserID != userID {
		logger.Warn("[chat.DeleteSession] 权限验证失败",
			zap.Uint("user_id", userID),
			zap.Uint("session_id", sessionID))
		return fmt.Errorf("未授权")
	}

	// 删除会话（级联删除消息）
	err = s.repo.DeleteSession(sessionID)
	if err != nil {
		logger.Error("[chat.DeleteSession] 删除失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return err
	}

	logger.Info("[chat.DeleteSession] 删除成功", zap.Uint("session_id", sessionID))
	return nil
}

// RecallMessages 撤回消息（用户撤回自己的问题+AI的回答）
func (s *Service) RecallMessages(userID, sessionID uint, messageIDs []uint) error {
	logger.Info("[chat.RecallMessages] 撤回消息",
		zap.Uint("user_id", userID),
		zap.Uint("session_id", sessionID),
		zap.Int("count", len(messageIDs)))

	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		logger.Error("[chat.RecallMessages] 会话查询失败",
			zap.Uint("session_id", sessionID),
			zap.Error(err))
		return err
	}
	if session.UserID != userID {
		logger.Warn("[chat.RecallMessages] 权限验证失败",
			zap.Uint("user_id", userID),
			zap.Uint("session_id", sessionID))
		return fmt.Errorf("未授权")
	}

	// 2. 删除消息
	err = s.repo.DeleteMessages(sessionID, messageIDs)
	if err != nil {
		logger.Error("[chat.RecallMessages] 撤回失败",
			zap.Uint("session_id", sessionID),
			zap.Int("count", len(messageIDs)),
			zap.Error(err))
		return err
	}

	logger.Info("[chat.RecallMessages] 撤回成功",
		zap.Uint("session_id", sessionID),
		zap.Int("count", len(messageIDs)))
	return nil
}

// ConnectWebSocket 连接WebSocket
func (s *Service) ConnectWebSocket(userID uint, conn *websocket.Conn) *ws.Client {
	if s.wsManager == nil {
		return nil
	}
	return s.wsManager.Register(userID, conn)
}

// DisconnectWebSocket 断开WebSocket连接
func (s *Service) DisconnectWebSocket(client *ws.Client) {
	if s.wsManager != nil && client != nil {
		s.wsManager.Unregister(client)
	}
}
