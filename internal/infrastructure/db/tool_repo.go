package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain"
)

// 编译时确保 ToolRepo 实现 ToolQuerier 接口
var _ domain.ToolQuerier = (*ToolRepo)(nil)

type ToolRepo struct {
	db *sqlx.DB
}

func NewToolRepo(db *sqlx.DB) *ToolRepo {
	return &ToolRepo{db: db}
}

// ListEnabled 查询所有启用的工具（含关联的 Project base_url）
func (r *ToolRepo) ListEnabled() ([]domain.Tool, error) {
	sql := `SELECT t.*, p.base_url
FROM tools t
JOIN projects p ON t.project_id = p.project_id
WHERE t.status = 1 AND p.status = 1`

	var tools []domain.Tool
	if err := r.db.Select(&tools, sql); err != nil {
		return nil, err
	}
	return tools, nil
}

// FindByName 根据工具名查询（含关联的 Project base_url）
func (r *ToolRepo) FindByName(name string) (*domain.Tool, error) {
	sql := `SELECT t.*, p.base_url
FROM tools t
JOIN projects p ON t.project_id = p.project_id
WHERE t.name = ? AND t.status = 1 AND p.status = 1
LIMIT 1`

	var tool domain.Tool
	if err := r.db.Get(&tool, sql, name); err != nil {
		return nil, err
	}
	return &tool, nil
}
