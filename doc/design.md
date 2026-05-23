# MCP Gateway 设计文档

## 1. 项目背景

公司内部多个项目对外提供了 REST API 接口。随着 AI Agent 的普及，需要将这些 HTTP 接口以 **MCP 协议**（Model Context Protocol）的形式统一暴露给 Agent 使用。MCP Gateway 作为统一入口，负责鉴权、协议转换、路由转发、限流、熔断等能力，让存量 HTTP 服务**零改造**接入 MCP 生态。

## 2. 核心工作流程

```
Agent (Cursor/Claude)              MCP Gateway                     下游 HTTP 服务
      │                                │                                │
      │── tools/list ──────────────────▶│                                │
      │                                │  查询数据库工具定义             │
      │◄── ToolDefinition 列表 ────────│                                │
      │                                │                                │
      │── tools/call(account_get_info)─▶│                                │
      │                                │  ① 鉴权（校验收到的 API Key）  │
      │                                │  ② 限流检查                   │
      │                                │  ③ 查询工具定义 & 参数映射规则 │
      │                                │  ④ 协议转换: MCP 参数 → HTTP   │
      │                                │                                │
      │                                │─── HTTP 请求 ────────────────▶│
      │                                │◄── HTTP 响应 ─────────────────│
      │                                │                                │
      │                                │  ⑤ 响应处理: 2xx 原样返回      │
      │                                │     4xx/5xx 返回 MCP 错误码   │
      │◄── MCP 响应 ───────────────────│                                │
```

### 2.1 协议转换核心逻辑

Gateway 收到 MCP `tools/call` 后，执行协议转换：

```
Agent 发送: tools/call { name: "account_get_info", arguments: { "user_id": "123" } }

步骤:
  1. 根据参数映射配置，逐参数分配位置:
     - user_id → path（替换 URL 模板中的 {user_id}）
     - 非必填且未传的参数 → 跳过
     - header 参数 → 从 Gateway 凭证仓库中注入
  2. 组装 HTTP 请求:
     - URL: 路径模板替换后拼接 Query 参数
     - Header: 注入固定头 + 鉴权头
     - Body: POST 请求时，body 参数序列化为 JSON
  3. 发起 HTTP 调用
  4. 检查响应:
     - 2xx → 响应体原样塞入 MCP content 字段
     - 4xx/5xx → 返回 MCP 错误码，不暴露服务端堆栈

最终 HTTP 请求: GET /api/v1/accounts/123  Authorization: Bearer xxx
最终 MCP 响应:   { "content": [{ "type": "text", "text": "{...}" }] }
```

## 3. 架构分层设计

采用 DDD 分层架构：

```
┌──────────────────────────────────────────────────────────┐
│  Interface 层（MCP 协议端点 + Admin REST API）           │
│  - 接收 tools/list / tools/call → 解析 JSON-RPC 2.0     │
│  - 管理后台：clients / projects / tools CRUD             │
├──────────────────────────────────────────────────────────┤
│  Application 层（用例编排）                              │
│  ListToolsUseCase:   查库 → 提取 ToolDefinition → 返回  │
│  CallToolUseCase:    查库 → 参数路由 → HTTP 调用 → 返回 │
│  ManageClientUseCase: 客户端 CRUD 编排                   │
├──────────────────────────────────────────────────────────┤
│  Domain 层（核心业务逻辑）                               │
│  - 参数映射（path/query/body/header 分配）               │
│  - 路径模板替换（{user_id} → 123）                      │
│  - HTTP 请求组装（URL + Header + Body）                  │
│  - 响应处理（状态码判断、结果包装）                      │
│  - 限流策略（滑动窗口 / 令牌桶）                         │
├──────────────────────────────────────────────────────────┤
│  Infrastructure 层（基础设施）                           │
│  - 数据库读写（工具定义、参数映射、客户端信息）          │
│  - HTTP 客户端（调用下游服务）                            │
│  - Redis（限流计数器、缓存）                              │
│  - 审计日志写入（Kafka / ClickHouse）                    │
└──────────────────────────────────────────────────────────┘
```

