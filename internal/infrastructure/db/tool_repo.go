package db

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

var (
	_ repository.ToolQuerier        = (*ToolRepo)(nil)
	_ repository.ToolAdminRepository = (*ToolRepo)(nil)
)

// ToolRepo 工具仓储
type ToolRepo struct {
	db *sqlx.DB
}

// NewToolRepo 创建 ToolRepo
func NewToolRepo(db *sqlx.DB) *ToolRepo {
	return &ToolRepo{db: db}
}

// ======================== ToolQuerier ========================

// ListEnabled 查询所有启用的工具（含关联的 Project base_url）
func (r *ToolRepo) ListEnabled() ([]entity.Tool, error) {
	sqlStr := `SELECT t.*, p.base_url
	FROM tools t
	JOIN projects p ON t.project_id = p.project_id
	WHERE t.status = 1 AND p.status = 1`

	var models []model.ToolModel
	if err := r.db.Select(&models, sqlStr); err != nil {
		return nil, err
	}

	tools := make([]entity.Tool, len(models))
	for i, m := range models {
		tools[i] = m.ToEntity()
	}
	return tools, nil
}

// FindByName 根据工具名查询（含关联的 Project base_url）
func (r *ToolRepo) FindByName(name string) (*entity.Tool, error) {
	sqlStr := `SELECT t.*, p.base_url
	FROM tools t
	JOIN projects p ON t.project_id = p.project_id
	WHERE t.name = ? AND t.status = 1 AND p.status = 1
	LIMIT 1`

	var m model.ToolModel
	if err := r.db.Get(&m, sqlStr, name); err != nil {
		return nil, err
	}

	tool := m.ToEntity()
	return &tool, nil
}

// ======================== Admin CRUD ========================

// ListAll 分页查询所有工具（含禁用的，含 base_url）
func (r *ToolRepo) ListAll(page, size int) ([]entity.Tool, int, error) {
	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM tools"); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT t.*, p.base_url
	FROM tools t
	JOIN projects p ON t.project_id = p.project_id
	ORDER BY t.tool_id DESC LIMIT ? OFFSET ?`

	var models []model.ToolModel
	if err := r.db.Select(&models, query, size, offset); err != nil {
		return nil, 0, err
	}

	tools := make([]entity.Tool, len(models))
	for i, m := range models {
		tools[i] = m.ToEntity()
	}
	return tools, total, nil
}

// FindByID 按 tool_id 查找
func (r *ToolRepo) FindByID(toolID int64) (*entity.Tool, error) {
	query := `SELECT t.*, p.base_url
	FROM tools t
	JOIN projects p ON t.project_id = p.project_id
	WHERE t.tool_id = ?`

	var m model.ToolModel
	if err := r.db.Get(&m, query, toolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	t := m.ToEntity()
	return &t, nil
}

// Create 新建工具，params、retry_config 和 rate_limit_config 为 JSON
func (r *ToolRepo) Create(tool *entity.Tool) error {
	query := `INSERT INTO tools (project_id, name, title, description, http_method, url_template, timeout_ms, params, retry_config, rate_limit_config, status)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query,
		tool.ProjectID, tool.Name, tool.Title, tool.Description,
		tool.HTTPMethod, tool.URLTemplate, tool.TimeoutMs, tool.Params,
		retryConfigToRaw(tool.RetryConfig),
		rateLimitConfigToRaw(tool.RateLimitConfig),
		tool.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	tool.ToolID = id
	return nil
}

// Update 更新工具
func (r *ToolRepo) Update(tool *entity.Tool) error {
	query := `UPDATE tools SET project_id=?, name=?, title=?, description=?, http_method=?,
	url_template=?, timeout_ms=?, params=?, retry_config=?, rate_limit_config=?, status=? WHERE tool_id=?`
	_, err := r.db.Exec(query,
		tool.ProjectID, tool.Name, tool.Title, tool.Description,
		tool.HTTPMethod, tool.URLTemplate, tool.TimeoutMs, tool.Params,
		retryConfigToRaw(tool.RetryConfig),
		rateLimitConfigToRaw(tool.RateLimitConfig),
		tool.Status, tool.ToolID,
	)
	return err
}

// retryConfigToRaw 将 RetryConfig 转为 json.RawMessage（nil → nil）
func retryConfigToRaw(cfg *entity.RetryConfig) *json.RawMessage {
	if cfg == nil {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	raw := json.RawMessage(data)
	return &raw
}

// rateLimitConfigToRaw 将 RateLimitConfig 转为 json.RawMessage（nil → nil）
func rateLimitConfigToRaw(cfg *entity.RateLimitConfig) *json.RawMessage {
	if cfg == nil {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	raw := json.RawMessage(data)
	return &raw
}

// Delete 删除工具
func (r *ToolRepo) Delete(toolID int64) error {
	_, err := r.db.Exec("DELETE FROM tools WHERE tool_id=?", toolID)
	return err
}

// DeleteByProjectID 按项目ID删除所有工具
func (r *ToolRepo) DeleteByProjectID(projectID string) error {
	_, err := r.db.Exec("DELETE FROM tools WHERE project_id=?", projectID)
	return err
}

// FindToolIDsByProjectID 按项目ID查询所有工具ID
func (r *ToolRepo) FindToolIDsByProjectID(projectID string) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids, "SELECT tool_id FROM tools WHERE project_id=?", projectID)
	return ids, err
}

// Count 总数
func (r *ToolRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM tools")
	return count, err
}
