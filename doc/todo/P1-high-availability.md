# P1 - 高可用 & 可扩展

> 目标：支持多实例水平扩展，保护系统稳定
> 预估周期：2 周

---

## 1. Redis 集成 — Session 持久化

### 现状
Session 纯内存 `map[string]*Session` + `sync.RWMutex`：
- 单实例，无法水平扩展
- 服务重启后所有 Session 丢失
- `config.yaml` 已有 Redis 配置但未连接

### 方案
引入 `github.com/redis/go-redis/v9`：
- Session 数据序列化为 JSON 存入 Redis
- Key: `session:{session_id}`，TTL: 30 分钟（或可配置）
- SSE 推送通道改为 Redis Pub/Sub 跨实例广播
- 保留内存作为一级缓存（读性能），Redis 作为二级存储

### 实现要点
- 新建 `internal/infrastructure/redis/client.go`
- 重构 `internal/application/session/manager.go`
  - 定义 `SessionStore` 接口（MemoryStore / RedisStore）
  - 通过配置选择存储后端
- Redis Pub/Sub 用于跨实例 SSE 消息投递

---

## 2. 限流 (Rate Limiting)

### 现状
JSON-RPC 错误码 `-32002` 已预留，文档中有滑动窗口方案设计，但未实现。

### 方案

#### 2.1 按客户端限流
- 每个 Client 可配置 `rate_limit`（每秒调用次数）
- `clients` 表增加 `rate_limit_per_sec` 字段（0 表示不限）

#### 2.2 按 IP 限流
- 全局限流（所有请求总和）
- 防止恶意攻击

#### 2.3 滑动窗口 + Redis Lua
```
Key: rate_limit:{client_id}:{window_ts}
Lua: INCR + EXPIRE + 判断是否超限
```

### 实现要点
- 新建 `internal/domain/service/rate_limiter.go`（接口）
- 新建 `internal/infrastructure/redis/rate_limiter.go`（Redis 实现）
- Chi 中间件挂载到 `/sse` 和 `/messages` 路由
- 限流命中返回错误码 `-32002`，HTTP 429
- Admin Dashboard 展示每个客户端限流状态

---

## 3. Prometheus 监控指标

### 现状
无任何指标暴露，运维无法感知系统状态。

### 方案
引入 `github.com/prometheus/client_golang`：

```go
// 必须指标
mcp_tools_calls_total{client_id, tool_id, status}       // Counter
mcp_tools_call_duration_seconds{client_id, tool_id}      // Histogram
mcp_sessions_active                                       // Gauge
mcp_downstream_errors_total{project_id, status_code}     // Counter
mcp_rate_limit_hits_total{client_id}                     // Counter
mcp_http_requests_total{method, path, status}            // Counter
```

### 实现要点
- 暴露 `GET /metrics` 端点（Prometheus 抓取）
- 不经过 Admin Auth 中间件（或支持 Prometheus basic auth）
- 可选：提供 Grafana Dashboard JSON 模板
- 指标 label 注意基数爆炸（tool_id 过多时用 exemplar 替代）

---

## 4. TLS/HTTPS 支持

### 现状
仅 HTTP，TLS 完全依赖外部 Nginx。

### 方案
- 配置增加 `server.tls` 段：
  ```yaml
  server:
    tls:
      enabled: true
      cert_file: "/etc/certs/server.crt"
      key_file: "/etc/certs/server.key"
  ```
- 支持 mTLS（下游服务认证）：
  ```yaml
  server:
    tls:
      client_ca_file: "/etc/certs/ca.crt"  # 可选，开启 mTLS
  ```
- `tools` 表增加 `tls_skip_verify` 字段（下游自签名证书场景）

### 实现要点
- `main.go` 中判断配置选择 `http.ListenAndServe` 或 `http.ListenAndServeTLS`
- 下游 HTTPClient 支持自定义 TLS 配置

---

## 5. 优雅关闭 (Graceful Shutdown)

### 现状
未找到 `signal.NotifyContext` 或 graceful shutdown 逻辑。

### 方案
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// 启动 server
go srv.ListenAndServe()

<-ctx.Done()
// 停止接收新请求
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
// 关闭 DB 连接
// 清理资源
```

---

## 6. 数据库连接池优化

### 现状
sqlx 默认连接池参数可能不适用生产。

### 方案
```go
db.SetMaxOpenConns(25)        // 最大打开连接数
db.SetMaxIdleConns(10)        // 最大空闲连接数
db.SetConnMaxLifetime(5*time.Minute)  // 连接最大存活时间
db.SetConnMaxIdleTime(1*time.Minute)  // 空闲连接最大存活时间
```
这些值应可配置（`config.yaml` 的 `database` 段）。

---

## 7. Admin API 增强

### 7.1 分页完善
当前分页解析较基础，增加：
- 返回 `total`、`page`、`page_size`、`total_pages`
- 支持排序（`sort_by`、`sort_order`）

### 7.2 搜索/筛选
- 工具列表：按名称模糊搜索、按项目筛选、按状态筛选
- 客户端列表：按名称搜索、按状态筛选

### 7.3 批量操作
- 批量启用/禁用工具
- 批量授权（多客户端 → 多工具）
