package dto

import "github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"

// ToolDTO 管理后台工具响应/请求，params 为 ParamRule 数组而非 RawMessage
type ToolDTO struct {
	ToolID      int64              `json:"tool_id"`
	ProjectID   string             `json:"project_id"`
	Name        string             `json:"name"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	HTTPMethod  string             `json:"http_method"`
	URLTemplate string             `json:"url_template"`
	BaseURL     string             `json:"base_url"`
	TimeoutMs   int                `json:"timeout_ms"`
	Params      []entity.ParamRule `json:"params"`
	RetryConfig *entity.RetryConfig `json:"retry_config,omitempty"`
	Status      int                `json:"status"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// ProjectDTO 项目管理请求/响应
type ProjectDTO struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// ClientDTO 客户端管理响应
type ClientDTO struct {
	ClientID     string `json:"client_id"`
	Name         string `json:"name"`
	APIKeyPrefix string `json:"api_key_prefix"`
	Description  string `json:"description"`
	Status       int    `json:"status"`
	ToolCount    int    `json:"tool_count"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// StatsDTO 仪表盘统计
type StatsDTO struct {
	Projects    int              `json:"projects"`
	Tools       int              `json:"tools"`
	Clients     int              `json:"clients"`
	Sessions    int              `json:"sessions"`
	SessionList []SessionInfoDTO `json:"session_list"`
}

// SessionInfoDTO 活跃会话信息
type SessionInfoDTO struct {
	ID              string `json:"id"`
	ClientID        string `json:"client_id"`
	ProtocolVersion string `json:"protocol_version"`
	Initialized     bool   `json:"initialized"`
	CreatedAt       string `json:"created_at"`
}
