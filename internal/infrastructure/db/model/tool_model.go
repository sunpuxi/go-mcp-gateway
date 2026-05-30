package model

import (
	"encoding/json"
	"time"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// ToolModel 对应数据库 tools 表结构，含 JOIN 查询的 project.base_url
type ToolModel struct {
	ToolID      int64            `db:"tool_id"`
	ProjectID   string           `db:"project_id"`
	Name        string           `db:"name"`
	Title       string           `db:"title"`
	Description string           `db:"description"`
	HTTPMethod  string           `db:"http_method"`
	URLTemplate string           `db:"url_template"`
	BaseURL     string           `db:"base_url"`
	TimeoutMs   int              `db:"timeout_ms"`
	Params      *json.RawMessage `db:"params"`
	RetryConfig *json.RawMessage `db:"retry_config"`
	Status      int              `db:"status"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
}

// ToEntity 将数据库模型转换为领域实体
func (m *ToolModel) ToEntity() entity.Tool {
	return entity.Tool{
		ToolID:      m.ToolID,
		ProjectID:   m.ProjectID,
		Name:        m.Name,
		Title:       m.Title,
		Description: m.Description,
		HTTPMethod:  m.HTTPMethod,
		URLTemplate: m.URLTemplate,
		BaseURL:     m.BaseURL,
		TimeoutMs:   m.TimeoutMs,
		Params:      m.Params,
		RetryConfig: parseRetryConfigSafe(m.RetryConfig),
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// parseRetryConfigSafe 安全解析 retry_config JSON，解析失败返回 nil（不阻塞主流程）
func parseRetryConfigSafe(raw *json.RawMessage) *entity.RetryConfig {
	cfg, err := entity.ParseRetryConfig(raw)
	if err != nil {
		return nil
	}
	return cfg
}
