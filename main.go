package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sunpuxi/go-mcp-gateway/config"
	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
	sessionpkg "github.com/sunpuxi/go-mcp-gateway/internal/application/session"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/circuitbreaker"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
	dbpkg "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db"
	infrahttp "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/http"
	adminhttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/admin"
	healthhttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/health"
	mcphttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/mcp"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志器（必须在其他组件之前）
	if err := logger.Init(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志器失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化数据库
	db, err := dbpkg.NewMySQL(cfg.Database.DSN)
	if err != nil {
		logger.Fatal("数据库连接失败", "error", err)
	}
	defer db.Close()

	// 创建 Session 管理器（30 分钟过期）
	sessionManager := sessionpkg.NewManager(30 * time.Minute)

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

	// 创建熔断器注册表（按 Project 维度，使用默认配置）
	cbRegistry := circuitbreaker.NewRegistry(circuitbreaker.DefaultConfig())

	// 创建 MCP 应用层服务
	mcpSvc := appservice.NewMCPService(toolRepo, httpClient, cbRegistry)

	// 创建 MCP Handler（传输层）
	mcpHandler := mcphttp.NewHandler(sessionManager, mcpSvc, authService)

	// 创建 Admin 应用层服务
	adminSvc := appservice.NewAdminService(projectRepo, toolRepo, clientRepo, sessionManager)

	// 创建 Admin Handler
	adminHandler := adminhttp.NewHandler(adminSvc)

	// 创建健康检查 Handler
	healthHandler := healthhttp.NewHandler(db)

	// 创建路由
	r := chi.NewRouter()
	r.Use(logger.ChiMiddleware) // 结构化请求日志 + trace_id
	r.Use(middleware.Recoverer)

	// === 健康检查（无需鉴权，供 K8s/Docker/负载均衡使用）===
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

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
	logger.Info("MCP Gateway 启动", "addr", addr)
	logger.Info("Admin API 路由已注册",
		"api_key_prefix", cfg.Admin.APIKey[:min(len(cfg.Admin.APIKey), 8)]+"...")
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal("服务启动失败", "error", err)
	}
}
