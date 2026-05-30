package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
)

// ToolHandler 工具相关 HTTP 处理器
type ToolHandler struct {
	adminService *appservice.AdminService
}

func (th *ToolHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	page, size := parsePaginate(r)
	data, total, err := th.adminService.ListTools(page, size)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  data,
		"total": total,
	})
}

func (th *ToolHandler) CreateTool(w http.ResponseWriter, r *http.Request) {
	var dto appservice.ToolDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if dto.Name == "" || dto.Title == "" || dto.URLTemplate == "" || dto.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name、title、url_template、project_id 为必填字段"})
		return
	}
	result, err := th.adminService.CreateTool(dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (th *ToolHandler) UpdateTool(w http.ResponseWriter, r *http.Request) {
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
	result, err := th.adminService.UpdateTool(id, dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (th *ToolHandler) DeleteTool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 tool_id"})
		return
	}
	if err := th.adminService.DeleteTool(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}
