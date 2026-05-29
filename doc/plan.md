# MCP Gateway — 最小可行性方案执行计划

> 目标：实现最基础的 MCP ↔ HTTP 协议转换和请求转发，共 12 步。

---

## 数据库设计（V1 仅 2 张表）

### `projects` —— 下游服务

```sql
CREATE TABLE projects (
    project_id     VARCHAR(64) PRIMARY KEY,
    name           VARCHAR(128) NOT NULL,
    base_url       VARCHAR(512) NOT NULL,
    description    TEXT,
    status         TINYINT NOT NULL DEFAULT 1,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### `tools` —— 工具定义（含参数映射）

```sql
CREATE TABLE tools (
    tool_id        BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id     VARCHAR(64) NOT NULL,
    name           VARCHAR(128) NOT NULL,
    title          VARCHAR(256),
    description    TEXT,
    http_method    VARCHAR(10) NOT NULL DEFAULT 'GET',
    url_template   VARCHAR(512) NOT NULL,
    timeout_ms     INT NOT NULL DEFAULT 5000,
    params         JSON,
    status         TINYINT NOT NULL DEFAULT 1,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_name (name),
    KEY idx_project_id (project_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

`params` JSON 格式示例：

```json
[
  {
    "name": "user_id",
    "type": "string",
    "location": "path",
    "required": true,
    "description": "用户 ID"
  },
  {
    "name": "fields",
    "type": "string",
    "location": "query",
    "required": false,
    "default_value": "name,email",
    "description": "返回字段列表"
  },
  {
    "name": "X-Trace-Id",
    "type": "string",
    "location": "header",
    "required": false
  }
]
```

---

## 步骤 1：项目依赖与骨架

**产出**：能跑起来的空服务

- `go get` 引入依赖：mysql driver、sqlx、chi
- 写 `main.go` 骨架：读取配置 → 启动 HTTP Server
- 写配置加载逻辑，从 `config.yaml` 读取 `database.dsn`

---

## 步骤 2：Model + 数据库初始化

**产出**：2 个 Go 结构体 + 建表脚本

- 定义 `Project` 和 `Tool` struct（`params` 字段为 `json.RawMessage` 或 `string`）
- 写 `init.sql`
- 写数据库连接初始化函数

---

## 步骤 3：JSON-RPC 2.0 消息模型

**产出**：`pkg/jsonrpc/` 包，可序列化/反序列化

```
pkg/jsonrpc/
  message.go → Request, Response, Error 结构体 + Marshal/Unmarshal
```

---

## 步骤 4：MCP 协议消息类型

**产出**：`pkg/mcp/` 包，定义 MCP 特有消息结构

```
pkg/mcp/
  initialize.go → InitializeRequest, InitializeResult, ServerCapabilities
  tools.go      → ToolListResult, ToolCallRequest, ToolCallResult
```

---

## 步骤 5：Session 管理器（内存）

**产出**：`internal/domain/session.go`，纯内存

```go
type Session struct {
    ID              string
    ProtocolVersion string
    CreatedAt       time.Time
}

type SessionManager struct {
    mu       sync.RWMutex
    sessions map[string]*Session
}
```

---

## 步骤 6：MCP HTTP 端点 + 路由分发

**产出**：`POST /mcp` 能收消息，按 method 分发给对应 Handler

```
internal/interface/mcp/
  handler.go → 统一入口，解析 JSON-RPC，路由到具体 method handler
  router.go  → chi 路由注册
```

---

## 步骤 7：initialize 处理

**产出**：Agent 能成功握手

1. 创建 Session
2. 在 HTTP Response Header 设置 `Mcp-Session-Id`
3. 返回 `InitializeResult`
4. 收到 `notifications/initialized` 标记 session 为已初始化

---

## 步骤 8：tools/list 处理

**产出**：Agent 能拿到工具列表

1. 从 Header 取 SessionID，验证 session 已初始化
2. `SELECT * FROM tools WHERE status = 1`
3. 每条 tool 的 `params` JSON 转 `inputSchema`（JSON Schema 格式）
4. 返回 `{tools: [...]}`

---

## 步骤 9：参数映射引擎（纯函数，无 IO）

**产出**：`internal/domain/param_mapper.go`，可单测

```go
type MappedRequest struct {
    Path   string
    Query  url.Values
    Header http.Header
    Body   []byte
}

func MapParams(tool *Tool, arguments map[string]interface{}) (*MappedRequest, error)
```

测试用例：path 替换、query 拼接、header 注入、body 序列化、必填缺失、默认值

---

## 步骤 10：HTTP 请求转发 + 响应处理

**产出**：

```
internal/domain/http_builder.go
  → 组装 *http.Request，发起调用

internal/domain/response_handler.go
  → HTTP 响应 + 错误 → MCP 响应格式
    2xx → content + isError:false
    4xx → content + isError:true
    5xx/超时 → JSON-RPC error -32005
```

---

## 步骤 11：tools/call 处理 —— 串联全流程

**产出**：Agent 能调工具并拿到结果

1. 验证 session
2. 查数据库：`SELECT t.*, p.base_url FROM tools t JOIN projects p ... WHERE t.name = ?`
3. 参数映射：`param_mapper.MapParams(tool, arguments)`
4. HTTP 转发：`http_builder.DoRequest(project, tool, mappedReq)`
5. 响应处理：`response_handler.WrapResult(httpResp, err)`
6. 返回 MCP 响应

---

## 步骤 12：主函数集成 + 启动验证

**产出**：完整的可运行服务

1. 加载配置
2. 连接数据库
3. 创建 sessionManager
4. 创建 toolRepo
5. 创建 paramMapper
6. 创建 mcpHandler（注入所有依赖）
7. 注册路由：`POST /mcp`
8. 启动 HTTP Server
9. 启动 session 清理后台 goroutine

**验证**：curl 模拟 Agent 请求，走通 `initialize → tools/list → tools/call` 完整链路

---

## 步骤依赖图

```
步骤 1 (骨架)
  └── 步骤 2 (Model + DB)
  └── 步骤 3 (JSON-RPC)
        └── 步骤 4 (MCP 消息类型)
              └── 步骤 5 (Session)
                    └── 步骤 6 (HTTP 端点 + 路由分发)
                          ├── 步骤 7 (initialize)
                          ├── 步骤 8 (tools/list)
                          └── 步骤 9 (参数映射) → 步骤 10 (HTTP 转发)
                                └── 步骤 11 (tools/call)
                                      └── 步骤 12 (主函数集成)
```
