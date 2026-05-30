package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
)

// Handler 管理后台 HTTP 处理器
type Handler struct {
	adminService *appservice.AdminService
}

// NewHandler 创建管理后台 Handler
func NewHandler(adminService *appservice.AdminService) *Handler {
	return &Handler{adminService: adminService}
}

// RegisterRoutes 注册所有管理后台路由到 chi.Router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/admin/api", func(r chi.Router) {
		// 仪表盘
		r.Get("/stats", h.GetStats)

		// 项目管理
		r.Get("/projects", h.ListProjects)
		r.Post("/projects", h.CreateProject)
		r.Put("/projects/{id}", h.UpdateProject)
		r.Delete("/projects/{id}", h.DeleteProject)

		// 工具管理
		r.Get("/tools", h.ListTools)
		r.Post("/tools", h.CreateTool)
		r.Put("/tools/{id}", h.UpdateTool)
		r.Delete("/tools/{id}", h.DeleteTool)

		// 客户端管理
		r.Get("/clients", h.ListClients)
		r.Post("/clients", h.CreateClient)
		r.Post("/clients/{id}/api-key", h.GenerateAPIKey)
		r.Put("/clients/{id}", h.UpdateClient)
		r.Delete("/clients/{id}", h.DeleteClient)

		// 权限管理
		r.Get("/clients/{id}/permissions", h.GetClientPermissions)
		r.Put("/clients/{id}/permissions", h.UpdateClientPermissions)
	})
}

// ======================== 辅助函数 ========================

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[admin] JSON 编码失败: %v", err)
	}
}

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

// ======================== 仪表盘 ========================

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ======================== 项目 ========================

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	page, size := parsePaginate(r)
	data, total, err := h.adminService.ListProjects(page, size)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  data,
		"total": total,
	})
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var dto appservice.ProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if dto.ProjectID == "" || dto.Name == "" || dto.BaseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project_id、name、base_url 为必填字段"})
		return
	}
	result, err := h.adminService.CreateProject(dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dto appservice.ProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	result, err := h.adminService.UpdateProject(id, dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.adminService.DeleteProject(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ======================== 工具 ========================

func (h *Handler) ListTools(w http.ResponseWriter, r *http.Request) {
	page, size := parsePaginate(r)
	data, total, err := h.adminService.ListTools(page, size)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  data,
		"total": total,
	})
}

func (h *Handler) CreateTool(w http.ResponseWriter, r *http.Request) {
	var dto appservice.ToolDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if dto.Name == "" || dto.Title == "" || dto.URLTemplate == "" || dto.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name、title、url_template、project_id 为必填字段"})
		return
	}
	result, err := h.adminService.CreateTool(dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateTool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 tool_id"})
		return
	}
	var dto appservice.ToolDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	result, err := h.adminService.UpdateTool(id, dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteTool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 tool_id"})
		return
	}
	if err := h.adminService.DeleteTool(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ======================== 客户端 ========================

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

// ======================== 权限 ========================

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
