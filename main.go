package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sunpuxi/go-mcp-gateway/internal/config"
	dbpkg "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/db"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain"
	mcphttp "github.com/sunpuxi/go-mcp-gateway/internal/interface/mcp"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db := dbpkg.NewMySQL(cfg.Database.DSN)
	defer db.Close()

	// 创建 Session 管理器（30 分钟过期）
	sessionManager := domain.NewSessionManager(30 * time.Minute)

	// 创建 Tool Repository
	toolRepo := dbpkg.NewToolRepo(db)

	// 创建 HTTP 客户端
	httpClient := domain.NewHTTPClient()

	// 创建 MCP Handler
	mcpHandler := mcphttp.NewHandler(sessionManager, toolRepo, httpClient)

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
