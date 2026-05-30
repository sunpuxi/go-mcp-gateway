package repository

import "github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"

// ProjectRepository 项目管理仓储接口（管理后台 CRUD）
type ProjectRepository interface {
	// List 分页查询所有项目
	List(page, size int) ([]entity.Project, int, error)

	// FindByID 按 project_id 查找
	FindByID(projectID string) (*entity.Project, error)

	// Create 新建项目
	Create(project *entity.Project) error

	// Update 更新项目（按 project_id）
	Update(project *entity.Project) error

	// Delete 删除项目
	Delete(projectID string) error

	// Count 总数
	Count() (int, error)
}
