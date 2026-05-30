package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
)

// Handler 管理后台 HTTP 处理器（入口）
type Handler struct {
	adminService *appservice.AdminService
}

// NewHandler 创建管理后台 Handler
func NewHandler(adminService *appservice.AdminService) *Handler {
	return &Handler{adminService: adminService}
}

// ======================== 仪表盘 ========================

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ======================== 客户端 CRUD ========================

func (h *Handler) ListClients(w http.ResponseWriter, r *http.Request) {
	page, size := parsePaginate(r)
	data, total, err := h.adminService.ListClients(page, size)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  data,
		"total": total,
	})
}

func (h *Handler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var dto appservice.ClientDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if dto.ClientID == "" || dto.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client_id、name 为必填字段"})
		return
	}
	result, err := h.adminService.CreateClient(dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	apiKey, err := h.adminService.GenerateAPIKey(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"api_key": apiKey})
}

func (h *Handler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dto appservice.ClientDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	result, err := h.adminService.UpdateClient(id, dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.adminService.DeleteClient(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ======================== 权限管理 ========================

func (h *Handler) GetClientPermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	toolIDs, err := h.adminService.GetClientPermissions(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if toolIDs == nil {
		toolIDs = []int64{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tool_ids": toolIDs})
}

func (h *Handler) UpdateClientPermissions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		ToolIDs []int64 `json:"tool_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if req.ToolIDs == nil {
		req.ToolIDs = []int64{}
	}
	if err := h.adminService.UpdateClientPermissions(id, req.ToolIDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "权限保存成功"})
}
