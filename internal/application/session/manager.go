package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// Manager 管理内存中的 MCP 会话（应用层状态管理）
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*entity.Session
	ttl      time.Duration
}

// NewManager 创建会话管理器，ttl 为会话过期时间，0 表示永不过期
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		sessions: make(map[string]*entity.Session),
		ttl:      ttl,
	}
	if ttl > 0 {
		go m.cleanupLoop()
	}
	return m
}

// Create 创建一个新会话，关联客户端 ID 和权限列表
func (m *Manager) Create(clientID string, permissions []string) *entity.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &entity.Session{
		ID:          uuid.New().String(),
		ClientID:    clientID,
		Permissions: permissions,
		CreatedAt:   time.Now(),
	}
	m.sessions[session.ID] = session
	return session
}

// Get 获取会话
func (m *Manager) Get(id string) (*entity.Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	if m.ttl > 0 && time.Since(s.CreatedAt) > m.ttl {
		return nil, false
	}
	return s, true
}

// Delete 删除会话
func (m *Manager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

// MarkInitialized 标记会话已初始化
func (m *Manager) MarkInitialized(id string, protocolVersion string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return false
	}
	s.Initialized = true
	s.ProtocolVersion = protocolVersion
	return true
}

// ActiveCount 返回当前活跃会话数
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// List 返回当前所有活跃会话
func (m *Manager) List() []*entity.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*entity.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

// cleanupLoop 定期清理过期会话
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, s := range m.sessions {
			if now.Sub(s.CreatedAt) > m.ttl {
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()
	}
}
