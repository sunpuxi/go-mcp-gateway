package admin

import "github.com/go-chi/chi/v5"

// RegisterRoutes 注册所有管理后台路由到 chi.Router
func (h *Handler) RegisterRoutes(r chi.Router) {
	ph := &ProjectHandler{adminService: h.adminService}
	th := &ToolHandler{adminService: h.adminService}

	r.Route("/admin/api", func(r chi.Router) {
		// 仪表盘
		r.Get("/stats", h.GetStats)

		// 项目管理
		r.Get("/projects", ph.ListProjects)
		r.Post("/projects", ph.CreateProject)
		r.Put("/projects/{id}", ph.UpdateProject)
		r.Delete("/projects/{id}", ph.DeleteProject)

		// 工具管理
		r.Get("/tools", th.ListTools)
		r.Post("/tools", th.CreateTool)
		r.Put("/tools/{id}", th.UpdateTool)
		r.Delete("/tools/{id}", th.DeleteTool)

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