## 4. 数据库表设计

### 4.1 `projects` — 下游项目表（被调方）

记录每个接入 Gateway 的后端 HTTP 项目的鉴权凭证和基础地址。

```sql
CREATE TABLE projects (
    project_id     VARCHAR(64) PRIMARY KEY,   -- 项目唯一标识
    name           VARCHAR(128) NOT NULL,     -- 项目名称
    description    TEXT,                      -- 项目描述
    api_key        VARCHAR(256),              -- 调用该项目时的凭证
    base_url       VARCHAR(512) NOT NULL,     -- 项目的基础地址，如 http://account-service:8080
    config         JSON,                      -- 通用配置（如项目级别限流）
    status         TINYINT DEFAULT 1,         -- 1启用 0禁用
    created_at     DATETIME,
    updated_at     DATETIME
);
```

### 4.2 `tools` — 接口定义表

描述每个 HTTP 接口如何暴露为 MCP 工具。

```sql
CREATE TABLE tools (
    tool_id        BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id     VARCHAR(64) NOT NULL,          -- 所属项目，关联 projects 表
    name           VARCHAR(128) NOT NULL UNIQUE,  -- MCP 工具名，如 account_get_info
    description    TEXT NOT NULL,                 -- 工具描述，直接影响 Agent 是否选用
    protocol       VARCHAR(16) DEFAULT 'HTTP',    -- 协议类型
    http_method    VARCHAR(10),                   -- GET / POST / PUT / DELETE
    url_template   VARCHAR(512) NOT NULL,         -- URL 模板，如 /api/v1/accounts/{user_id}
    timeout_ms     INT DEFAULT 5000,              -- 超时时间（毫秒）
    config         JSON,                          -- 通用配置（如接口级别限流）
    status         TINYINT DEFAULT 1,             -- 1启用 0禁用，禁用时 tools/list 不返回
    created_at     DATETIME,
    updated_at     DATETIME,
    INDEX idx_project (project_id)
);
```

### 4.3 `tool_params` — 参数转换表

定义每个 MCP 参数到 HTTP 请求的映射规则。

```sql
CREATE TABLE tool_params (
    param_id       BIGINT PRIMARY KEY AUTO_INCREMENT,
    tool_id        BIGINT NOT NULL,               -- 关联 tools 表
    name           VARCHAR(128) NOT NULL,         -- 参数名，如 user_id
    param_type     VARCHAR(32) DEFAULT 'string',  -- string / integer / boolean
    location       VARCHAR(16) NOT NULL,          -- path / query / body / header
    required       TINYINT DEFAULT 0,             -- 1必填 0非必填
    default_value  TEXT,                          -- 默认值（非必填参数时使用）
    description    VARCHAR(512),                  -- 参数说明
    sort_order     INT DEFAULT 0,                 -- 排序
    created_at     DATETIME,
    INDEX idx_tool (tool_id)
);
```

### 4.4 `clients` — 上游调用方表（谁调用 Gateway）

每个接入 Gateway 的团队或系统都需要注册一个 Client，获得 API Key。

```sql
CREATE TABLE clients (
    client_id      VARCHAR(64) PRIMARY KEY,
    name           VARCHAR(128) NOT NULL,         -- 团队/系统名称
    api_key        VARCHAR(256) NOT NULL,         -- 调用 Gateway 的凭证
    config         JSON,                          -- 客户端级别通用配置（限流等）
    status         TINYINT DEFAULT 1,
    created_at     DATETIME,
    updated_at     DATETIME
);
```

### 4.5 `client_tool_permissions` — 权限与组合限流关联表

多对多关系：一个 Client 可访问多个 Tool，一个 Tool 可被多个 Client 访问。

```sql
CREATE TABLE client_tool_permissions (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id      VARCHAR(64) NOT NULL,          -- 关联 clients 表
    tool_id        BIGINT NOT NULL,               -- 关联 tools 表
    config         JSON,                          -- 组合级别限流等配置
    created_at     DATETIME,
    UNIQUE KEY uk_client_tool (client_id, tool_id)
);
```

### 4.6 表关系总结

