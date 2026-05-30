package repository

import "github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"

// ToolQuerier 是工具查询接口，供 MCP 运行时使用
type ToolQuerier interface {
	ListEnabled() ([]entity.Tool, error)
	FindByName(name string) (*entity.Tool, error)
}

// ToolAdminRepository 工具管理仓储接口（管理后台 CRUD）
type ToolAdminRepository interface {
	ToolQuerier

	// ListAll 分页查询所有工具（含禁用的）
	ListAll(page, size int) ([]entity.Tool, int, error)

	// FindByID 按 tool_id 查找
	FindByID(toolID int64) (*entity.Tool, error)

	// Create 新建工具
	Create(tool *entity.Tool) error

	// Update 更新工具
	Update(tool *entity.Tool) error

	// Delete 删除工具
	Delete(toolID int64) error

	// Count 总数
	Count() (int, error)
}
