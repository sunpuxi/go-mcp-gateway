package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/didi/gendry/builder"
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
	template := `SELECT t.*, p.base_url
		FROM tools t
		JOIN projects p ON t.project_id = p.project_id
		WHERE t.status = 1 AND p.status = 1`
	sqlStr, args, err := builder.NamedQuery(template, nil)
	if err != nil {
		return nil, err
	}

	var models []model.ToolModel
	if err := r.db.Select(&models, sqlStr, args...); err != nil {
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
	template := `SELECT t.*, p.base_url
		FROM tools t
		JOIN projects p ON t.project_id = p.project_id
		WHERE t.name = {{name}} AND t.status = 1 AND p.status = 1
		LIMIT 1`
	params := map[string]interface{}{"name": name}
	sqlStr, args, err := builder.NamedQuery(template, params)
	if err != nil {
		return nil, err
	}

	var m model.ToolModel
	if err := r.db.Get(&m, sqlStr, args...); err != nil {
		return nil, err
	}

	tool := m.ToEntity()
	return &tool, nil
}

// ======================== Admin CRUD ========================

// ListAll 分页查询所有工具（含禁用的，含 base_url）
func (r *ToolRepo) ListAll(page, size int) ([]entity.Tool, int, error) {
	res, err := builder.AggregateQuery(context.Background(), r.db.DB, "tools", nil, builder.AggregateCount("*"))
	if err != nil {
		return nil, 0, err
	}
	total := int(res.Int64())

	offset := (page - 1) * size
	template := `SELECT t.*, p.base_url
		FROM tools t
		JOIN projects p ON t.project_id = p.project_id
		ORDER BY t.tool_id DESC LIMIT {{limit}} OFFSET {{offset}}`
	params := map[string]interface{}{
		"limit":  size,
		"offset": offset,
	}
	sqlStr, args, err := builder.NamedQuery(template, params)
	if err != nil {
		return nil, 0, err
	}

	var models []model.ToolModel
	if err := r.db.Select(&models, sqlStr, args...); err != nil {
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
	template := `SELECT t.*, p.base_url
		FROM tools t
		JOIN projects p ON t.project_id = p.project_id
		WHERE t.tool_id = {{tool_id}}`
	params := map[string]interface{}{"tool_id": toolID}
	sqlStr, args, err := builder.NamedQuery(template, params)
	if err != nil {
		return nil, err
	}

	var m model.ToolModel
	if err := r.db.Get(&m, sqlStr, args...); err != nil {
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
	data := []map[string]interface{}{{
		"project_id":        tool.ProjectID,
		"name":              tool.Name,
		"title":             tool.Title,
		"description":       tool.Description,
		"http_method":       tool.HTTPMethod,
		"url_template":      tool.URLTemplate,
		"timeout_ms":        tool.TimeoutMs,
		"params":            derefRawMessage(tool.Params),
		"retry_config":      derefRawMessage(retryConfigToRaw(tool.RetryConfig)),
		"rate_limit_config": derefRawMessage(rateLimitConfigToRaw(tool.RateLimitConfig)),
		"status":            tool.Status,
	}}
	sqlStr, args, err := builder.BuildInsert("tools", data)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(sqlStr, args...)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	tool.ToolID = id
	return nil
}

// Update 更新工具
func (r *ToolRepo) Update(tool *entity.Tool) error {
	where := map[string]interface{}{"tool_id": tool.ToolID}
	update := map[string]interface{}{
		"project_id":        tool.ProjectID,
		"name":              tool.Name,
		"title":             tool.Title,
		"description":       tool.Description,
		"http_method":       tool.HTTPMethod,
		"url_template":      tool.URLTemplate,
		"timeout_ms":        tool.TimeoutMs,
		"params":            derefRawMessage(tool.Params),
		"retry_config":      derefRawMessage(retryConfigToRaw(tool.RetryConfig)),
		"rate_limit_config": derefRawMessage(rateLimitConfigToRaw(tool.RateLimitConfig)),
		"status":            tool.Status,
	}
	sqlStr, args, err := builder.BuildUpdate("tools", where, update)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// derefRawMessage 将 *json.RawMessage 转为 interface{}，nil 指针返回 nil（对应 SQL NULL）
func derefRawMessage(raw *json.RawMessage) interface{} {
	if raw == nil {
		return nil
	}
	return *raw
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
	where := map[string]interface{}{"tool_id": toolID}
	sqlStr, args, err := builder.BuildDelete("tools", where)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// DeleteByProjectID 按项目ID删除所有工具
func (r *ToolRepo) DeleteByProjectID(projectID string) error {
	where := map[string]interface{}{"project_id": projectID}
	sqlStr, args, err := builder.BuildDelete("tools", where)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// FindToolIDsByProjectID 按项目ID查询所有工具ID
func (r *ToolRepo) FindToolIDsByProjectID(projectID string) ([]int64, error) {
	where := map[string]interface{}{"project_id": projectID}
	sqlStr, args, err := builder.BuildSelect("tools", where, []string{"tool_id"})
	if err != nil {
		return nil, err
	}
	var ids []int64
	err = r.db.Select(&ids, sqlStr, args...)
	return ids, err
}

// Count 总数
func (r *ToolRepo) Count() (int, error) {
	res, err := builder.AggregateQuery(context.Background(), r.db.DB, "tools", nil, builder.AggregateCount("*"))
	if err != nil {
		return 0, err
	}
	return int(res.Int64()), nil
}
