package repository

import "github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"

// ToolQuerier 是工具查询接口，便于测试时 mock
type ToolQuerier interface {
	// ListEnabled 查询所有启用的工具
	ListEnabled() ([]entity.Tool, error)

	// FindByName 根据工具名查询
	FindByName(name string) (*entity.Tool, error)
}
