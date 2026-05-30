package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// Handler 健康检查 Handler
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建健康检查 Handler
func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

// healthResponse 健康检查响应
type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// readyResponse 就绪检查响应
type readyResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Health 存活检查 —— 进程存活即可
// GET /health
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready 就绪检查 —— 检查所有依赖是否就绪
// GET /ready
func (h *Handler) Ready(w http.ResponseWriter, _ *http.Request) {
	checks := make(map[string]string)
	allOk := true

	// 检查数据库连接
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			checks["database"] = "error: " + err.Error()
			allOk = false
		} else {
			checks["database"] = "ok"
		}
	}

	// 后续可扩展：检查 Redis、Kafka 等其他依赖

	status := "ok"
	httpStatus := http.StatusOK
	if !allOk {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	writeJSON(w, httpStatus, readyResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	})
}

// writeJSON 写 JSON 响应
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"status":"error"}`, http.StatusInternalServerError)
	}
}
