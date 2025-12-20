package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Manager WebSocket 连接管理器
type Manager struct {
	clients map[string]*websocket.Conn
	mu      sync.RWMutex
}

// NewManager 创建新的 WebSocket 管理器
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*websocket.Conn),
	}
}

// Register 注册用户连接
func (m *Manager) Register(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[userID] = conn
	log.Printf("用户 %s 已连接", userID)
}

// Unregister 注销用户连接
func (m *Manager) Unregister(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.clients[userID]; ok {
		conn.Close()
		delete(m.clients, userID)
		log.Printf("用户 %s 已断开连接", userID)
	}
}

// IsOnline 检查用户是否在线
func (m *Manager) IsOnline(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[userID]
	return ok
}

// SendMessage 发送消息给指定用户
func (m *Manager) SendMessage(userID string, message []byte) error {
	m.mu.RLock()
	conn, ok := m.clients[userID]
	m.mu.RUnlock()

	if !ok {
		return nil // 用户不在线，不报错
	}

	return conn.WriteMessage(websocket.TextMessage, message)
}

// GetConnection 获取用户连接
func (m *Manager) GetConnection(userID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.clients[userID]
	return conn, ok
}
