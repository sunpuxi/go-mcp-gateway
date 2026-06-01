package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/didi/gendry/builder"
	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

var _ repository.ProjectRepository = (*ProjectRepo)(nil)

// projectSelectFields 是 projects 表查询的共享字段列表
var projectSelectFields = []string{
	"project_id", "name", "base_url", "description",
	"status", "created_at", "updated_at",
}

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
	res, err := builder.AggregateQuery(context.Background(), r.db.DB, "projects", nil, builder.AggregateCount("*"))
	if err != nil {
		return nil, 0, err
	}
	total := int(res.Int64())

	offset := (page - 1) * size
	where := map[string]interface{}{
		"_orderby": "created_at DESC",
		"_limit":   []uint{uint(offset), uint(size)},
	}
	sqlStr, args, err := builder.BuildSelect("projects", where, projectSelectFields)
	if err != nil {
		return nil, 0, err
	}

	var models []model.ProjectModel
	if err := r.db.Select(&models, sqlStr, args...); err != nil {
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
	where := map[string]interface{}{"project_id": projectID}
	sqlStr, args, err := builder.BuildSelect("projects", where, projectSelectFields)
	if err != nil {
		return nil, err
	}

	var m model.ProjectModel
	if err := r.db.Get(&m, sqlStr, args...); err != nil {
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
	data := []map[string]interface{}{{
		"project_id":  project.ProjectID,
		"name":        project.Name,
		"base_url":    project.BaseURL,
		"description": project.Description,
		"status":      project.Status,
	}}
	sqlStr, args, err := builder.BuildInsert("projects", data)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// Update 更新项目
func (r *ProjectRepo) Update(project *entity.Project) error {
	where := map[string]interface{}{"project_id": project.ProjectID}
	update := map[string]interface{}{
		"name":        project.Name,
		"base_url":    project.BaseURL,
		"description": project.Description,
		"status":      project.Status,
	}
	sqlStr, args, err := builder.BuildUpdate("projects", where, update)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// Delete 删除项目
func (r *ProjectRepo) Delete(projectID string) error {
	where := map[string]interface{}{"project_id": projectID}
	sqlStr, args, err := builder.BuildDelete("projects", where)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sqlStr, args...)
	return err
}

// Count 总数
func (r *ProjectRepo) Count() (int, error) {
	res, err := builder.AggregateQuery(context.Background(), r.db.DB, "projects", nil, builder.AggregateCount("*"))
	if err != nil {
		return 0, err
	}
	return int(res.Int64()), nil
}
