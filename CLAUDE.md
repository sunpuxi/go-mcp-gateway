# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

MCP Gateway 是一个 MCP（Model Context Protocol）协议转换网关，将 AI Agent 的 MCP 工具调用（JSON-RPC over SSE）翻译为下游 HTTP RESTful 请求。项目采用 Go 1.25 后端 + React 18/Ant Design 5 前端管理后台。

## 常用命令

```bash
# 构建
go build -o mcp-gateway.exe .

# 运行（需要先启动 MySQL 并导入 doc/sql/*.sql）
./mcp-gateway

# 运行测试
go test ./...

# 运行单个包的测试
go test ./internal/domain/circuitbreaker/
go test ./internal/application/service/

# 运行单个测试函数
go test ./internal/domain/circuitbreaker/ -run TestCircuitBreaker_StateTransitions -v

# 前端开发
cd web && npm install && npm run dev

# Docker Compose 全栈启动（MySQL + 后端 + 前端）
docker-compose up -d

# Go 依赖管理
go mod tidy
go mod download
```

## 架构分层

项目采用 **DDD 分层架构**，通过 `main.go` 手动依赖注入组装：

```
interface（传输层） → application（用例/应用层） → domain（领域层）
                               ↓
                      infrastructure（基础设施层）
```

### 分层职责

| 层 | 目录 | 职责 |
|---|---|---|
| **interface** | `internal/interface/` | HTTP Handler：MCP SSE 传输（`/sse`, `/messages`）、Admin REST API、健康检查 |
| **application** | `internal/application/` | 用例编排：MCPService（工具调用链路）、AdminService（CRUD）、SessionManager、重试逻辑 |
| **domain** | `internal/domain/` | 核心业务逻辑：实体、仓储接口、参数映射引擎、熔断器、鉴权服务、Schema 构建 |
| **infrastructure** | `internal/infrastructure/` | 技术实现：MySQL（sqlx）、Redis 限流器、HTTP 客户端、熔断器落地 |
| **pkg** | `pkg/` | 共享工具包：JSON-RPC 2.0 类型、MCP 协议消息、结构化日志（zap） |

### 关键依赖方向

- `domain/repository/` 定义仓储 **接口**（如 `ToolQuerier`、`ClientRepository`），`infrastructure/db/` 提供实现
- `application/service/` 依赖 domain 接口和 entity，不依赖 infrastructure 具体实现
- `main.go` 是唯一组装点，创建具体实现并注入

## 核心流程

### MCP 工具调用链路（tools/call）

`GET /sse` 建立 SSE 长连接 → `POST /messages` 接收 JSON-RPC 请求 → 鉴权（Bearer Token + SHA256 hash 匹配）→ 查数据库加载工具定义 → 参数映射（path/query/body/header）→ 限流检查（Redis 滑动窗口）→ 熔断检查（按 Project 维度）→ 转发 HTTP 到下游 → 重试（指数退避 + jitter）→ 结果写回 SSE 通道

### 参数映射引擎

`domain/mapper/param_mapper.go` 是核心转换逻辑。每个 Tool 有一条 `params` JSON 字段（`[]ParamRule`），定义 MCP 参数如何映射到 HTTP 请求的 path（`{var}` 模板替换）、query、body、header。必填参数缺失会报错，可选参数支持默认值。

### 鉴权模型

- **MCP 端点**：客户端在 `Authorization: Bearer <apiKey>` 中传入 API Key，网关 SHA256 后与 `clients.api_key_hash` 匹配。每个 Client 通过 `client_tool_permissions` 表控制可调用的 Tool 列表。
- **Admin 端点**：通过 `AdminAuthMiddleware` 校验 `admin.api_key` 配置文件中的固定 Key。

### 弹性策略

- **熔断器**：`CircuitBreaker` 按 Project 维度，默认连续 5 次失败打开 → 30s 后半开探测 → 1 次成功后关闭。线程安全，通过 `Registry` 懒初始化。
- **限流器**：`RateLimiter` 接口，Tool 级别滑动窗口。Redis 实现，Redis 不可用时自动降级为 `NoopLimiter`（全放行）。
- **重试**：`doRequestWithRetry` 支持 fixed/指数退避 + ±25% jitter，默认只对 GET + 502/503/504 触发重试。

## 配置

`config/config.yaml` 是主配置，所有字段支持环境变量覆盖（`SERVER_HOST`, `DB_DSN`, `ADMIN_API_KEY`, `REDIS_ADDR` 等）。Docker 部署完全通过环境变量配置。

## 数据库

4 张核心表：`projects`（下游服务）、`tools`（工具定义 + 参数映射 JSON）、`clients`（API 客户端 + api_key_hash）、`client_tool_permissions`（多对多权限关联）。

## 日志

通过 `pkg/logger` 全局 zap SugaredLogger。HTTP 请求通过 `ChiMiddleware` 自动注入 `trace_id`（优先从 `X-Trace-ID` 请求头透传，否则生成 UUID）。使用 `logger.Info(msg, keyvals...)` 结构化风格，支持 `json`/`console` 两种格式。

## 前端

`web/` 是 Vite 5 + React 18 + TypeScript + Ant Design 5 管理后台。API 层在 `web/src/api/index.ts`，通过 `apiClient`（axios 封装）调用后端 `/admin/api/*` 接口。登录态通过 `AdminAuthMiddleware` 的 Bearer Token 校验。
