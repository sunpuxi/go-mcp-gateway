package entity

import (
	"encoding/json"
	"time"
)

// Tool 表示一个 MCP 工具定义（映射到下游 HTTP 接口）
type Tool struct {
	ToolID      int64            `json:"tool_id"`
	ProjectID   string           `json:"project_id"`
	Name        string           `json:"name"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	HTTPMethod  string           `json:"http_method"`
	URLTemplate string           `json:"url_template"`
	BaseURL     string           `json:"base_url"`
	TimeoutMs   int              `json:"timeout_ms"`
	Params      *json.RawMessage `json:"params"`
	RetryConfig     *RetryConfig     `json:"retry_config,omitempty"`
	RateLimitConfig *RateLimitConfig `json:"rate_limit_config,omitempty"`
	Status          int              `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ParseParams 将 Tool 中的 params JSON 解析为 ParamRule 切片
func (t *Tool) ParseParams() ([]ParamRule, error) {
	if t.Params == nil || len(*t.Params) == 0 {
		return nil, nil
	}
	var rules []ParamRule
	if err := json.Unmarshal(*t.Params, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
