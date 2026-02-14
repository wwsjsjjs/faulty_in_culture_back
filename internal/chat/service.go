// Package chat - AI聊天模块业务逻辑层
// 功能：实现聊天会话和消息管理的业务规则
// 特点：异步AI调用、消息历史管理、会话CRUD
package chat

import (
	"context"
	"fmt"
	"time"
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
}

// WSManager WebSocket管理器接口（依赖注入）
type WSManager interface {
	SendToUser(userID uint, message interface{}) error
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
	wsManager WSManager
	cache     Cache
}

// NewService 创建聊天服务实例（依赖注入）
func NewService(repo Repository, aiClient AIClient, wsManager WSManager, cache Cache) *Service {
	return &Service{
		repo:      repo,
		aiClient:  aiClient,
		wsManager: wsManager,
		cache:     cache,
	}
}

// StartChat 开始新对话
func (s *Service) StartChat(userID uint, title string) (*Session, error) {
	if title == "" {
		title = fmt.Sprintf("对话-%s", time.Now().Format("0102-1504"))
	}

	session := &Session{
		UserID: userID,
		Title:  title,
		Type:   1,
	}
	err := s.repo.CreateSession(session)
	return session, err
}

// SendMessage 发送消息（异步）
func (s *Service) SendMessage(userID, sessionID uint, content string) (*Message, error) {
	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("未授权")
	}

	// 2. 检查消息数量限制（200条）
	count, err := s.repo.CountMessagesBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	if count >= 200 {
		return nil, fmt.Errorf("消息数量过多")
	}

	// 3. 保存用户消息
	userMsg := &Message{
		SessionID: sessionID,
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMsg); err != nil {
		return nil, err
	}

	// 4. 异步调用AI并推送
	go s.callAIAndPush(userID, sessionID)

	return userMsg, nil
}

// callAIAndPush 异步调用AI并通过WebSocket推送
func (s *Service) callAIAndPush(userID, sessionID uint) {
	ctx := context.Background()

	// 1. 获取历史消息
	messages, _ := s.repo.FindMessagesBySessionID(sessionID, 0, 50)

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
	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("未授权")
	}

	// 2. 获取消息列表
	messages, err := s.repo.FindMessagesBySessionID(sessionID, offset, limit)
	if err != nil {
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

	return &HistoryResponse{
		Session:  sessionVO,
		Messages: messageVOs,
	}, nil
}

// ListSessions 获取会话列表
func (s *Service) ListSessions(userID uint, offset, limit int) ([]*Session, error) {
	return s.repo.ListSessionsByUserID(userID, offset, limit)
}

// GetSession 获取会话详情
func (s *Service) GetSession(userID, sessionID uint) (*Session, error) {
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("未授权")
	}
	return session, nil
}

// UpdateSession 更新会话
func (s *Service) UpdateSession(userID, sessionID uint, title string) (*Session, error) {
	// 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("未授权")
	}

	// 更新标题
	session.Title = title
	if err := s.repo.UpdateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// DeleteSession 删除会话
func (s *Service) DeleteSession(userID, sessionID uint) error {
	// 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return fmt.Errorf("未授权")
	}

	// 删除会话（级联删除消息）
	return s.repo.DeleteSession(sessionID)
}

// RecallMessages 撤回消息（用户撤回自己的问题+AI的回答）
func (s *Service) RecallMessages(userID, sessionID uint, messageIDs []uint) error {
	// 1. 验证会话
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return fmt.Errorf("未授权")
	}

	// 2. 删除消息
	return s.repo.DeleteMessages(sessionID, messageIDs)
}
