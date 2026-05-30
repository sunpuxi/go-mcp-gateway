package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
)

// writeJSON 统一 JSON 响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Admin JSON 编码失败", "error", err)
	}
}

// parsePaginate 解析分页参数
func parsePaginate(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}
