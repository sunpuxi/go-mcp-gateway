package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

// 编译时确保 ToolRepo 实现 ToolQuerier 接口
var _ repository.ToolQuerier = (*ToolRepo)(nil)

type ToolRepo struct {
	db *sqlx.DB
}

func NewToolRepo(db *sqlx.DB) *ToolRepo {
	return &ToolRepo{db: db}
}

// ListEnabled 查询所有启用的工具（含关联的 Project base_url）
func (r *ToolRepo) ListEnabled() ([]entity.Tool, error) {
	sql := `SELECT t.*, p.base_url
FROM tools t
JOIN projects p ON t.project_id = p.project_id
WHERE t.status = 1 AND p.status = 1`

	var models []model.ToolModel
	if err := r.db.Select(&models, sql); err != nil {
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
	sql := `SELECT t.*, p.base_url
FROM tools t
JOIN projects p ON t.project_id = p.project_id
WHERE t.name = ? AND t.status = 1 AND p.status = 1
LIMIT 1`

	var m model.ToolModel
	if err := r.db.Get(&m, sql, name); err != nil {
		return nil, err
	}

	tool := m.ToEntity()
	return &tool, nil
}
