package domain

import (
	"encoding/json"
	"time"
)

// Project 表示下游 HTTP 服务
type Project struct {
	ProjectID   string    `db:"project_id" json:"project_id"`
	Name        string    `db:"name" json:"name"`
	BaseURL     string    `db:"base_url" json:"base_url"`
	Description string    `db:"description" json:"description"`
	Status      int       `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Tool 表示一个 MCP 工具定义（映射到下游 HTTP 接口）
type Tool struct {
	ToolID      int64            `db:"tool_id" json:"tool_id"`
	ProjectID   string           `db:"project_id" json:"project_id"`
	Name        string           `db:"name" json:"name"`
	Title       string           `db:"title" json:"title"`
	Description string           `db:"description" json:"description"`
	HTTPMethod  string           `db:"http_method" json:"http_method"`
	URLTemplate string           `db:"url_template" json:"url_template"`
	BaseURL     string           `db:"base_url" json:"base_url"`
	TimeoutMs   int              `db:"timeout_ms" json:"timeout_ms"`
	Params      *json.RawMessage `db:"params" json:"params"`
	Status      int              `db:"status" json:"status"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
}

// ParamRule 是 params JSON 字段中单条参数映射规则
type ParamRule struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Location     string `json:"location"` // path / query / body / header
	Required     bool   `json:"required,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	Description  string `json:"description,omitempty"`
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

// BuildInputSchema 将 ParamRule 切片转为 MCP inputSchema（JSON Schema 格式）
func BuildInputSchema(rules []ParamRule) json.RawMessage {
	type propertySchema struct {
		Type        string `json:"type"`
		Description string `json:"description,omitempty"`
		Default     any    `json:"default,omitempty"`
	}

	type inputSchema struct {
		Type       string                    `json:"type"`
		Properties map[string]propertySchema `json:"properties"`
		Required   []string                  `json:"required,omitempty"`
	}

	schema := inputSchema{
		Type:       "object",
		Properties: make(map[string]propertySchema),
	}

	for _, r := range rules {
		prop := propertySchema{
			Type:        r.Type,
			Description: r.Description,
		}
		if r.DefaultValue != "" && !r.Required {
			prop.Default = r.DefaultValue
		}
		schema.Properties[r.Name] = prop
		if r.Required {
			schema.Required = append(schema.Required, r.Name)
		}
	}

	data, _ := json.Marshal(schema)
	return data
}
