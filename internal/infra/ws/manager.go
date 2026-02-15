// Package ws 提供WebSocket实时双向通信
// 功能：管理WebSocket连接，支持实时消息推送
package ws

import (
	"encoding/json"
	"sync"
	"time"

	"faulty_in_culture/go_back/internal/infra/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client WebSocket客户端连接
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte // 发送消息通道
	mu     sync.Mutex
}

// Manager WebSocket连接管理器
type Manager struct {
	clients map[uint][]*Client // userID -> clients
	mu      sync.RWMutex
}

// NewManager 创建WebSocket管理器
func NewManager() *Manager {
	return &Manager{
		clients: make(map[uint][]*Client),
	}
}

// Register 注册新客户端
func (m *Manager) Register(userID uint, conn *websocket.Conn) *Client {
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	m.mu.Lock()
	m.clients[userID] = append(m.clients[userID], client)
	m.mu.Unlock()

	logger.Info("WebSocket客户端已连接", zap.Uint("user_id", userID))
	return client
}

// Unregister 注销客户端
func (m *Manager) Unregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.clients[client.UserID]
	for i, c := range clients {
		if c == client {
			m.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(m.clients[client.UserID]) == 0 {
		delete(m.clients, client.UserID)
	}

	close(client.Send)
	logger.Info("WebSocket客户端已断开", zap.Uint("user_id", client.UserID))
}

// SendToUser 向指定用户的所有客户端推送消息
func (m *Manager) SendToUser(userID uint, message interface{}) error {
	m.mu.RLock()
	clients := m.clients[userID]
	m.mu.RUnlock()

	if len(clients) == 0 {
		logger.Debug("用户不在线，跳过WebSocket推送", zap.Uint("user_id", userID))
		return nil
	}

	// 序列化消息
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 向所有客户端推送
	for _, client := range clients {
		select {
		case client.Send <- data:
			logger.Debug("WebSocket消息已推送", zap.Uint("user_id", userID))
		default:
			logger.Warn("WebSocket客户端缓冲区已满", zap.Uint("user_id", userID))
		}
	}

	return nil
}

// WritePump 处理向客户端写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket读取错误", zap.Error(err))
			}
			break
		}
		// 这里可以处理客户端发来的消息（如果需要）
	}
}

// GetConnectionCount 获取在线用户数
func (m *Manager) GetConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}
