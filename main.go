package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sunpuxi/go-mcp-gateway/config"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
	dbpkg "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db"
	infrahttp "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/http"
	adminhttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/admin"
	mcphttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/mcp"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db := dbpkg.NewMySQL(cfg.Database.DSN)
	defer db.Close()

	// 创建 Session 管理器（30 分钟过期）
	sessionManager := domainservice.NewSessionManager(30 * time.Minute)

	// 创建工具仓储（同时满足 ToolQuerier + ToolAdminRepository）
	toolRepo := dbpkg.NewToolRepo(db)

	// 创建客户端仓储（同时满足运行时查询 + 管理 CRUD）
	clientRepo := dbpkg.NewClientRepo(db)

	// 创建项目仓储
	projectRepo := dbpkg.NewProjectRepo(db)

	// 创建 HTTP 客户端
	httpClient := infrahttp.NewHTTPClient()

	// 创建鉴权服务
	authService := domainservice.NewAuthService(clientRepo)

	// 创建 MCP 应用层服务
	mcpSvc := appservice.NewMCPService(toolRepo, httpClient)

	// 创建 MCP Handler（传输层）
	mcpHandler := mcphttp.NewHandler(sessionManager, mcpSvc, authService)

	// 创建 Admin 应用层服务
	adminSvc := appservice.NewAdminService(projectRepo, toolRepo, clientRepo, sessionManager)

	// 创建 Admin Handler
	adminHandler := adminhttp.NewHandler(adminSvc)

	// 创建路由
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// === MCP SSE 传输端点（对外开放，使用客户端 API Key 鉴权）===
	r.Get("/sse", mcpHandler.HandleSSE)
	r.Post("/messages", mcpHandler.HandleMessage)

	// === Admin 管理后台（使用 admin_api_key 鉴权）===
	r.Group(func(r chi.Router) {
		r.Use(adminhttp.AdminAuthMiddleware(cfg.Admin.APIKey))
		adminHandler.RegisterRoutes(r)
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("MCP Gateway 启动，监听 %s", addr)
	log.Printf("Admin API 路由已注册 (api_key: %s)", cfg.Admin.APIKey)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
