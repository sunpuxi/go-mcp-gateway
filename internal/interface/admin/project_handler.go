package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
)

// ProjectHandler 项目相关 HTTP 处理器
type ProjectHandler struct {
	adminService *appservice.AdminService
}

func (ph *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	page, size := parsePaginate(r)
	data, total, err := ph.adminService.ListProjects(page, size)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  data,
		"total": total,
	})
}

func (ph *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var dto appservice.ProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	if dto.ProjectID == "" || dto.Name == "" || dto.BaseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project_id、name、base_url 为必填字段"})
		return
	}
	result, err := ph.adminService.CreateProject(dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (ph *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dto appservice.ProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON 解析失败: " + err.Error()})
		return
	}
	result, err := ph.adminService.UpdateProject(id, dto)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (ph *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := ph.adminService.DeleteProject(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}
