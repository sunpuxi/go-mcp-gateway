package db

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

var _ repository.ProjectRepository = (*ProjectRepo)(nil)

// ProjectRepo 是 ProjectRepository 的 MySQL 实现
type ProjectRepo struct {
	db *sqlx.DB
}

// NewProjectRepo 创建 ProjectRepo
func NewProjectRepo(db *sqlx.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

// List 分页查询所有项目
func (r *ProjectRepo) List(page, size int) ([]entity.Project, int, error) {
	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM projects"); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT project_id, name, base_url, description, status, created_at, updated_at
FROM projects ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var models []model.ProjectModel
	if err := r.db.Select(&models, query, size, offset); err != nil {
		return nil, 0, err
	}

	projects := make([]entity.Project, len(models))
	for i, m := range models {
		projects[i] = m.ToEntity()
	}
	return projects, total, nil
}

// FindByID 按 project_id 查找
func (r *ProjectRepo) FindByID(projectID string) (*entity.Project, error) {
	query := `SELECT project_id, name, base_url, description, status, created_at, updated_at
FROM projects WHERE project_id = ?`

	var m model.ProjectModel
	if err := r.db.Get(&m, query, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p := m.ToEntity()
	return &p, nil
}

// Create 新建项目
func (r *ProjectRepo) Create(project *entity.Project) error {
	query := `INSERT INTO projects (project_id, name, base_url, description, status)
VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, project.ProjectID, project.Name, project.BaseURL, project.Description, project.Status)
	return err
}

// Update 更新项目
func (r *ProjectRepo) Update(project *entity.Project) error {
	query := `UPDATE projects SET name=?, base_url=?, description=?, status=? WHERE project_id=?`
	_, err := r.db.Exec(query, project.Name, project.BaseURL, project.Description, project.Status, project.ProjectID)
	return err
}

// Delete 删除项目
func (r *ProjectRepo) Delete(projectID string) error {
	_, err := r.db.Exec("DELETE FROM projects WHERE project_id=?", projectID)
	return err
}

// Count 总数
func (r *ProjectRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM projects")
	return count, err
}
