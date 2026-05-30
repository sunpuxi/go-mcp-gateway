package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

var _ repository.ClientRepository = (*ClientRepo)(nil)

// ClientRepo 是 ClientRepository 的 MySQL 实现
type ClientRepo struct {
	db *sqlx.DB
}

// NewClientRepo 创建 ClientRepo
func NewClientRepo(db *sqlx.DB) *ClientRepo {
	return &ClientRepo{db: db}
}

// ======================== 运行时查询 ========================

// FindByAPIKeyHash 根据 API Key 的 SHA256 哈希查找客户端
func (r *ClientRepo) FindByAPIKeyHash(ctx context.Context, hash string) (*entity.Client, error) {
	query := `SELECT client_id, name, api_key_hash, api_key_prefix, description, status, created_at, updated_at
FROM clients WHERE api_key_hash = ? LIMIT 1`

	var m model.ClientModel
	if err := r.db.GetContext(ctx, &m, query, hash); err != nil {
		return nil, err
	}
	client := m.ToEntity()
	return &client, nil
}

// FindPermissions 查询客户端有权调用的工具名称列表
func (r *ClientRepo) FindPermissions(ctx context.Context, clientID string) ([]string, error) {
	query := `SELECT t.name FROM client_tool_permissions p
JOIN tools t ON p.tool_id = t.tool_id WHERE p.client_id = ?`
	var names []string
	if err := r.db.SelectContext(ctx, &names, query, clientID); err != nil {
		return nil, err
	}
	return names, nil
}

// ======================== 管理后台 CRUD ========================

// ListAll 分页查询所有客户端，toolCounts 与 clients 一一对应
func (r *ClientRepo) ListAll(page, size int) ([]entity.Client, []int, int, error) {
	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM clients"); err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT client_id, name, api_key_hash, api_key_prefix, description, status, created_at, updated_at
FROM clients ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var models []model.ClientModel
	if err := r.db.Select(&models, query, size, offset); err != nil {
		return nil, nil, 0, err
	}

	clients := make([]entity.Client, len(models))
	clientIDs := make([]string, len(models))
	for i, m := range models {
		clients[i] = m.ToEntity()
		clientIDs[i] = m.ClientID
	}

	// 批量查询每个客户端的已授权工具数
	toolCounts := make([]int, len(clients))
	toolCountMap := make(map[string]int)
	type row struct {
		ClientID string `db:"client_id"`
		Count    int    `db:"cnt"`
	}
	if len(clientIDs) > 0 {
		q, args, _ := sqlx.In(
			"SELECT client_id, COUNT(*) AS cnt FROM client_tool_permissions WHERE client_id IN (?) GROUP BY client_id",
			clientIDs,
		)
		var rows []row
		if err := r.db.Select(&rows, q, args...); err == nil {
			for _, rr := range rows {
				toolCountMap[rr.ClientID] = rr.Count
			}
		}
	}
	for i, c := range clients {
		toolCounts[i] = toolCountMap[c.ClientID]
	}

	return clients, toolCounts, total, nil
}

// FindByID 按 client_id 查找
func (r *ClientRepo) FindByID(ctx context.Context, clientID string) (*entity.Client, error) {
	query := `SELECT client_id, name, api_key_hash, api_key_prefix, description, status, created_at, updated_at
FROM clients WHERE client_id = ?`
	var m model.ClientModel
	if err := r.db.GetContext(ctx, &m, query, clientID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	c := m.ToEntity()
	return &c, nil
}

// Create 新建客户端
func (r *ClientRepo) Create(ctx context.Context, client *entity.Client) error {
	query := `INSERT INTO clients (client_id, name, api_key_hash, api_key_prefix, description, status)
VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		client.ClientID, client.Name, client.APIKeyHash, client.APIKeyPrefix,
		client.Description, client.Status,
	)
	return err
}

// Update 更新客户端信息
func (r *ClientRepo) Update(ctx context.Context, client *entity.Client) error {
	query := `UPDATE clients SET name=?, description=?, status=? WHERE client_id=?`
	_, err := r.db.ExecContext(ctx, query,
		client.Name, client.Description, client.Status, client.ClientID,
	)
	return err
}

// UpdateAPIKey 更新客户端的 API Key（hash + prefix）
func (r *ClientRepo) UpdateAPIKey(ctx context.Context, clientID, apiKeyHash, apiKeyPrefix string) error {
	query := `UPDATE clients SET api_key_hash=?, api_key_prefix=? WHERE client_id=?`
	_, err := r.db.ExecContext(ctx, query, apiKeyHash, apiKeyPrefix, clientID)
	return err
}

// Delete 删除客户端
func (r *ClientRepo) Delete(ctx context.Context, clientID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM clients WHERE client_id=?", clientID)
	return err
}

// Count 总数
func (r *ClientRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM clients")
	return count, err
}

// ======================== 权限管理 ========================

// FindToolPermissions 查询客户端有权调用的 tool_id 列表
func (r *ClientRepo) FindToolPermissions(ctx context.Context, clientID string) ([]int64, error) {
	query := `SELECT tool_id FROM client_tool_permissions WHERE client_id = ?`
	var toolIDs []int64
	if err := r.db.SelectContext(ctx, &toolIDs, query, clientID); err != nil {
		return nil, err
	}
	return toolIDs, nil
}

// SavePermissions 全量替换客户端权限（先删后插）
func (r *ClientRepo) SavePermissions(ctx context.Context, clientID string, toolIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除旧权限
	if _, err := tx.ExecContext(ctx, "DELETE FROM client_tool_permissions WHERE client_id=?", clientID); err != nil {
		return err
	}

	// 批量插入新权限
	if len(toolIDs) > 0 {
		stmt := `INSERT INTO client_tool_permissions (client_id, tool_id) VALUES (?, ?)`
		for _, tid := range toolIDs {
			if _, err := tx.ExecContext(ctx, stmt, clientID, tid); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
