package model

import (
	"time"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// ClientModel 对应数据库 clients 表结构
type ClientModel struct {
	ClientID     string    `db:"client_id"`
	Name         string    `db:"name"`
	APIKeyHash   string    `db:"api_key_hash"`
	APIKeyPrefix string    `db:"api_key_prefix"`
	Description  string    `db:"description"`
	Status       int       `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ToEntity 将数据库模型转换为领域实体
func (m *ClientModel) ToEntity() entity.Client {
	return entity.Client{
		ClientID:     m.ClientID,
		Name:         m.Name,
		APIKeyHash:   m.APIKeyHash,
		APIKeyPrefix: m.APIKeyPrefix,
		Description:  m.Description,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