```
clients ─────┐
              ├── client_tool_permissions ──┬── tools ──── projects
clients ─────┘                              │
                                    tool_params
```

- `projects` 表管理**被调方**（下游 HTTP 服务）的鉴权和基础信息
- `tools` 表将每个 HTTP 接口描述为一个 MCP 工具
- `tool_params` 表定义 MCP 参数到 HTTP 请求位置的映射规则
- `clients` 表管理**调用方**（上游 Agent 团队）的鉴权
- `client_tool_permissions` 表管理调用方与工具的授权关系，以及组合级别的限流配置

## 5. 限流策略

### 5.1 三级限流粒度

| 级别 | 含义 | 配置位置 |
|------|------|----------|
| 客户端级别 | 某团队每秒钟最多调 100 次（不限接口） | `clients.config` |
| 接口级别 | 某接口每秒钟最多调 1000 次（不限调用方） | `tools.config` |
| 组合级别 | 某团队调用某接口每秒钟最多调 50 次 | `client_tool_permissions.config` |

### 5.2 实现方式

- **算法**：推荐滑动窗口或令牌桶
- **存储**：建议使用 Redis（INCR + EXPIRE 或 Lua 脚本），Gateway 多实例部署时可共享限流状态
- **策略**：三级限流依次判断，任一触发即拒绝

## 6. 错误处理策略

| 场景 | Gateway 处理 | 返回给 Agent |
|------|-------------|-------------|
| API Key 无效 | 直接拒绝 | MCP 错误码 -32001 鉴权失败 |
| 限流触发 | 拒绝请求 | MCP 错误码 -32002 请求过于频繁 |
| 参数校验失败 | 返回具体字段错误 | MCP 错误码 -32003 参数错误 |
| 下游 4xx | 原样返回状态码和错误信息 | MCP 错误码 -32004 + 错误描述（不暴露堆栈） |
| 下游 5xx / 超时 | 统一返回服务不可用 | MCP 错误码 -32005 下游服务异常 |
| 工具不存在 | 查询数据库未找到 | MCP 错误码 -32602 工具不存在 |

## 7. 建议的迭代计划

### V1（1-2 周）
- 鉴权（API Key 校验）
- 基础 MCP 协议（SSE 传输、tools/list、tools/call）
- 协议转换（参数映射、HTTP 请求组装）
- CRUD 管理 API（clients / projects / tools 增删改查）

### V2（2-4 周）
- 审计日志（全量请求/响应记录）
- 限流（三级限流、Redis 计数器）
- 健康检查与熔断

### V3（4-6 周）
- 缓存（`tools/list` 等幂等请求做 TTL 缓存）
- 参数校验（JSON Schema 校验）
- 敏感信息脱敏

### V4+
- 灰度发布（新版工具流量切分）
- WebSocket 传输支持
- gRPC 协议支持
- 管理控制台 Web UI

## 8. 项目目录结构

```
mcp-gateway/
├── cmd/
│   └── gateway/
│       └── main.go              # 入口
├── internal/
│   ├── interface/               # 接口层
│   │   ├── mcp/                 # MCP 协议端点（tools/list, tools/call）
│   │   └── admin/               # 管理后台 REST API（clients/projects/tools CRUD）
│   ├── application/             # 应用层（用例编排）
│   │   ├── list_tools.go
│   │   ├── call_tool.go
│   │   └── manage_client.go
│   ├── domain/                  # 领域层（核心业务逻辑）
│   │   ├── param_mapper.go      # 参数映射
│   │   ├── http_builder.go      # HTTP 请求组装
│   │   ├── response_handler.go  # 响应处理
│   │   └── rate_limiter.go      # 限流策略
│   ├── infrastructure/          # 基础设施层
│   │   ├── db/                  # 数据库读写
│   │   ├── http_client.go       # HTTP 客户端
│   │   ├── redis.go             # Redis 客户端
│   │   └── audit.go             # 审计日志
│   └── config/
│       └── config.go            # 配置
├── doc/
│   └── design.md                # 本设计文档
├── config.yaml
└── go.mod
```
