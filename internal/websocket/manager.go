package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Manager WebSocket 连接管理器
type Manager struct {
	clients    map[string]*websocket.Conn
	mu         sync.RWMutex
	lastActive map[string]time.Time // 记录用户最后活跃时间
}

// NewManager 创建新的 WebSocket 管理器
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*websocket.Conn),
		lastActive: make(map[string]time.Time),
	}
}

// Register 注册用户连接
func (m *Manager) Register(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[userID] = conn
	m.lastActive[userID] = time.Now()
	log.Printf("用户 %s 已连接", userID)
}

// Unregister 注销用户连接
func (m *Manager) Unregister(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.clients[userID]; ok {
		conn.Close()
		delete(m.clients, userID)
		delete(m.lastActive, userID)
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
	// 发送消息时也更新活跃时间
	m.mu.Lock()
	m.lastActive[userID] = time.Now()
	m.mu.Unlock()

	return conn.WriteMessage(websocket.TextMessage, message)
}

// UpdateActiveTime 更新用户活跃时间（收到消息或Pong时调用）
func (m *Manager) UpdateActiveTime(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastActive[userID] = time.Now()
}

// StartHeartbeat 启动心跳检测和超时清理
func (m *Manager) StartHeartbeat(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			m.mu.RLock()
			for userID, conn := range m.clients {
				// 发送Ping
				if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					log.Printf("Ping 用户 %s 失败: %v", userID, err)
				}
			}
			m.mu.RUnlock()

			// 检查超时
			now := time.Now()
			m.mu.RLock()
			var toRemove []string
			for userID, last := range m.lastActive {
				if now.Sub(last) > timeout {
					toRemove = append(toRemove, userID)
				}
			}
			m.mu.RUnlock()
			for _, userID := range toRemove {
				m.Unregister(userID)
				log.Printf("用户 %s 心跳超时，已强制下线", userID)
			}
		}
	}()
}

// GetConnection 获取用户连接
func (m *Manager) GetConnection(userID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.clients[userID]
	return conn, ok
}
