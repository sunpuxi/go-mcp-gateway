package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/didi/gendry/builder"
	"github.com/jmoiron/sqlx"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db/model"
)

var _ repository.ClientRepository = (*ClientRepo)(nil)

// clientSelectFields 是 clients 表查询的共享字段列表
var clientSelectFields = []string{
	"client_id", "name", "api_key_hash", "api_key_prefix",
	"description", "status", "created_at", "updated_at",
}

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
	where := map[string]interface{}{
		"api_key_hash": hash,
		"_limit":       []uint{0, 1},
	}
	sqlStr, args, err := builder.BuildSelect("clients", where, clientSelectFields)
	if err != nil {
		return nil, err
	}

	var m model.ClientModel
	if err := r.db.GetContext(ctx, &m, sqlStr, args...); err != nil {
		return nil, err
	}
	client := m.ToEntity()
	return &client, nil
}

// FindPermissions 查询客户端有权调用的工具名称列表
func (r *ClientRepo) FindPermissions(ctx context.Context, clientID string) ([]string, error) {
	template := `SELECT t.name FROM client_tool_permissions p
	JOIN tools t ON p.tool_id = t.tool_id WHERE p.client_id = {{client_id}}`
	params := map[string]interface{}{"client_id": clientID}
	sqlStr, args, err := builder.NamedQuery(template, params)
	if err != nil {
		return nil, err
	}
	var names []string
	if err := r.db.SelectContext(ctx, &names, sqlStr, args...); err != nil {
		return nil, err
	}
	return names, nil
}

// ======================== 管理后台 CRUD ========================

// ListAll 分页查询所有客户端，toolCounts 与 clients 一一对应
func (r *ClientRepo) ListAll(page, size int) ([]entity.Client, []int, int, error) {
	res, err := builder.AggregateQuery(context.Background(), r.db.DB, "clients", nil, builder.AggregateCount("*"))
	if err != nil {
		return nil, nil, 0, err
	}
	total := int(res.Int64())

	offset := (page - 1) * size
	where := map[string]interface{}{
		"_orderby": "created_at DESC",
		"_limit":   []uint{uint(offset), uint(size)},
	}
	sqlStr, args, err := builder.BuildSelect("clients", where, clientSelectFields)
	if err != nil {
		return nil, nil, 0, err
	}

	var models []model.ClientModel
	if err := r.db.Select(&models, sqlStr, args...); err != nil {
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
		// 使用 NamedQuery 动态生成 IN 子句的命名参数，避免 sqlx.In
		paramKeys := make([]string, len(clientIDs))
		params := make(map[string]interface{}, len(clientIDs))
		for i, id := range clientIDs {
			key := fmt.Sprintf("id%d", i)
			paramKeys[i] = fmt.Sprintf("{{%s}}", key)
			params[key] = id
		}
		template := fmt.Sprintf(
			"SELECT client_id, COUNT(*) AS cnt FROM client_tool_permissions WHERE client_id IN (%s) GROUP BY client_id",
			strings.Join(paramKeys, ","),
		)
		q, qArgs, buildErr := builder.NamedQuery(template, params)
		if buildErr == nil {
			var rows []row
			if err := r.db.Select(&rows, q, qArgs...); err == nil {
				for _, rr := range rows {
					toolCountMap[rr.ClientID] = rr.Count
				}
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
	where := map[string]interface{}{"client_id": clientID}
	sqlStr, args, err := builder.BuildSelect("clients", where, clientSelectFields)
	if err != nil {
		return nil, err
	}

	var m model.ClientModel
	if err := r.db.GetContext(ctx, &m, sqlStr, args...); err != nil {
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
	data := []map[string]interface{}{{
		"client_id":      client.ClientID,
		"name":           client.Name,
		"api_key_hash":   client.APIKeyHash,
		"api_key_prefix": client.APIKeyPrefix,
		"description":    client.Description,
		"status":         client.Status,
	}}
	sqlStr, args, err := builder.BuildInsert("clients", data)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}

// Update 更新客户端信息
func (r *ClientRepo) Update(ctx context.Context, client *entity.Client) error {
	where := map[string]interface{}{"client_id": client.ClientID}
	update := map[string]interface{}{
		"name":        client.Name,
		"description": client.Description,
		"status":      client.Status,
	}
	sqlStr, args, err := builder.BuildUpdate("clients", where, update)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}

// UpdateAPIKey 更新客户端的 API Key（hash + prefix）
func (r *ClientRepo) UpdateAPIKey(ctx context.Context, clientID, apiKeyHash, apiKeyPrefix string) error {
	where := map[string]interface{}{"client_id": clientID}
	update := map[string]interface{}{
		"api_key_hash":   apiKeyHash,
		"api_key_prefix": apiKeyPrefix,
	}
	sqlStr, args, err := builder.BuildUpdate("clients", where, update)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}

// Delete 删除客户端
func (r *ClientRepo) Delete(ctx context.Context, clientID string) error {
	where := map[string]interface{}{"client_id": clientID}
	sqlStr, args, err := builder.BuildDelete("clients", where)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}

// Count 总数
func (r *ClientRepo) Count(ctx context.Context) (int, error) {
	res, err := builder.AggregateQuery(ctx, r.db.DB, "clients", nil, builder.AggregateCount("*"))
	if err != nil {
		return 0, err
	}
	return int(res.Int64()), nil
}

// ======================== 权限管理 ========================

// FindToolPermissions 查询客户端有权调用的 tool_id 列表
func (r *ClientRepo) FindToolPermissions(ctx context.Context, clientID string) ([]int64, error) {
	where := map[string]interface{}{"client_id": clientID}
	sqlStr, args, err := builder.BuildSelect("client_tool_permissions", where, []string{"tool_id"})
	if err != nil {
		return nil, err
	}
	var toolIDs []int64
	if err := r.db.SelectContext(ctx, &toolIDs, sqlStr, args...); err != nil {
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
	delWhere := map[string]interface{}{"client_id": clientID}
	delSQL, delArgs, err := builder.BuildDelete("client_tool_permissions", delWhere)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, delSQL, delArgs...); err != nil {
		return err
	}

	// 批量插入新权限（单次多行 INSERT，替换逐行循环）
	if len(toolIDs) > 0 {
		data := make([]map[string]interface{}, len(toolIDs))
		for i, tid := range toolIDs {
			data[i] = map[string]interface{}{
				"client_id": clientID,
				"tool_id":   tid,
			}
		}
		insSQL, insArgs, err := builder.BuildInsert("client_tool_permissions", data)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, insSQL, insArgs...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeletePermissionsByToolIDs 按工具ID列表批量删除权限
func (r *ClientRepo) DeletePermissionsByToolIDs(ctx context.Context, toolIDs []int64) error {
	if len(toolIDs) == 0 {
		return nil
	}
	// 使用 Gendry 原生 IN 操作符，无需 sqlx.In + Rebind
	inArgs := make([]interface{}, len(toolIDs))
	for i, id := range toolIDs {
		inArgs[i] = id
	}
	where := map[string]interface{}{"tool_id IN": inArgs}
	sqlStr, args, err := builder.BuildDelete("client_tool_permissions", where)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return err
}
