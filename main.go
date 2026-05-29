package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
	"github.com/sunpuxi/go-mcp-gateway/config"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
	infrahttp "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/http"
	dbpkg "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db"
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

	// 创建 Tool Repository
	toolRepo := dbpkg.NewToolRepo(db)

	// 创建 HTTP 客户端
	httpClient := infrahttp.NewHTTPClient()

	// 创建 MCP 应用层服务
	mcpSvc := appservice.NewMCPService(toolRepo, httpClient)

	// 创建 MCP Handler（传输层）
	mcpHandler := mcphttp.NewHandler(sessionManager, mcpSvc)

	// 创建路由
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// MCP SSE 传输端点
	r.Get("/sse", mcpHandler.HandleSSE)             // SSE 流（服务端推送通道）
	r.Post("/messages", mcpHandler.HandleMessage)   // 消息接收（客户端请求）

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("MCP Gateway 启动，监听 %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
