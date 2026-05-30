# P0 - 生产就绪（必须做）

> 目标：让项目达到可安全上生产的水平
> 预估周期：1 周

---

## 1. 健康检查 & 就绪探针

### 现状
无 `/health` 或 `/ready` 端点，K8s/Docker Compose 无法做健康检测和自动重启。

### 方案
```go
GET /health  → 200 OK (进程存活即可)
GET /ready   → 200 OK (DB Ping 通过 + 其他依赖检查)
```

### 实现要点
- `health` 始终返回 200
- `ready` 检查 MySQL 连接、Redis 连接（如有）
- 在 `internal/interface/` 下新增 `health_handler.go`

---

## 2. 结构化日志

### 现状
使用 Go 标准 `log` 包，无日志级别、无结构化字段、无 trace_id。

### 方案
引入 `go.uber.org/zap`：
- 日志级别：DEBUG / INFO / WARN / ERROR，配置可切换
- 每个请求注入 `trace_id`，贯穿 SSE/MCP 全链路
- JSON 格式输出，方便对接 ELK / Loki / 云日志服务
- 请求日志：method、path、status、duration、client_id

### 实现要点
- 新建 `pkg/logger/logger.go` 封装 zap
- 在 Chi 路由上挂载日志中间件
- MCP Handler 中携带 trace_id 到下游调用

---

## 3. 请求重试 & 熔断器

### 现状
下游 HTTP 调用失败立即返回错误，无重试、无熔断保护。

### 方案

#### 3.1 可配置重试策略
在工具配置中增加字段：
```json
{
  "retry_count": 3,
  "retry_backoff": "exponential",  // fixed / exponential
  "retry_on_status": [502, 503, 504]
}
```
- 仅对幂等请求 (GET) 或 5xx 错误重试
- 指数退避：1s → 2s → 4s

#### 3.2 熔断器
- 引入 `github.com/sony/gobreaker` 或自实现
- 按 project（下游服务）维度熔断
- 连续失败 N 次 → OPEN（快速失败）→ 半开探测 → CLOSE
- 配置：`max_failures`、`timeout`、`half_open_max_requests`

### 实现要点
- 修改 `internal/infrastructure/http/http_client.go`，增加重试 & 熔断逻辑
- 熔断器状态通过 Admin API 暴露，前后端展示

---

## 4. 参数校验增强

### 现状
`ParamMapper` 仅校验必填项（`required: true`），无类型校验。

### 方案
根据 JSON Schema 类型做完整校验：

| 类型 | 校验规则 |
|------|----------|
| `string` | `maxLength`、`minLength`、`pattern`（正则） |
| `number` / `integer` | `minimum`、`maximum`、`multipleOf` |
| `enum` | 值必须在白名单内 |
| `array` | `minItems`、`maxItems`、元素类型校验 |
| `object` | `required` 子属性校验 |

### 实现要点
- 新建 `internal/domain/mapper/param_validator.go`
- 校验失败返回 JSON-RPC 错误码 `-32003`（已预留）
- 错误信息清晰指出哪个参数、什么原因

---

## 5. 下游超时按工具配置

### 现状
HTTP Client 只有全局超时，无法针对不同工具设置不同超时。

### 方案
- `tools` 表增加 `timeout_ms` 字段（默认 30000）
- `HTTPClient.DoRequest()` 按工具配置创建 context.WithTimeout
- 超时返回明确错误信息（区别于下游 5xx）
