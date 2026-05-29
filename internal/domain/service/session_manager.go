package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// SessionManager 管理内存中的 MCP 会话
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*entity.Session
	ttl      time.Duration
}

// NewSessionManager 创建会话管理器，ttl 为会话过期时间，0 表示永不过期
func NewSessionManager(ttl time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*entity.Session),
		ttl:      ttl,
	}
	if ttl > 0 {
		go sm.cleanupLoop()
	}
	return sm
}

// Create 创建一个新会话
func (sm *SessionManager) Create() *entity.Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &entity.Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
	}
	sm.sessions[session.ID] = session
	return session
}

// Get 获取会话
func (sm *SessionManager) Get(id string) (*entity.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[id]
	if !ok {
		return nil, false
	}
	if sm.ttl > 0 && time.Since(s.CreatedAt) > sm.ttl {
		return nil, false
	}
	return s, true
}

// Delete 删除会话
func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

// MarkInitialized 标记会话已初始化
func (sm *SessionManager) MarkInitialized(id string, protocolVersion string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return false
	}
	s.Initialized = true
	s.ProtocolVersion = protocolVersion
	return true
}

// cleanupLoop 定期清理过期会话
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, s := range sm.sessions {
			if now.Sub(s.CreatedAt) > sm.ttl {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}
