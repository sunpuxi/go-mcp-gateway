package entity

import "time"

// Session 表示一个 MCP 会话
type Session struct {
	ID              string    `json:"id"`
	ProtocolVersion string    `json:"protocol_version"`
	Initialized     bool      `json:"initialized"`
	CreatedAt       time.Time `json:"created_at"`
	SSECh           chan []byte `json:"-"` // SSE 通道，用于向客户端推送响应
}
