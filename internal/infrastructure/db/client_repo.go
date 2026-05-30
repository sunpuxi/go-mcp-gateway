package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

// 编译时确保 ClientRepo 实现 ClientRepository 接口
var _ repository.ClientRepository = (*ClientRepo)(nil)

// ClientRepo 是 ClientRepository 的 MySQL 实现
type ClientRepo struct {
	db *sqlx.DB
}

// NewClientRepo 创建 ClientRepo
func NewClientRepo(db *sqlx.DB) *ClientRepo {
	return &ClientRepo{db: db}
}

// FindByAPIKeyHash 根据 API Key 的 SHA256 哈希查找客户端
func (r *ClientRepo) FindByAPIKeyHash(ctx context.Context, hash string) (*entity.Client, error) {
	query := `SELECT client_id, name, api_key_hash, api_key_prefix, description, status, created_at, updated_at
FROM clients
WHERE api_key_hash = ?
LIMIT 1`

	var m model.ClientModel
	if err := r.db.GetContext(ctx, &m, query, hash); err != nil {
		return nil, err
	}

	client := m.ToEntity()
	return &client, nil
}

// FindPermissions 查询客户端有权调用的工具名称列表
func (r *ClientRepo) FindPermissions(ctx context.Context, clientID string) ([]string, error) {
	query := `SELECT t.name
FROM client_tool_permissions p
JOIN tools t ON p.tool_id = t.tool_id
WHERE p.client_id = ?`

	var names []string
	if err := r.db.SelectContext(ctx, &names, query, clientID); err != nil {
		return nil, err
	}
	return names, nil
}
