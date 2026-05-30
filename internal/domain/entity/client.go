package entity

import "time"

// Client 表示一个 API 客户端（如大数据项目组）
type Client struct {
	ClientID     string    `json:"client_id"`
	Name         string    `json:"name"`
	APIKeyHash   string    `json:"-"`
	APIKeyPrefix string    `json:"api_key_prefix"`
	Description  string    `json:"description"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
