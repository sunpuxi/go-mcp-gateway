package circuitbreaker

import "sync"

// Registry 管理所有 Project 级别的熔断器实例
// 懒初始化：首次访问某个 Project 时自动创建
type Registry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker // key = ProjectID
	cfg      Config                      // 所有新 CB 使用同一份配置
}

// NewRegistry 创建一个新的熔断器注册表
func NewRegistry(cfg Config) *Registry {
	return &Registry{
		breakers: make(map[string]*CircuitBreaker),
		cfg:      cfg,
	}
}

// Get 获取或创建指定 Project 的熔断器
func (r *Registry) Get(projectID string) *CircuitBreaker {
	// 快速路径：读锁
	if cb, ok := r.peek(projectID); ok {
		return cb
	}

	// 慢速路径：写锁创建
	r.mu.Lock()
	defer r.mu.Unlock()

	// 双重检查
	if cb, ok := r.breakers[projectID]; ok {
		return cb
	}

	cb := New(r.cfg)
	r.breakers[projectID] = cb
	return cb
}

// peek 读锁下查找已有实例
func (r *Registry) peek(projectID string) (*CircuitBreaker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cb, ok := r.breakers[projectID]
	return cb, ok
}

// Stats 返回所有 Project 的熔断器统计信息
func (r *Registry) Stats() map[string]Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Stats, len(r.breakers))
	for projectID, cb := range r.breakers {
		result[projectID] = cb.Stats()
	}
	return result
}

// Reset 重置指定 Project 的熔断器（主要用于测试和手动恢复）
func (r *Registry) Reset(projectID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.breakers, projectID)
}
