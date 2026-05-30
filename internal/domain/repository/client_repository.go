package repository

import (
	"context"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// ClientRepository 客户端查询 + 管理接口
type ClientRepository interface {
	// --- 运行时查询（MCP 鉴权用）---

	FindByAPIKeyHash(ctx context.Context, hash string) (*entity.Client, error)
	FindPermissions(ctx context.Context, clientID string) ([]string, error)

	// --- 管理后台 CRUD ---

	// ListAll 分页查询所有客户端，返回列表 + 每个客户端的已授权工具数
	ListAll(page, size int) ([]entity.Client, []int, int, error)

	// FindByID 按 client_id 查找
	FindByID(ctx context.Context, clientID string) (*entity.Client, error)

	// Create 新建客户端（api_key_hash/api_key_prefix 由调用方生成）
	Create(ctx context.Context, client *entity.Client) error

	// Update 更新客户端信息
	Update(ctx context.Context, client *entity.Client) error

	// UpdateAPIKey 更新客户端的 API Key（hash + prefix）
	UpdateAPIKey(ctx context.Context, clientID, apiKeyHash, apiKeyPrefix string) error

	// Delete 删除客户端
	Delete(ctx context.Context, clientID string) error

	// Count 总数
	Count(ctx context.Context) (int, error)

	// --- 权限管理 ---

	// FindToolPermissions 查询客户端有权调用的 tool_id 列表
	FindToolPermissions(ctx context.Context, clientID string) ([]int64, error)

	// SavePermissions 全量替换客户端权限（先删后插）
	SavePermissions(ctx context.Context, clientID string, toolIDs []int64) error

	// DeletePermissionsByToolIDs 按工具ID列表批量删除权限
	DeletePermissionsByToolIDs(ctx context.Context, toolIDs []int64) error
}
