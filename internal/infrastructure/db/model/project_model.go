package model

import (
	"time"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// ProjectModel 对应数据库 projects 表
type ProjectModel struct {
	ProjectID   string    `db:"project_id"`
	Name        string    `db:"name"`
	BaseURL     string    `db:"base_url"`
	Description string    `db:"description"`
	Status      int       `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// ToEntity 将数据库模型转换为领域实体
func (m *ProjectModel) ToEntity() entity.Project {
	return entity.Project{
		ProjectID:   m.ProjectID,
		Name:        m.Name,
		BaseURL:     m.BaseURL,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
