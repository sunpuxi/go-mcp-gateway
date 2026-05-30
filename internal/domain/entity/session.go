package entity

import "time"

// Session 表示一个 MCP 会话
type Session struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`        // 关联的客户端 ID（鉴权后填充）
	Permissions     []string  `json:"-"`                // 允许调用的工具名称列表
	ProtocolVersion string    `json:"protocol_version"`
	Initialized     bool      `json:"initialized"`
	CreatedAt       time.Time `json:"created_at"`
	SSECh           chan []byte `json:"-"` // SSE 通道，用于向客户端推送响应
}

// HasPermission 检查是否拥有某个工具的调用权限
func (s *Session) HasPermission(toolName string) bool {
	for _, p := range s.Permissions {
		if p == toolName {
			return true
		}
	}
	return false
}
