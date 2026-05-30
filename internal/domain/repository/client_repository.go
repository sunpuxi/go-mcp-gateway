package repository

import (
	"context"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// ClientRepository 是客户端查询接口，便于测试时 mock
type ClientRepository interface {

	// FindByAPIKeyHash 根据 API Key 的 SHA256 哈希查找客户端
	FindByAPIKeyHash(ctx context.Context, hash string) (*entity.Client, error)

	// FindPermissions 查询客户端有权调用的工具名称列表
	FindPermissions(ctx context.Context, clientID string) ([]string, error)
}
