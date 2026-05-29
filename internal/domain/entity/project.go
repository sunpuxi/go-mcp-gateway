package entity

import "time"

// Project 表示下游 HTTP 服务
type Project struct {
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"base_url"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
