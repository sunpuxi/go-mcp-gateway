# MCP Gateway 逐点详解

本系列从零开始逐步讲解 MCP Gateway 的核心概念和实现原理。

> 写作思路：一个概念接一个概念，理解一个再继续下一个。

---

## 第 1 点：MCP 是什么？Gateway 要解决什么问题？

### 1.1 MCP（Model Context Protocol）是什么

想象一下：你是一个 AI Agent（比如 Cursor、Claude Desktop），你想操作外部系统——查数据库、调用 API、读文件。MCP 就是**定义 Agent 怎么跟外部系统对话的标准协议**。

类比：

| 协议 | 作用 |
|------|------|
| HTTP | 浏览器跟 Web 服务器对话 |
| JDBC | Java 程序跟数据库对话 |
| MCP  | AI Agent 跟工具/数据源对话 |

MCP 的核心交互只有三个动作：

| 动作 | 含义 | 类比 HTTP |
|------|------|-----------|
| `tools/list` | Agent 问：你有什么工具？ | 看菜单 |
| `tools/call` | Agent 说：帮我执行这个工具 | 点菜 |
| `initialize` | 握手：我和你都支持什么能力？ | 建立连接 |

### 1.2 MCP Gateway 要解决什么问题

**问题场景**：公司内部有几十个 HTTP 服务（订单服务、用户服务、支付服务...），现在想让 AI Agent 能调用它们。

**方案 A（直连）**：让每个 HTTP 服务自己实现一个 MCP Server。

- 缺点：每个服务都要改造，成本高，排期长

**方案 B（Gateway）**：做一个统一的 MCP Gateway，它对外说 MCP 语言，对内帮你把请求转成 HTTP。存量服务**零改造**。

```
Agent 说 MCP 语言          Gateway 翻译             下游说 HTTP
┌─────┐   tools/call      ┌────────────┐  GET/POST  ┌──────────┐
│ Cursor│ ──────────────▶  │ MCP Gateway│ ────────▶ │ 旧 HTTP  │
└─────┘                    │(协议翻译官)│            │ 服务     │
                           └────────────┘            └──────────┘
```

Gateway 的核心职责就是一句话：**听懂 Agent 的 MCP 指令，翻译成后端 HTTP 服务能理解的请求，再把结果翻译回去**。

---

## 第 2 点：Initialize —— MCP 的"握手"阶段

### 2.1 为什么需要 initialize？

**一句话回答**：因为 MCP 是一个**有状态的协议**，两端必须先对齐版本和能力，才能干活。

类比：你打电话给客服说"帮我查一下订单"，客服会先确认"你好，这里是 XX 客服，请问有什么可以帮您？"然后你才能说需求。如果跳过确认直接说"查订单"，客服会懵。

**initialize 在 MCP 里做三件事：**

1. **版本对齐**：Agent 和 Gateway 可能使用不同版本的 MCP 协议（如 2024-11-05 版和 2025-06-18 版）。握手时确定双方都支持哪个版本，否则后续消息格式会解析错误。

2. **能力声明**：Agent 告诉 Gateway 自己支持什么（比如：我能接收服务端主动推送的通知吗？），Gateway 也告诉 Agent 自己支持什么（比如：我有 tools、有 resources）。

3. **身份识别**：Agent 告诉 Gateway "我是谁"（客户端名称、版本），方便后续鉴权和日志。

### 2.2 initialize 是怎么进行的？

分三步走，每步都是 JSON-RPC 2.0 消息：

#### 第一步：Agent 发 `initialize` 请求

```
Agent（Cursor）                              Gateway
    │                                            │
    │── initialize request ─────────────────────▶│
    │   {                                        │
    │     "jsonrpc": "2.0",                      │
    │     "id": 1,                               │
    │     "method": "initialize",                │
    │     "params": {                            │
    │       "protocolVersion": "2025-06-18",     │  ← 我用这个版本
    │       "capabilities": {                    │  ← 我支持这些能力
    │         "tools": {},                       │     我能使用工具
    │         "resources": {}                    │     我能读取资源
    │       },                                   │
    │       "clientInfo": {                      │  ← 我是谁
    │         "name": "cursor",                  │
    │         "version": "1.0.0"                 │
    │       }                                    │
    │     }                                      │
    │   }                                        │
    │                                            │
```

#### 第二步：Gateway 回复 `InitializeResult`

```
    │                                            │
    │◄── initialize response ───────────────────│
    │   {                                        │
    │     "jsonrpc": "2.0",                      │
    │     "id": 1,                               │
    │     "result": {                            │
    │       "protocolVersion": "2025-06-18",     │  ← OK，我也用这个版本
    │       "capabilities": {                    │  ← 我支持这些
    │         "tools": {                         │     我有工具
    │           "listChanged": true              │     工具变了我会通知你
    │         },                                 │
    │         "resources": {}                    │     我也有资源
    │       },                                   │
    │       "serverInfo": {                      │  ← 我是谁
    │         "name": "mcp-gateway",             │
    │         "version": "1.0.0"                 │
    │       }                                    │
    │     }                                      │
    │   }                                        │
    │                                            │
```

#### 第三步：Agent 发送 `notifications/initialized`

这只是一个**通知**（没有 `id`，不要求回复），表示"我已收到你的能力声明，准备好了"。

```
    │── notifications/initialized ─────────────▶│
    │   {                                        │
    │     "jsonrpc": "2.0",                      │
    │     "method": "notifications/initialized"  │  ← 我准备好了！
    │   }                                        │
    │                                            │
    │           从此之后，才能发 tools/list 等请求     │
    │                                            │
```

### 2.3 常见的设计误区

如果没有实现 initialize，直接从 `tools/list` 开始，就会出现：

- Agent 连上 Gateway，Gateway 应该先问"你好，你是谁，用什么协议版本？"
- 但实际 Gateway 直接说"点菜吧"（暴露 tools/list 端点）
- 实际 MCP 客户端（Cursor、Claude Desktop）会等待初始化完成才发后续请求，导致连接失败

---

## 第 3 点：Streamable HTTP —— 消息在网络层面怎么收发

### 3.1 先明确问题

前面讲的 initialize、tools/list、tools/call 是"应用层"的 JSON-RPC 消息。现在要回答一个更底层的问题：**这些 JSON 消息，通过什么机制从 Agent 送到 Gateway，又从 Gateway 送回 Agent？**

这就是 MCP 的**传输层**（Transport Layer）要解决的。

### 3.2 两种传输方式

MCP 规范定义了两个标准传输：

| 传输方式 | 适用场景 | 一句话理解 |
|---------|---------|-----------|
| **stdio** | Agent 和 Server 跑在同一台机器上 | Agent 启动 Server 作为子进程，通过 stdin/stdout 通信 |
| **Streamable HTTP** | Agent 和 Server 跨网络 | 通过 HTTP 请求来收发消息 |

Gateway 是独立部署的网络服务，Agent 通过网络连过来，所以用 **Streamable HTTP**。

### 3.3 Streamable HTTP 的核心规则

两条规则：

```
Agent → Gateway: 每条消息都用一个单独的 HTTP POST 发过去
Gateway → Agent: 可以走 HTTP POST 的响应体返回，也可以走 SSE 长连接推送
```

完整看一次 initialize 在网络层的交互：

```
Agent（Cursor）                                    Gateway（HTTP Server）
    │                                                    │
    │  POST /mcp   (第一条消息：initialize 请求)           │
    │  Headers:                                           │
    │    Content-Type: application/json                   │
    │    Accept: application/json, text/event-stream       │
    │    MCP-Protocol-Version: 2025-06-18                  │
    │  Body:                                              │
    │  {                                                  │
    │    "jsonrpc": "2.0",                                │
    │    "id": 1,                                         │
    │    "method": "initialize",                          │
    │    "params": { ... }                                │
    │  }                                                  │
    │ ──────────────────────────────────────────────────▶  │
    │                                                     │
    │  HTTP 200                                           │
    │  Content-Type: application/json                     │
    │  Mcp-Session-Id: abc-123                            │  ← 服务端创建会话
    │  Body:                                              │
    │  {                                                  │
    │    "jsonrpc": "2.0",                                │
    │    "id": 1,                                         │
    │    "result": { ... }                                │
    │  }                                                  │
    │ ◄──────────────────────────────────────────────────  │
    │                                                     │
    │  POST /mcp  (第二条消息：notifications/initialized)   │
    │  Headers:                                           │
    │    Mcp-Session-Id: abc-123                          │  ← 后续请求带上 session
    │  Body:                                              │
    │  {                                                  │
    │    "jsonrpc": "2.0",                                │
    │    "method": "notifications/initialized"             │
    │  }                                                  │
    │ ──────────────────────────────────────────────────▶  │
    │                                                     │
    │  HTTP 202  (通知不需要回复，202 表示已接收)            │
    │  (无 body)                                          │
    │ ◄──────────────────────────────────────────────────  │
```

### 3.4 两种响应方式

| 情况 | Gateway 怎么回 |
|------|--------------|
| 普通响应（如 tools/list） | HTTP 200 + `Content-Type: application/json`，body 就是 JSON-RPC 响应 |
| 带推送的响应 | HTTP 200 + `Content-Type: text/event-stream`（SSE），后续还可以推送额外消息 |
| 通知类（丢弃不回复） | HTTP 202 Accepted，无 body |

### 3.5 GET 方法——服务端主动推送

除了 POST，Agent 还可以发 HTTP GET 打开一个 SSE 长连接，专门用来接收 Gateway 主动推送的消息（比如工具列表变更通知）。

```
Agent                                    Gateway
  │                                           │
  │  GET /mcp                                 │
  │  Accept: text/event-stream                │
  │  Mcp-Session-Id: abc-123                  │
  │ ──────────────────────────────────────▶    │
  │                                           │
  │  HTTP 200                                 │
  │  Content-Type: text/event-stream          │
  │                                           │
  │  data: {"jsonrpc":"2.0",                  │  ← Gateway 主动推送
  │         "method":"notifications/          │     "工具有更新了"
  │           tools/list_changed"}             │
  │ ◄──────────────────────────────────────    │
```

---

## 第 4 点：为什么需要会话管理？

### 4.1 背景：HTTP 是无状态的

你的 design doc 里想象的是这样：

```
Agent ── TCP 长连接 ──▶ Gateway
       一条连接一直通着，来回发消息
```

但实际 Streamable HTTP 是这样：

```
Agent                     Gateway
  │                          │
  │  POST /mcp (心跳消息)    │  HTTP 请求1，发完就断开
  │ ──────────────────────▶  │
  │ ◄──── HTTP 200 ────────  │
  │                          │
  │  POST /mcp (新消息)      │  HTTP 请求2，全新的请求
  │ ──────────────────────▶  │  Gateway：你是谁？
  │                          │
```

**问题**：Gateway 收到第二条 POST，怎么知道和第一条是同一个 Agent？

### 4.2 没有会话会怎样？

```
Agent                            Gateway
  │                                  │
  │  POST (initialize)               │  ✅ 握手成功
  │ ────────────────────────────▶    │
  │ ◄── InitializeResult ──────────  │
  │                                  │
  │  POST (notifications/initialized)│  ❓ 等一下，你谁啊？
  │ ────────────────────────────▶    │  没有 session，Gateway 不知道
  │                                  │  你初始化过，让你重新 init
  │                                  │
  │  POST (tools/call)               │  ❓ 又来？重新 init！
  │ ────────────────────────────▶    │
```

结果是：每次 POST 都要重新走一遍 initialize，永远进不了真正干活阶段。

### 4.3 一个类比

| 概念 | 类比 |
|------|------|
| **API Key** | 身份证（证明"你是谁"） |
| **Session ID** | 桌号（证明"这是你正在吃的这顿饭"） |

你进餐厅：
1. 出示身份证（API Key）→ 证明你是会员
2. 服务员给你一个桌号（Session ID）→ 记录"你坐这桌"
3. 之后你加菜、结账，说"3 号桌"就行了，不需要再掏身份证

没有 Session 相当于每次点菜都要重新掏身份证。

### 4.4 Session 具体解决了什么

**第 1 点：维护"已初始化"状态**

MCP 是有状态协议——initialize 之后才能发其他消息。Session 让 Gateway 能记住：

```
Session abc-123 → 已经初始化过了（protocolVersion: 2025-06-18）
→ 这个 session 来的请求，不需要重新握手
```

**第 2 点：关联多次请求到一个对话**

一次 AI Agent 交互过程中，可能会发多个 `tools/call`：

```
tools/call (account_get_info)    → Session abc-123
tools/call (order_create)        → Session abc-123
tools/list                       → Session abc-123
```

没有 session，Gateway 没法把这些请求归到同一个"用户会话"中，也就没法做：
- 全链路追踪（trace ID 跨请求串联）
- 审计日志拼接（这次 Agent 操作到底调了哪几个工具）
- 限流统计（"这个客户端本秒已经调了 80 次"——需要跨请求统计）

**第 3 点：服务端可以管理资源**

有 session 就可以：
- **过期**：session 超过 5 分钟无活动，自动清理
- **销毁**：Agent 发 HTTP DELETE 通知 Gateway "我用完了"
- **统计**：在线 session 数量、每个 session 的请求量

### 4.5 极简实现示例

```go
// 用一个 map 维护 session（生产环境用 Redis）
type Session struct {
    ID              string
    ClientID        string    // API Key 鉴权后关联的 client
    ProtocolVersion string
    CreatedAt       time.Time
    ExpiresAt       time.Time
}

// initialize 处理函数中：
func handleInitialize(req InitializeRequest) (InitializeResult, string) {
    sessionID := uuid.New().String()
    // 创建 session，存起来
    // 在 HTTP Response Header 中返回 Mcp-Session-Id: <sessionID>
    return result, sessionID
}

// 后续所有请求先查 session：
func handleToolCall(sessionID string, req ToolCallRequest) {
    session := lookupSession(sessionID)
    if !session.Initialized {
        return error("please initialize first")
    }
    // 继续处理
}
```

**总结一句话**：HTTP 本身不记得你是谁，Session 让 Gateway 能"记住"这个连接上下文，把多次独立的 HTTP POST 串联成一次有状态的 MCP 会话。

---

## 第 5 点：`tools/call` 的内部处理链路

前面讲了 Agent 怎么连上来（initialize）、消息怎么在网络上收发（Streamable HTTP）、Gateway 怎么记住会话（Session）。现在进入**最核心的部分**：Agent 说"帮我执行这个工具"，Gateway 内部到底发生了什么。

### 5.1 一句话总结整条链路

```
Agent 发来 tools/call

  ↓

Gateway 做 6 件事，按顺序：

  ① 鉴权          → 检查 Session 关联的 API Key
  ② 限流检查      → 三级限流逐级判断
  ③ 查工具定义    → 从数据库读取 tools + tool_params
  ④ 参数映射      → MCP 参数 → HTTP 位置（path/query/body/header）
  ⑤ 发起 HTTP 请求 → 组装 URL/Header/Body，发给下游
  ⑥ 响应处理      → HTTP 响应 → MCP 响应（含错误码转换）

  ↓

Gateway 返回 MCP 响应给 Agent
```

### 5.2 逐步骤详解

#### 步骤①：鉴权

```
Agent 带着 Mcp-Session-Id 发来 tools/call

Gateway：
  1. 查 Session → 拿到 ClientID
  2. 查 clients 表 → Client 是否启用？API Key 是否有效？
  3. 查 client_tool_permissions 表 → 这个 Client 有权限调这个 tool 吗？

  任一不通过 → 返回 MCP 错误码 -32001（鉴权失败）
```

关键点：鉴权不是在 tools/call 这一步才做，而是在 **initialize 握手阶段**就已经验证了 API Key。tools/call 主要是**权限检查**（这个工具这个客户端能不能用）。

#### 步骤②：限流检查

三级限流逐级判断，**任何一个触发就拒绝**：

```
① 客户端级别：检查 clients.config 中的限流配置
   → "这个 Client 每秒最多 100 次"

② 接口级别：检查 tools.config 中的限流配置
   → "这个工具每秒最多 1000 次"

③ 组合级别：检查 client_tool_permissions.config 中的限流配置
   → "这个 Client 调这个工具每秒最多 50 次"

任一触发 → 返回 MCP 错误码 -32002（请求过于频繁）
```

每次检查就是 Redis 一条命令：
```
INCR rate_limit:{scope}:{key}
EXPIRE rate_limit:{scope}:{key} 1  (1 秒窗口)
```

#### 步骤③：查工具定义

```
Gateway 查数据库：

  tools 表:
    SELECT * FROM tools WHERE name = 'account_get_info' AND status = 1

  tool_params 表:
    SELECT * FROM tool_params WHERE tool_id = ? ORDER BY sort_order

  得到：
    - http_method: GET
    - url_template: /api/v1/accounts/{user_id}
    - timeout_ms: 5000
    - params: [{ name: "user_id", location: "path", required: true, type: "string" }]
```

#### 步骤④：参数映射（核心翻译逻辑）

这是整个 Gateway 最关键的翻译逻辑：

```
Agent 传来的 arguments:
  { "user_id": "123", "fields": "name,email" }

Gateway 从 tool_params 取出映射规则，逐参数处理：

  user_id  → location: path
             ↓ 替换 url_template 中的 {user_id}
             /api/v1/accounts/{user_id} → /api/v1/accounts/123

  fields   → location: query
             ↓ 拼接到 URL 后面作为查询参数
             /api/v1/accounts/123?fields=name,email

  (header 类型的参数 → 写入 HTTP Header)
  (body 类型的参数   → 序列化为 JSON 放入请求体)
  (非必填且未传的参数 → 跳过，或使用 default_value)
```

#### 步骤⑤：发起 HTTP 请求

参数映射完成后，Gateway 组装好完整的 HTTP 请求并发出去：

```
最终组装：

  URL:     GET /api/v1/accounts/123?fields=name,email
  Header:  Authorization: Bearer xxx
           Content-Type: application/json
           X-Request-Id: trace-uuid-123
  Body:    （GET 请求没有 body）
  Timeout: 5000ms

Gateway 发起 HTTP 调用 → 等待下游响应
```

注意这里：**Header 参数从哪里来？**
- 一部分来自参数映射（用户传的 header 类型参数）
- 一部分来自 `projects` 表里的凭证（`api_key` 字段自动注入）
- 一部分是 Gateway 自己加的（trace ID、content-type 等）

#### 步骤⑥：响应处理

收到下游 HTTP 响应后，Gateway 做最终转换：

```
下游返回 200 OK
  Body: { "user_id": "123", "name": "张三", "email": "zhang@example.com" }

Gateway 包装成 MCP 响应格式：
  {
    "jsonrpc": "2.0",
    "id": 3,                    ← 对应 tools/call 请求的 id
    "result": {
      "content": [
        {
          "type": "text",
          "text": "{\"user_id\":\"123\",\"name\":\"张三\",\"email\":\"zhang@example.com\"}"
        }
      ]
    }
  }
```

错误处理：

| 下游返回 | Gateway 处理 |
|---------|------------|
| 2xx | 正常返回，响应体原样塞入 MCP `content` |
| 4xx | 返回 MCP 结果，但 `isError: true`，错误描述中不暴露堆栈 |
| 5xx / 超时 | 返回 MCP 错误码 -32005（下游服务异常） |

### 5.3 完整链路时序图

```
Agent              Gateway                         下游 HTTP 服务
  │                    │                                │
  │  tools/call        │                                │
  │  {name:"account_   │                                │
  │   get_info",args:  │                                │
  │   {user_id:"123"}} │                                │
  │ ──────────────────▶│                                │
  │                    │  ① 鉴权检查                    │
  │                    │  ② 限流检查                    │
  │                    │  ③ 查工具定义 + 参数映射规则    │
  │                    │  ④ 参数映射:                   │
  │                    │     user_id → path             │
  │                    │     URL → /api/v1/accounts/123 │
  │                    │                                │
  │                    │  GET /api/v1/accounts/123      │
  │                    │  Authorization: Bearer xxx     │
  │                    │ ──────────────────────────────▶│
  │                    │                                │
  │                    │  200 OK                        │
  │                    │  { "user_id":"123","name":"张三" }│
  │                    │ ◄──────────────────────────────│
  │                    │                                │
  │                    │  ⑤ 组装 MCP 响应               │
  │  MCP 响应          │                                │
  │  {content:[{       │                                │
  │   type:"text",     │                                │
  │   text:"{...}"}]}  │                                │
  │ ◄──────────────────│                                │
```

### 5.4 对 design doc 的修正建议

原 design doc 把这 6 步放在了 Domain 层，但实际上：

| 步骤 | 建议放在哪层 | 理由 |
|------|-------------|------|
| ① 鉴权 | Infrastructure | 涉及查询数据库、验证密码 |
| ② 限流 | Infrastructure | 涉及 Redis 操作（接口实现） |
| ③ 查工具定义 | Infrastructure | 数据库查询 |
| ④ 参数映射 | **Domain** | 纯逻辑：字段映射、模板替换（原设计放对了） |
| ⑤ HTTP 请求 | Infrastructure | 实际发起网络调用 |
| ⑥ 响应处理 | **Domain** | 状态码转换、结果包装是纯逻辑 |

具体建议：
- `domain/param_mapper.go` 和 `domain/response_handler.go` 位置正确
- `domain/http_builder.go` 建议拆分：参数拼装逻辑留 Domain，实际发起 HTTP 调用移到 Infrastructure
- `domain/rate_limiter.go` 应该是 Domain 定义接口，Infrastructure 提供实现（RedisRateLimiter）

---

## 第 6 点：`tools/list` 的处理

### 6.1 tools/list 是干什么的

`tools/list` 是 Agent 用来**发现你的 Gateway 有哪些工具可用**的请求。

用餐厅类比：

```
tools/call  = 点菜（使用工具）
tools/list  = 看菜单（发现工具）
```

### 6.2 和 tools/call 的对比

| 对比项 | tools/call | tools/list |
|--------|-----------|-----------|
| 工作量 | 重：查数据库、参数映射、发 HTTP 请求 | 轻：查一次数据库，组装成列表 |
| 是否有鉴权？ | 有（参数级别 + 权限检查） | 有（需要知道你是谁，但不能暴露你没权限的工具） |
| 是否有限流？ | 有 | 建议有，但阈值可以更宽 |
| 参数 | 需要传工具名 + 参数 | 无参数（或可传 cursor 做分页） |
| 返回什么 | 工具的执行结果 | 工具的定义信息（名称、描述、输入 Schema） |

### 6.3 处理流程

`tools/list` 比 `tools/call` 简单得多：

```
Agent 发来 tools/list
  │
  ↓
① 鉴权 → 从 Session 拿到 ClientID
② 查数据库 → 查出该 Client 有权限的所有 tool（关联 client_tool_permissions）
③ 组装 MCP Schema → 把每一条 tool 转成 MCP 规范的工具定义格式
④ 返回给 Agent
```

### 6.4 返回什么数据

每条 tool 在列表中的格式，MCP 有明确规范：

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "account_get_info",
        "title": "查询账户信息",
        "description": "根据用户 ID 获取账户基本信息，包括姓名、邮箱、注册时间等",
        "inputSchema": {
          "type": "object",
          "properties": {
            "user_id": {
              "type": "string",
              "description": "用户 ID"
            }
          },
          "required": ["user_id"]
        }
      },
      {
        "name": "order_create",
        "title": "创建订单",
        "description": "为用户创建一笔新订单",
        "inputSchema": {
          "type": "object",
          "properties": {
            "user_id": { "type": "string", "description": "用户 ID" },
            "product_id": { "type": "string", "description": "商品 ID" },
            "quantity": { "type": "integer", "description": "数量", "default": 1 }
          },
          "required": ["user_id", "product_id"]
        }
      }
    ]
  }
}
```

关键字段：

| 字段 | 重要程度 | 说明 |
|------|---------|------|
| `name` | **必须** | 工具的唯一标识，tools/call 时通过它找到对应工具 |
| `description` | **极其重要** | Agent 的 LLM 通过这段文字决定是否选用这个工具。描述越清楚，Agent 越可能正确调用 |
| `inputSchema` | **必须** | JSON Schema 格式的参数描述，LLM 据此生成调用参数 |

### 6.5 `description` 为什么重要？

很多人会低估 `description` 的价值：

```
❌ 差的描述：
  "查询用户信息"

✅ 好的描述：
  "根据用户 ID 查询账户基本信息，包含姓名、邮箱、注册时间、账户余额。
   适用于用户身份核实、账户信息展示等场景。当 Agent 需要获取用户资料时
   应优先使用此工具。如果只需要用户列表，请使用 user_list 工具。"
```

好的描述能让 LLM 更准确地**选择合适的工具、生成正确的参数**。这是 MCP Gateway 和传统 API 网关一个很大的不同——传统 API 的调用方是程序员；MCP 的调用方是 LLM，它只能通过这段文字来理解你的工具。

### 6.6 权限过滤

`tools/list` 返回的列表应该是**过滤后的**——只返回当前 Client 有权调用的工具：

```sql
SELECT t.* FROM tools t
JOIN client_tool_permissions p ON t.tool_id = p.tool_id
WHERE p.client_id = ? AND t.status = 1
```

两个 Client 看到的工具列表可能完全不同：

```
Client A（订单团队）  →  能看到：order_create, order_query, refund
Client B（客服团队）  →  能看到：order_query, user_query, ticket_create
```

### 6.7 `listChanged` 通知机制

前面讲 initialize 时提到，Gateway 可以在能力声明中说 `"listChanged": true`。这意味着当工具列表发生变化时，Gateway 可以**主动通知** Agent。

**什么时候触发**：
- 管理后台新增了工具
- 管理后台禁用了工具
- Client 的权限发生了变化

**发生了什么**：

```
Gateway                                    Agent
  │                                           │
  │  (管理员在后台新增了一个工具)               │
  │                                           │
  │  POST (SSE stream 上主动推送)             │
  │  {                                        │
  │    "jsonrpc": "2.0",                      │
  │    "method": "notifications/tools/        │
  │              list_changed"                │
  │  }                                        │
  │ ─────────────────────────────────────▶     │
  │                                           │
  │  Agent 收到通知后，重新发 tools/list       │
  │  ◀─────────────────────────────────────    │
  │  获取最新的工具列表                         │
```

Agent 不用轮询，Gateway 有变化就推，效率高。

### 6.8 对 design doc 的影响

你的 `tools` 表中已有 `status` 字段（1启用/0禁用），这很好——`tools/list` 只需要返回 `status=1` 的。但还需要考虑：

1. **`inputSchema` 的生成**：数据库里存的是 `tool_params` 的扁平结构，但 MCP 要求的返回格式是 **JSON Schema**。需要一个转换逻辑把 `tool_params` 行转成 `inputSchema` 对象。

2. **权限过滤**：返回的列表要基于 `client_tool_permissions` 做过滤，不是全量返回。

3. **缓存**：`tools/list` 返回的数据变化不频繁，可以在 Redis 中缓存，减少数据库查询。但缓存失效时要通知 Agent `list_changed`。

---

## 第 7 点：Resources 和 Prompts——MCP 的另外两个原语

### 7.1 MCP 三大原语

MCP 定义了三个核心原语（Primitives），你的设计文档只考虑了 `tools`：

```
┌────────────────────────────────────────────────┐
│  MCP Server 能暴露什么                           │
│                                                  │
│  Tools     → AI Agent 可以"做"什么（动作）       │
│  Resources → AI Agent 可以"读"什么（数据）       │
│  Prompts   → AI Agent 可以用什么模板来交互       │
└────────────────────────────────────────────────┘
```

| 原语 | 类比 | 执行方式 |
|------|------|---------|
| **Tools** | 遥控器按钮（按一下执行动作） | tools/call |
| **Resources** | 书架上的书（翻开就能读） | resources/read |
| **Prompts** | 模板试卷（照着题目回答） | prompts/get |

### 7.2 Resources 详解

#### 是什么

Resources 让 Agent **读取数据**，而不是执行动作。适合暴露查询类的接口。

```
Agent: "查一下 ID 为 123 的用户信息"
       ── resources/read ──▶  Gateway ──▶ 下游 GET /api/v1/users/123
       ◀── 用户数据 ────────
```

#### 和 Tools 的区别

| 对比项 | Tools | Resources |
|--------|-------|-----------|
| 语义 | "帮我做一件事" | "帮我读一份数据" |
| LLM 的理解 | 有副作用（创建/修改/删除） | 无副作用（只读） |
| 类比 HTTP | POST/PUT/DELETE | GET |
| MCP 方法 | `tools/call` | `resources/read` |
| 参数传递 | 通过 arguments 对象 | 通过 URI（类似 RESTful 路径） |

#### 交互方式

**发现资源**：

```json
// Agent 发 resources/list
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/list"
}

// Gateway 返回
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "resources": [
      {
        "uri": "user://123",
        "name": "用户信息",
        "description": "用户的基本信息资料",
        "mimeType": "application/json"
      },
      {
        "uri": "order://list?status=pending",
        "name": "待处理订单列表",
        "description": "所有待处理的订单",
        "mimeType": "application/json"
      }
    ]
  }
}
```

**读取资源**：

```json
// Agent 发 resources/read
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "resources/read",
  "params": {
    "uri": "user://123"
  }
}

// Gateway 返回
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "contents": [
      {
        "uri": "user://123",
        "mimeType": "application/json",
        "text": "{\"user_id\":\"123\",\"name\":\"张三\",\"email\":\"zhang@example.com\"}"
      }
    ]
  }
}
```

#### Resources 的 URI 机制

Resources 使用 **URI** 来定位数据，这个 URI 是 Gateway 自己定义的，不是真正的网络 URL：

```
user://{user_id}                              → 用户资源
order://{order_id}                            → 订单资源
file:///var/log/app.log                       → 文件资源
db://users/schema                             → 数据库 schema
```

Gateway 收到 `resources/read` 请求后，解析 URI，查数据库找到对应的 URI 模式配置，然后转成 HTTP 请求。这和 tools 的参数映射本质是一样的，只是入口不同。

#### 对你的 Gateway 的意义

你的下游 HTTP 服务有很多 GET 接口——这些天然适合暴露为 **Resources**（只读）而不是 Tools（动作）。Agent 的 LLM 在推理时会更倾向于用 Resources 来做数据查询，用 Tools 来做数据变更。

给你的数据库加一张 `resources` 表：

```sql
CREATE TABLE resources (
    resource_id    BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id     VARCHAR(64) NOT NULL,
    uri_pattern    VARCHAR(256) NOT NULL,   -- 如 user://{user_id}
    name           VARCHAR(128),
    description    TEXT,
    http_method    VARCHAR(10) DEFAULT 'GET',
    url_template   VARCHAR(512) NOT NULL,   -- 如 /api/v1/users/{user_id}
    mime_type      VARCHAR(64) DEFAULT 'application/json',
    status         TINYINT DEFAULT 1,
    created_at     DATETIME,
    updated_at     DATETIME
);
```

然后 `resources/read` 的处理链路几乎可以复用 `tools/call` 的参数映射和 HTTP 调用逻辑。

### 7.3 Prompts 详解

#### 是什么

Prompts 是**预定义的提示模板**，Agent 可以用它来快速构造某种特定场景的交互提示。

一个例子：

```json
// Agent 发 prompts/list
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "prompts/list"
}

// Gateway 返回
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "prompts": [
      {
        "name": "user_query_assist",
        "description": "客服查用户信息时的提问模板",
        "arguments": [
          {
            "name": "user_id",
            "description": "用户 ID",
            "required": true
          }
        ]
      }
    ]
  }
}
```

**获取具体模板**：

```json
// Agent 发 prompts/get
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/get",
  "params": {
    "name": "user_query_assist",
    "arguments": {
      "user_id": "123"
    }
  }
}

// Gateway 返回
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "description": "客服查用户信息时的提问模板",
    "messages": [
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "请查询用户 123 的信息，包含姓名、邮箱和最近订单。\n如果用户存在，请展示其基本资料。\n如果不存在，请告知无法找到。"
        }
      }
    ]
  }
}
```

#### Prompts 在 Gateway 中的价值

对于你的场景，Prompts 可以：

1. **让下游 HTTP 服务的交互更可控**：模板里可以嵌入对 HTTP API 返回结果的解释逻辑
2. **降低 Agent 出错概率**：LLM 直接使用模板，比自己构造 prompt 更准确
3. **统一交互风格**：所有 Agent 都用同一套模板，行为一致

### 7.4 三个原语联合工作的示例

一次完整的对话交互可能同时用到三者：

```
Agent 需要做的：查用户信息 → 如果用户存在，发起退款

1. prompts/get "客服工作流"          → 拿到工作流模板
2. resources/read "user://123"      → 读取用户数据，发现用户存在
3. tools/call "refund_create"       → 创建退款（有副作用的操作）
```

### 7.5 对 design doc 的影响

| 原语 | V1 是否需要 | 理由 |
|------|-----------|------|
| Tools | **需要** | 核心功能，Gateway 的主要存在价值 |
| Resources | **建议 V2** | 很多下游 GET 接口更适合用 Resource 语义暴露，但可以通过 Tools 先兜底 |
| Prompts | V3+ | 锦上添花，不是必须 |

建议在 V1 的 `tools` 表 `protocol` 字段上多留几个值，为未来扩展做准备：

```
protocol: 'HTTP'       → 转为 HTTP 请求
protocol: 'MCP_STDIO'  → 转发给本地 MCP Server
protocol: 'MCP_HTTP'   → 转发给远程 MCP Server
```

这样未来支持 Resources 时，参数映射和 HTTP 调用逻辑可以复用。也不排除未来直接暴露一个后端 MCP Server 而不是 HTTP 服务的情况——你的 Gateway 可以变成 **MCP Hub**，聚合多个 MCP Server。

---

## 第 8 点：鉴权的完整设计——API Key、Session、权限三级联动

前面讲 Session 时提到过一个类比：

```
API Key  = 身份证（证明你是谁）
Session  = 桌号（证明你在吃哪顿饭）
权限     = 你能点什么菜（菜单上有些菜你不能点）
```

现在把这三者串起来，看一次完整的鉴权流程。

### 8.1 三级鉴权体系

```
┌─────────────────────────────────────────────────────────────┐
│  三级鉴权                                                    │
│                                                              │
│  第 1 级：传输层鉴权（API Key）                              │
│  → 验证 API Key 是否有效、Client 是否启用                     │
│  → 发生在每一次 HTTP 请求上                                   │
│                                                              │
│  第 2 级：会话层鉴权（Session）                               │
│  → 验证 Session 是否有效、是否已初始化、是否过期               │
│  → 发生在每一条 JSON-RPC 消息上                               │
│                                                              │
│  第 3 级：业务层鉴权（Permissions）                          │
│  → 验证 Client 是否有权限调用指定 Tool / 读取指定 Resource   │
│  → 发生在 tools/call 或 resources/read 时                    │
└─────────────────────────────────────────────────────────────┘
```

### 8.2 完整流程

**第 1 级：传输层鉴权（API Key）**

Agent 连接到 Gateway 时，API Key 通过 HTTP Header 传递：

```
POST /mcp
Authorization: Bearer sk-abc123def456          ← API Key 放在这里
Mcp-Protocol-Version: 2025-06-18
Content-Type: application/json

{
  "method": "initialize",
  ...
}
```

Gateway 收到后立即验证：

```
① 从 Authorization Header 提取 API Key
② 查数据库：SELECT * FROM clients WHERE api_key = 'sk-abc123def456' AND status = 1
③ 查到 → 拿到 client_id，继续
④ 没查到或已禁用 → 返回 HTTP 401
```

**第 2 级：会话层鉴权（Session）**

initialize 成功通过后，Gateway 创建 Session 并返回 `Mcp-Session-Id`：

```
HTTP 200
Mcp-Session-Id: sess_550e8400-e29b-41d4-a716-446655440000

Session 内部关联的信息：
  SessionID:    sess_550e8400-...
  ClientID:     client_001
  ProtocolVer:  2025-06-18
  Status:       initialized
  CreatedAt:    2026-05-23T10:00:00Z
  ExpiresAt:    2026-05-23T10:30:00Z
```

之后的每一次 POST，Gateway 先查 Session：

```
POST /mcp
Mcp-Session-Id: sess_550e8400-...          ← 带上 Session
Authorization: Bearer sk-abc123def456       ← 仍需带 API Key

Gateway:
  ① 查 Mcp-Session-Id Header → 拿到 SessionID
  ② 从存储（内存/Redis）查找 Session
  ③ Session 不存在或已过期 → 返回 HTTP 404，要求重新 initialize
  ④ Session 存在且有效 → 继续
```

**第 3 级：业务层鉴权（Permissions）**

当收到 `tools/call` 时，查权限表：

```
tools/call { name: "account_get_info", arguments: {...} }

Gateway：
  ① 从 Session 拿到 ClientID
  ② 查数据库：
     SELECT * FROM client_tool_permissions p
     JOIN tools t ON p.tool_id = t.tool_id
     WHERE p.client_id = 'client_001'
       AND t.name = 'account_get_info'
       AND t.status = 1
  ③ 查到记录 → 允许执行
  ④ 没有记录 → 返回 MCP 错误码 -32001（鉴权失败）
```

### 8.3 为什么第 1 级和第 2 级都要传凭证？

你可能注意到：**每条 POST 都带了 API Key（Authorization Header），同时又带了 Session ID（Mcp-Session-Id Header）。这不重复吗？**

不是的，两者用途不同：

| | API Key | Session ID |
|--|---------|-----------|
| 验证什么 | "你是谁"（静态身份） | "你的会话状态"（动态上下文） |
| 谁颁发的 | 管理员在后台创建 Client 时生成 | Gateway 在 initialize 时自动创建 |
| 有效期 | 长期有效（除非管理员撤销） | 短时（比如 30 分钟无活动过期） |
| 能做什么 | 证明身份 | 记录状态（已初始化、已鉴权） |

**为什么每次都要传 API Key？**

因为 HTTP 是无状态的——Session 可能丢失（比如 Gateway 重启、Redis 宕机）。如果 Agent 只传 Session 不传 API Key，Gateway 在 Session 丢失后将无法重建上下文，只能拒绝请求。

传了 API Key 就多一层保障：Session 丢了，Gateway 可以用 API Key 重新创建 Session，不影响 Agent 继续工作。

### 8.4 对 design doc 的修正建议

你的设计文档里，鉴权的描述比较粗略：

```
① 鉴权（校验收到的 API Key）
```

实际上**鉴权的时机和位置**需要明确：

| 动作 | 你的设计 | 实际需要的鉴权 |
|------|---------|--------------|
| HTTP 连接 | 未提及 | 每条 POST 检查 API Key（传输层） |
| initialize | 未提及 | 验证 API Key，创建 Session（第 1 + 2 级） |
| tools/list | 鉴权 | 检查 Session + 权限过滤（第 2 + 3 级） |
| tools/call | 鉴权 | 检查 Session + 权限检查（第 2 + 3 级） |

另外，API Key 的存储建议**哈希后存储**，不要存明文：

```sql
CREATE TABLE clients (
    client_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    api_key_hash    VARCHAR(256) NOT NULL,   -- 哈希后的 API Key
    api_key_prefix  VARCHAR(16) NOT NULL,    -- 明文前缀，方便区分（如 "sk_abc..."）
    status          TINYINT DEFAULT 1,
    ...
);
```

Agent 拿到的是明文 `sk-abc123def456`，数据库存的是 `sha256("sk-abc123def456")`。这样即使数据库泄露，API Key 也不会直接暴露。

### 8.5 一张图看清完整鉴权流程

```
Agent                                Gateway
  │                                     │
  │ ① POST /mcp (initialize)            │
  │   Authorization: Bearer sk-xxx      │
  │ ──────────────────────────────────▶  │
  │                                     │ ② 查 clients 表验证 API Key
  │                                     │ ③ 通过 → 创建 Session
  │   HTTP 200                          │ ④ SessionID + ClientID 关联存储
  │   Mcp-Session-Id: sess_abc          │
  │ ◄──────────────────────────────────  │
  │                                     │
  │ ⑤ POST /mcp (tools/list)            │
  │   Authorization: Bearer sk-xxx      │
  │   Mcp-Session-Id: sess_abc          │
  │ ──────────────────────────────────▶  │
  │                                     │ ⑥ 验证 API Key（传输层）
  │                                     │ ⑦ 验证 Session 有效性（会话层）
  │                                     │ ⑧ 查权限表，过滤工具列表（业务层）
  │   HTTP 200 + 工具列表               │
  │ ◄──────────────────────────────────  │
```

---

## 第 9 点：MCP 的错误处理——JSON-RPC error vs `isError`

### 9.1 核心区分：两种错误

MCP 里有**两种完全不同的错误机制**，这是新人最容易搞混的地方：

```
错误类型 1：JSON-RPC 协议错误
  → 请求本身有问题，没法正常处理
  → 返回 error 对象
  → Agent 框架层就拦截了，LLM 看不到

错误类型 2：工具执行结果中的业务错误
  → 请求处理成功，但业务执行失败
  → 返回 result，但 content 中标记 isError: true
  → LLM 能看到结果，可以决定下一步
```

### 9.2 类型 1：JSON-RPC 协议错误

#### 什么时候用

Gateway 层面出了问题，**无法正常处理这个请求**的时候：

| 场景 | 做法 |
|------|------|
| 协议版本不兼容 | JSON-RPC error，-32600（Invalid Request） |
| 方法名不存在（比如发了 `tools/xxxx`） | JSON-RPC error，-32601（Method not found） |
| 参数格式错误（JSON 解析失败） | JSON-RPC error，-32700（Parse error） |
| 鉴权失败 | JSON-RPC error，自定义码 -32001 |
| 限流触发 | JSON-RPC error，自定义码 -32002 |

#### 响应格式

```json
// Agent 发了一个不存在的 Method
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "non_existent_tool",
    "arguments": {}
  }
}

// Gateway 返回 JSON-RPC 错误
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32602,
    "message": "Tool not found: non_existent_tool",
    "data": {
      "available_tools": ["account_get_info", "order_create", "refund"]
    }
  }
}
```

**关键特征**：响应里有 `error` 字段，没有 `result` 字段。MCP 客户端收到后直接用这个错误信息通知用户，**不会交给 LLM 处理**。

### 9.3 类型 2：业务错误（`isError`）

#### 什么时候用

**工具调用本身成功了**（Gateway 成功发出了 HTTP 请求，收回了响应），但执行结果不是成功的——比如下游返回了 4xx。

**核心原则**：只要 Gateway 成功完成了参数映射、发出了 HTTP 请求、收到了响应，就属于"工具执行成功"，应该走 `result` 而不是 `error`。

#### 响应格式

```json
// tools/call 成功执行，但下游返回了 400 Bad Request
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"error\":\"invalid_user_id\",\"message\":\"用户 ID 格式错误\"}"
      }
    ],
    "isError": true
  }
}
```

**关键特征**：
- 响应里仍然有 `result`，没有 `error`
- `result` 里面多了 `isError: true`
- `content` 里是下游返回的原始错误信息（但要**脱敏**，不能暴露堆栈）

Agent 的 LLM 看到这个结果后，可以自己决定怎么处理：重试、换参数、或者告诉用户"出错了"。

### 9.4 两种错误对比

| | JSON-RPC error | `isError: true` |
|--|---------------|----------------|
| 响应结构 | 有 `error` 字段，无 `result` | 有 `result` 字段，无 `error` |
| Gateway 做了什么 | 没能完成处理 | 正常完成了处理 |
| 谁看到了 | Agent 框架层（LLM 看不到） | LLM 能看到 |
| LLM 能否处理 | ❌ 不能，直接报错给用户 | ✅ 能，可以决定下一步 |
| 类比 HTTP | 502 Bad Gateway（网关自身出问题） | 400 Bad Request（业务上拒绝） |

### 9.5 一条区分原则

> **做决定了再做错了 → `isError`**
> **还没决定就出问题了 → JSON-RPC error**

```
Agent 发来 tools/call

  ↓

Gateway 解析 JSON-RPC
  └─ JSON 格式错误              → JSON-RPC error（还没开始干活呢）
  └─ 方法名无效                 → JSON-RPC error
  └─ 工具名不存在              → JSON-RPC error（复用 -32602，message 区分）

  ↓

Gateway 执行路由（这步"决定了"）
  └─ 鉴权失败                  → JSON-RPC error（Gateway 无法继续）
  └─ 限流触发                  → JSON-RPC error（Gateway 无法继续）

  ↓

Gateway 发出 HTTP 请求（这步"决定了"）
  └─ 下游超时/5xx              → JSON-RPC error（Gateway 自己也没拿到结果）
  └─ 下游返回 4xx              → `isError: true`（拿到结果了，只是业务拒绝了）
  └─ 下游返回 2xx              → 正常 result
```

### 9.6 对 design doc 的修正建议

你的设计文档中的错误表：

| 场景 | 返回给 Agent |
|------|-------------|
| API Key 无效 | MCP 错误码 -32001 鉴权失败 |
| 限流触发 | MCP 错误码 -32002 请求过于频繁 |
| 参数校验失败 | MCP 错误码 -32003 参数错误 |
| 下游 4xx | MCP 错误码 -32004 + 错误描述 |
| 下游 5xx / 超时 | MCP 错误码 -32005 下游服务异常 |
| 工具不存在 | MCP 错误码 -32602 工具不存在 |

修正后的错误表：

| 场景 | 响应方式 | 状态码 | 说明 |
|------|---------|--------|------|
| API Key 无效 | JSON-RPC error | -32001 | Gateway 无法继续 |
| 限流触发 | JSON-RPC error | -32002 | Gateway 无法继续 |
| MCP 参数必填缺失 | JSON-RPC error | -32602 | 复用标准 Invalid params |
| 工具不存在 | JSON-RPC error | -32602 | 复用标准码，message 区分 |
| JSON 解析失败 | JSON-RPC error | -32700 | 标准 Parse error |
| 不存在的 method | JSON-RPC error | -32601 | 标准 Method not found |
| **下游 4xx** | `result` + `isError: true` | 无 | Gateway 成功获取了下游响应 |
| **下游 5xx / 超时** | JSON-RPC error | -32005 | Gateway 没能获得有效响应 |

---

## 第 10 点：限流实现——Redis 滑动窗口 + Lua 脚本方案

### 10.1 三条原则

MCP Gateway 的限流和传统 API 网关的限流本质一样，但要先明确三个约束：

```
① 多实例部署      → 限流状态必须在 Redis 共享，不能存在本地内存
② 毫秒级判断      → 每个 tools/call 都要走限流，延迟必须 < 5ms
③ 三级粒度        → 客户端 / 接口 / 组合，分别计数，一个触发就拒绝
```

### 10.2 为什么用滑动窗口，不用计数器

**简单计数器（INCR + EXPIRE）** 有边界问题：

```
时间轴（每秒限 10 次）：
  0.9s 时：来了 10 个请求（计数器归零前一瞬间用满了额度）
  1.1s 时：来了 10 个请求（新的一秒又用满了额度）

实际 0.9s ~ 1.1s 这 0.2s 内通过了 20 个请求，超过限流阈值
```

**滑动窗口** 解决这个问题：始终看"当前时间往前 1 秒内"总共多少请求。

### 10.3 Redis 滑动窗口实现

#### 数据结构

对每个限流维度，用一个 Redis **有序集合（Sorted Set）**：

```
Key:    rate_limit:client:client_001
Score:  1684850400123（毫秒时间戳）
Member: req_uuid_001（请求唯一 ID，确保集合元素唯一）

Key:    rate_limit:tool:account_get_info
Score:  1684850400123
Member: req_uuid_002
```

#### Lua 脚本（原子操作）

整个限流检查用一条 Lua 脚本完成，避免并发问题：

```lua
-- KEYS[1] = Redis key，如 rate_limit:client:client_001
-- ARGV[1] = 当前毫秒时间戳
-- ARGV[2] = 窗口大小（毫秒），如 1000
-- ARGV[3] = 最大请求数
-- ARGV[4] = 本次请求的唯一 ID

local window_start = tonumber(ARGV[1]) - tonumber(ARGV[2])

-- 1. 清除窗口外的旧记录
redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, window_start)

-- 2. 统计窗口内请求数
local current_count = redis.call('ZCARD', KEYS[1])

-- 3. 判断是否超限
if current_count >= tonumber(ARGV[3]) then
    return 0  -- 拒绝
end

-- 4. 记录本次请求
redis.call('ZADD', KEYS[1], ARGV[1], ARGV[4])
redis.call('EXPIRE', KEYS[1], tonumber(ARGV[2]) / 1000 + 1)

return 1  -- 通过
```

调用方式：

```
EVALSHA <script_sha> 1 rate_limit:client:client_001 1684850400123 1000 100 req_abc_123
```

#### 不用令牌桶？

令牌桶是次优选择：

| 算法 | 优点 | 缺点 |
|------|------|------|
| 滑动窗口 | 精确反映"过去 1 秒内的实际请求数" | 需要清理旧数据，略耗内存 |
| 令牌桶 | 允许突发流量（积累的令牌一次用完） | 不适合"平滑限流"的 API 网关场景 |

对于 MCP Gateway，Agent 调工具的请求频率相对稳定，滑动窗口更合适。

### 10.4 三级限流的调用链

每个 `tools/call` 进来，Gateway 依次调用三次 Lua 脚本：

```go
func checkRateLimit(clientID string, toolName string) error {
    now := time.Now().UnixMilli()
    reqID := uuid.New().String()

    // ① 客户端级别
    ok := redis.EvalSha(scriptSHA, 1,
        "rate_limit:client:" + clientID,
        now, 1000, clientRateLimit, reqID)
    if !ok { return ErrRateLimit("客户端请求过于频繁") }

    // ② 接口级别
    ok = redis.EvalSha(scriptSHA, 1,
        "rate_limit:tool:" + toolName,
        now, 1000, toolRateLimit, reqID)
    if !ok { return ErrRateLimit("接口请求过于频繁") }

    // ③ 组合级别
    ok = redis.EvalSha(scriptSHA, 1,
        "rate_limit:combo:" + clientID + ":" + toolName,
        now, 1000, comboRateLimit, reqID)
    if !ok { return ErrRateLimit("组合请求过于频繁") }

    return nil
}
```

**注意**：如果三级中有一级的配置为空（没设置限流），就直接跳过该级的检查。

### 10.5 限流配置从哪里来

原 design doc 把限流配置放在各个表的 `config` JSON 字段里。JSON 字段的查询和更新不方便。建议独立成一张表：

```sql
CREATE TABLE rate_limit_rules (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    scope_type  VARCHAR(16) NOT NULL,    -- 'client' / 'tool' / 'combo'
    scope_id    VARCHAR(128) NOT NULL,    -- client_001 / account_get_info / client_001:account_get_info
    max_requests INT NOT NULL,           -- 窗口内最大请求数
    window_ms   INT NOT NULL DEFAULT 1000, -- 窗口大小（毫秒）
    status      TINYINT DEFAULT 1,
    updated_at  DATETIME
);

CREATE UNIQUE INDEX idx_scope ON rate_limit_rules(scope_type, scope_id);
```

Gateway 启动时用一次查询加载所有规则到本地缓存：

```go
type RateLimiter struct {
    rules map[string]RateLimitRule  // key = "client:client_001"
    redis *redis.Client
}

func (r *RateLimiter) LoadRules() {
    rows := db.Query("SELECT * FROM rate_limit_rules WHERE status = 1")
    for rows.Next() {
        key := rows.scope_type + ":" + rows.scope_id
        r.rules[key] = RateLimitRule{
            MaxRequests: rows.max_requests,
            WindowMs:    rows.window_ms,
        }
    }
}
```

### 10.6 限流的 HTTP 响应

当限流触发时，还应该在 HTTP 响应头里告诉 Agent 限流信息：

```
HTTP 200
Content-Type: application/json
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1684850401

Body:
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32002,
    "message": "Rate limit exceeded",
    "data": {
      "scope": "client:client_001",
      "limit": 100,
      "remaining": 0,
      "reset_at": 1684850401
    }
  }
}
```

### 10.7 对 design doc 的修正

原设计：
> **存储**：Redis（INCR + EXPIRE 或 Lua 脚本）

建议改为 **Lua 脚本 + Sorted Set**（滑动窗口）：

| 方案 | 边界精度 | 原子性 | 推荐 |
|------|---------|--------|------|
| INCR + EXPIRE | 有窗口边界毛刺 | ✅ | ❌ |
| Lua + Sorted Set | ✅ 精确滑动窗口 | ✅ | ✅ |
| Lua + List（LPUSH + LTRIM） | ✅ 精确 | ✅ | 略差于 ZSet |

原设计：
> **策略**：三级限流依次判断，任一触发即拒绝

这条是正确的，不用改。

---

## 第 11 点：Streamable HTTP 的两种响应模式——什么时候回 JSON，什么时候回 SSE

### 11.1 问题背景

回顾第 3 点：Agent 的每条消息都是一个独立的 HTTP POST。那 Gateway 怎么**返回**响应？

MCP 2025-06-18 规范定义了两种响应方式：

```
Agent POST 一条消息过来

Gateway 有两种选择：

模式 A（即时响应）：HTTP 200 + Content-Type: application/json
  → 响应体就是完整的 JSON-RPC 响应
  → 一条请求，一条响应，简单直接

模式 B（流式响应）：HTTP 200 + Content-Type: text/event-stream
  → 先开一个 SSE 流
  → 后面可以推多条消息
  → 适合需要推送额外信息的场景
```

### 11.2 模式 A：即时 JSON 响应

#### 什么时候用

绝大多数场景。initialize、tools/list、tools/call 都可以用这种方式。

```
Agent ── POST (tools/list) ──▶ Gateway
                                   │
                                   │ 查询数据库 → 组装列表
                                   │
Agent ◀── HTTP 200 ───────────────│
       Content-Type: application/json
       {
         "jsonrpc": "2.0",
         "id": 3,
         "result": {
           "tools": [...]
         }
       }
```

**特点**：
- 一个 POST → Gateway 处理 → 一个响应
- 请求和响应一一对应
- 最简单，延迟最低

#### Go 实现

```go
func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    msg := parseJSONRPC(body)

    result := routeMessage(msg)

    // 模式 A：直接返回 JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(result)
}
```

### 11.3 模式 B：SSE 流式响应

#### 什么时候用

需要 Gateway **在处理过程中主动推送额外消息**的时候。

**场景 1：tools/call 处理耗时较长**

Gateway 先告诉 Agent "正在处理"，处理完再给结果：

```
Agent ── POST (tools/call) ──▶ Gateway
                                   │
Agent ◀── SSE stream opens ───────│
         data: {"jsonrpc":"2.0",   ← 进度通知
                "method":"notifications/progress",
                "params":{"progress":0.5}}
                                   │  ...还在处理...
         data: {"jsonrpc":"2.0",   ← 进度通知
                "method":"notifications/progress",
                "params":{"progress":0.8}}
                                   │  处理完了
         data: {"jsonrpc":"2.0",   ← 最终结果
                "id":3,
                "result":{"content":[...]}}
Agent ◀── 流关闭 ───────────────────
```

**场景 2：Gateway 需要向 Agent 发请求（反向请求）**

```
Agent ── POST (tools/call) ──▶ Gateway
                                   │  Gateway 想确认用户意愿
Agent ◀── SSE stream ─────────────│
         data: {"jsonrpc":"2.0",   ← Gateway 反过来向 Agent 请求确认
                "id":100,
                "method":"elicitation/create",
                "params":{"prompt":"确认要删除用户 123 吗？"}}
                                   │  Agent 通过另一个 POST 回复
Agent ── POST (回复 elicitation) ─▶
                                   │  收到确认
Agent ◀── SSE stream ─────────────│
         data: {"jsonrpc":"2.0",   ← 最终结果
                "id":3,
                "result":{"content":[...]}}
```

#### Go 实现

```go
func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    msg := parseJSONRPC(body)

    if needsStreaming(msg) {
        // 模式 B：SSE 流式响应
        flusher, ok := w.(http.Flusher)
        if !ok {
            http.Error(w, "streaming not supported", 500)
            return
        }

        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.WriteHeader(200)

        go func() {
            // 发送进度通知
            sendSSE(w, flusher, map[string]interface{}{
                "jsonrpc": "2.0",
                "method":  "notifications/progress",
                "params":  map[string]float64{"progress": 0.5},
            })

            // ... 处理 ...

            // 发送最终结果
            sendSSE(w, flusher, map[string]interface{}{
                "jsonrpc": "2.0",
                "id":      3,
                "result":  result,
            })
        }()
    } else {
        // 模式 A：直接返回 JSON
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(result)
    }
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, data interface{}) {
    jsonBytes, _ := json.Marshal(data)
    fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
    flusher.Flush()
}
```

### 11.4 Agent 怎么决定用哪种模式

Agent 通过 **Accept Header** 告诉 Gateway 它支持什么：

```
Accept: application/json, text/event-stream
  → 两种都支持，Gateway 自己选

Accept: application/json
  → 只支持即时 JSON 响应，别开 SSE 流
```

Gateway 根据 Agent 的 Accept 和自身处理需要来选择模式。

### 11.5 SSE 流结束

SSE 流发完所有消息后，Gateway 关闭连接。不发特别标记，关闭 HTTP 连接即可。

### 11.6 对 design doc 的影响

对于 V1 阶段，**模式 A（即时 JSON）就够了**：

| 场景 | V1 是否需要 SSE |
|------|----------------|
| initialize | ❌ 即时返回就行 |
| tools/list | ❌ 查个数据库，毫秒级返回 |
| tools/call（HTTP 协议转换） | ❌ 发 HTTP 请求等响应，一次性返回 |
| 进度通知 | V3+ 才考虑 |
| elicitation（反向请求） | V4+ 才考虑 |

V1 可以简化成：

```
所有 Agent 的 POST → Gateway 处理 → 一律 Content-Type: application/json 返回
GET 端点 → 打开 SSE 流，用于接收 listChanged 推送
```

---

## 第 12 点：Admin API 设计——管理后台与 MCP 协议的联动

### 12.1 Admin API 的定位

Admin API 和 MCP 协议端点**跑在同一个进程里**，不同路径：

```
同一个 Gateway 进程

/mcp         → MCP 协议端点（Agent 连接）
/admin/api/v1/* → Admin REST API（管理员操作）

两个路径隔离，互不影响
```

### 12.2 Admin API 管理什么

| 资源 | 对应表 | 主要操作 |
|------|--------|---------|
| Clients | clients | 注册/注销、生成/轮换 API Key |
| Projects | projects | 登记下游 HTTP 服务、管理凭证 |
| Tools | tools + tool_params | 注册/更新工具定义 |
| Permissions | client_tool_permissions | 授权 Client 访问 Tool |
| Rate Limits | rate_limit_rules | 配置限流规则 |

### 12.3 Admin API 端点设计

```
基础路径：/admin/api/v1

┌────────────┬──────────────────────────────┬──────────────────────┐
│ 方法        │ 路径                         │ 说明                  │
├────────────┼──────────────────────────────┼──────────────────────┤
│ POST       │ /clients                     │ 创建调用方            │
│ GET        │ /clients                     │ 列表查询              │
│ GET        │ /clients/:id                 │ 查看详情              │
│ PUT        │ /clients/:id                 │ 更新                  │
│ DELETE     │ /clients/:id                 │ 删除（软删除）        │
│ POST       │ /clients/:id/rotate-key      │ 重新生成 API Key      │
├────────────┼──────────────────────────────┼──────────────────────┤
│ POST       │ /projects                    │ 创建下游项目          │
│ GET        │ /projects                    │ 列表查询              │
│ GET        │ /projects/:id                │ 查看详情              │
│ PUT        │ /projects/:id                │ 更新                  │
│ DELETE     │ /projects/:id                │ 删除（软删除）        │
├────────────┼──────────────────────────────┼──────────────────────┤
│ POST       │ /tools                       │ 创建工具定义          │
│ GET        │ /tools                       │ 列表查询              │
│ GET        │ /tools/:id                   │ 查看详情              │
│ PUT        │ /tools/:id                   │ 更新                  │
│ DELETE     │ /tools/:id                   │ 删除（软删除）        │
│ PUT        │ /tools/:id/status            │ 启用/禁用             │
├────────────┼──────────────────────────────┼──────────────────────┤
│ POST       │ /permissions                 │ 授权                  │
│ GET        │ /permissions                 │ 查询权限              │
│ DELETE     │ /permissions/:id             │ 取消授权              │
├────────────┼──────────────────────────────┼──────────────────────┤
│ POST       │ /rate-limits                 │ 创建限流规则          │
│ GET        │ /rate-limits                 │ 查询限流规则          │
│ PUT        │ /rate-limits/:id             │ 更新限流规则          │
│ DELETE     │ /rate-limits/:id             │ 删除限流规则          │
└────────────┴──────────────────────────────┴──────────────────────┘
```

### 12.4 创建工具的完整流程（最复杂的操作）

创建工具同时涉及 `tools` 和 `tool_params` 两张表：

```
POST /admin/api/v1/tools
Body:
{
  "project_id": "proj_account_001",
  "name": "account_get_info",
  "description": "根据用户 ID 查询账户基本信息",
  "http_method": "GET",
  "url_template": "/api/v1/accounts/{user_id}",
  "timeout_ms": 5000,
  "params": [
    {
      "name": "user_id",
      "param_type": "string",
      "location": "path",
      "required": true,
      "description": "用户 ID"
    },
    {
      "name": "fields",
      "param_type": "string",
      "location": "query",
      "required": false,
      "default_value": "name,email",
      "description": "返回字段列表"
    }
  ]
}
```

Gateway 内部处理：

```
① 验证 project_id 存在
② 验证 name 不重复
③ 验证 params 中 path 类型的参数全部在 url_template 中有 {placeholder}
   （避免误配：params 说 user_id 是 path 类型，但 URL 里没有 {user_id}）
④ INSERT INTO tools
⑤ INSERT INTO tool_params
⑥ 通知 MCP 端点推 listChanged
```

### 12.5 Admin API 和 MCP 端点的联动

**联动点 1：创建/更新工具后推送 listChanged**

```
管理员                     Admin API                  MCP 端点                 Agent
  │                          │                          │                       │
  │ POST /tools              │                          │                       │
  │ ──────────────────────▶  │                          │                       │
  │                          │  写入数据库               │                       │
  │                          │  通知 MCP "工具变了"      │                       │
  │                          │ ──────────────────────▶  │                       │
  │                          │                          │  SSE 推 listChanged   │
  │                          │                          │ ────────────────────▶ │
  │                          │                          │                       │
  │ ◀── 201 Created ────────  │                          │                       │
```

实现：Admin API 和 MCP 端点通过 **channel 或 Redis Pub/Sub** 通信：

```go
// admin_handler.go — 创建工具后通知 MCP
func (h *AdminHandler) CreateTool(w http.ResponseWriter, r *http.Request) {
    tool := parseBody(r)
    db.InsertTool(tool)

    h.toolChangeNotifier <- ToolChangeEvent{Action: "created", ToolID: tool.ID}

    w.WriteHeader(201)
    json.NewEncoder(w).Encode(tool)
}
```

```go
// mcp_handler.go — 接收到通知后推 SSE
func (h *MCPHandler) WatchToolChanges() {
    for range h.toolChangeNotifier {
        h.sseHub.Broadcast(map[string]string{
            "jsonrpc": "2.0",
            "method":  "notifications/tools/list_changed",
        })
    }
}
```

**联动点 2：创建 Client 后生成 API Key**

```
POST /admin/api/v1/clients
Response:
{
  "client_id": "client_abc_001",
  "name": "客服团队",
  "api_key": "sk_live_abc123...",    ← 只有创建时返回一次明文
  "api_key_hash": "sha256$...",
  "created_at": "2026-05-23T10:00:00Z"
}
```

API Key 只在创建时返回一次明文，丢失只能重新生成（rotate key）。

**联动点 3：启用/禁用工具**

```
PUT /admin/api/v1/tools/:id/status
Body: { "status": 0 }  // 禁用

内部处理：
  ① UPDATE tools SET status = 0 WHERE tool_id = ?
  ② 通知 MCP 推 listChanged
  ③ 下次 Agent 发 tools/list 就不会再返回这个工具了
```

### 12.6 Admin API 的鉴权

Admin API 必须**独立鉴权**，和 MCP 的 API Key 分开：

```
MCP API Key：     Agent 使用，格式 sk_live_xxx
Admin API Key：   管理员使用，格式 admin_xxx，有更高权限

配置方式：
  config.yaml:
    admin:
      api_key: "admin_xxxx"           # 管理员密钥
      allowed_ips: ["10.0.0.0/8"]     # IP 白名单（可选）
```

简单场景可以只验证一个固定的管理员密钥，正式环境建议集成 OAuth 2.0。

### 12.7 对 design doc 的影响

补充几点：

1. **参数映射校验**：创建工具时校验 `params.location = 'path'` 的参数在 `url_template` 中是否存在对应的 `{placeholder}`
2. **联动通知**：Admin API 修改数据后必须通知 MCP 端点，否则 Agent 看到的工具列表是过期的
3. **Admin API 单独鉴权**：不要复用 MCP 的 API Key，否则 Agent token 泄露后攻击者也能管理后台

---

## 第 13 点：健康检查与熔断——下游服务不可用时 Gateway 怎么做降级

### 13.1 问题场景

Gateway 调用下游 HTTP 服务，但下游可能出各种问题：

```
下游服务挂了         → Gateway 调用超时，Agent 等 5 秒后收到错误
下游服务响应慢       → Gateway 线程被占用，影响其他请求
下游服务返回 500     → Gateway 每次都重试，越等越慢
```

不做保护的话，一次下游故障可能拖垮整个 Gateway。

### 13.2 两个概念

| 概念 | 做什么 | 类比 |
|------|--------|------|
| **健康检查** | 定期检查下游是否活着 | 每隔几秒 ping 一下 |
| **熔断** | 检测到下游频繁失败后，暂时不调它 | 知道下游不行了，干脆不发了 |

健康检查是主动探测，熔断是被动检测。

### 13.3 健康检查

最简单的实现：定期对下游健康检查端点发请求。

```
┌──────────┐  每隔 10 秒 GET /health    ┌──────────┐
│ Gateway  │ ─────────────────────────▶  │ 下游服务  │
│          │ ◀── HTTP 200 {"status":"ok"}│          │
└──────────┘                             └──────────┘
```

如果连续 N 次检查失败，把下游标记为 **unhealthy**：

```
Gateway 的 projects 内存状态：
  proj_account_001:
    base_url: http://account-service:8080
    healthy: false         ← 标记不可用
    last_check: 1684850400
    consecutive_failures: 3
```

在 `projects` 表加两个字段就够了：

```sql
ALTER TABLE projects ADD COLUMN (
    health_check_url  VARCHAR(512),     -- 如 /health
    health_check_interval INT DEFAULT 10, -- 检查间隔（秒）
    healthy           TINYINT DEFAULT 1
);
```

### 13.4 熔断

熔断**根据实际调用的成功率**自动判断。

**三种状态**：

```
关闭（Closed）      → 正常，所有请求都发
                      但 Gateway 统计最近请求的失败率

打开（Open）        → 失败率超过阈值（比如 50%）
                      直接拒绝请求，不调下游
                      等待一段时间（比如 30 秒）

半开（Half-Open）   → 过了等待时间
                      Gateway 放少量请求过去试试
                      成功了 → 回到 Closed
                      又失败了 → 回到 Open 重新计时
```

**状态转换**：

```
                   失败率 > 50%
   ┌──────────┐ ────────────────▶ ┌──────────┐
   │  Closed   │                  │   Open    │
   └──────────┘ ◀──────────────── └──────────┘
                   超时后放试探请求
                         │
                         │ 试探请求成功 → Closed
                         │ 试探请求失败 → Open（重新计时）
                         ▼
                    ┌──────────┐
                    │ Half-Open │
                    └──────────┘
```

### 13.5 Go 极简实现

```go
type CircuitBreaker struct {
    mu               sync.RWMutex
    state            string            // "closed" / "open" / "half-open"
    failures         int
    successThreshold int               // 半开状态下连续成功几次后关闭
    failureThreshold int               // 连续失败几次后打开
    timeout          time.Duration     // 打开后等多久进入半开
    lastFailureTime  time.Time
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()

    switch state {
    case "closed":
        return true
    case "open":
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.mu.Lock()
            cb.state = "half-open"
            cb.mu.Unlock()
            return true // 放一个试探请求
        }
        return false
    case "half-open":
        return true // 放行试探请求
    default:
        return true
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures = 0
    if cb.state == "half-open" {
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = "closed"
            cb.successCount = 0
        }
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures++
    cb.lastFailureTime = time.Now()
    if cb.failures >= cb.failureThreshold {
        cb.state = "open"
    }
}
```

### 13.6 在 tools/call 链路中嵌入

```go
func (uc *CallToolUseCase) Execute(ctx context.Context, req ToolCallRequest) (*ToolResult, error) {
    // ① 查工具定义
    tool := uc.toolRepo.FindByName(req.Name)

    // ② 查下游的熔断器
    cb := uc.circuitBreakerRepo.Get(tool.ProjectID)

    // ③ 熔断检查
    if !cb.Allow() {
        return nil, ErrCircuitOpen("下游服务正在熔断中，请稍后重试")
    }

    // ④ 参数映射 + 发 HTTP 请求
    result, err := uc.httpClient.Call(ctx, tool, req.Arguments)

    // ⑤ 记录结果到熔断器
    if err != nil {
        cb.RecordFailure()
        return nil, err
    }
    cb.RecordSuccess()
    return result, nil
}
```

熔断触发时，Agent 收到：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32005,
    "message": "下游服务正在熔断中，请稍后重试",
    "data": {
      "project": "account-service",
      "retry_after_seconds": 25
    }
  }
}
```

### 13.7 健康检查 + 熔断的配合

两者互补：

```
Health Check                     Circuit Breaker
─────────────────────────         ────────────────────────
主动探测                        被动检测
定期 ping 下游                   根据实际请求成功率
提前发现问题                     请求失败了才反应
适合标记"这个项目有问题"         适合保护 Gateway 不被拖垮
```

**配合策略**：

| 下游状态 | tools/list | tools/call |
|---------|-----------|-----------|
| 健康 | 返回工具 | 正常调用 |
| 健康检查失败 | 依然返回工具 | 正常调用，走熔断计数 |
| 熔断打开 | 依然返回工具 | 拒绝请求，提示"熔断中" |

这样 Agent 依然能看到工具有哪些参数，只是调用时会被告知暂时不可用。

---

## 第 14 点：代码架构落地——Go 项目中 DDD 分层的实际组织方式

### 14.1 先搞清楚一个问题

design doc 已经画了分层结构（Interface → Application → Domain → Infrastructure），但落地时经常遇到一个困惑：**"接口定义在哪里？同一个包里的文件怎么放？"**

这一章回答这个问题——把分层结构**落地成具体的 Go 文件和包**。

### 14.2 目录结构

```
cmd/
└── gateway/
    └── main.go              ← 入口：创建依赖、启动 HTTP 服务

internal/
├── interface/               ← 接住"外面来的请求"
│   ├── mcp/
│   │   ├── handler.go       ← HTTP Handler：解析 POST/GET，路由到 use case
│   │   ├── session.go       ← Session 管理（创建、校验、过期）
│   │   └── sse.go           ← SSE 推送 hub（listChanged 等）
│   │
│   └── admin/
│       └── handler.go       ← Admin REST API 路由注册
│
├── application/             ← "指挥别人干活"
│   ├── list_tools.go        ← ListToolsUseCase
│   ├── call_tool.go         ← CallToolUseCase
│   ├── initialize.go        ← InitializeUseCase
│   └── manage_client.go     ← ManageClientUseCase
│
├── domain/                  ← "业务规则，纯逻辑"
│   ├── param_mapper.go      ← MCP 参数 → HTTP 参数的映射逻辑
│   ├── response_handler.go  ← HTTP 响应 → MCP 响应的转换逻辑
│   ├── url_template.go      ← {user_id} → 123 的模板替换
│   └── schema_builder.go    ← tool_params → JSON Schema 的转换
│
├── infrastructure/          ← "跟外部世界打交道"
│   ├── db/
│   │   ├── mysql.go         ← MySQL 连接
│   │   ├── tool_repo.go     ← ToolRepository 的 MySQL 实现
│   │   ├── client_repo.go   ← ClientRepository 的 MySQL 实现
│   │   └── project_repo.go  ← ProjectRepository 的 MySQL 实现
│   │
│   ├── ratelimit/
│   │   ├── interface.go     ← RateLimiter 接口定义
│   │   └── redis.go         ← Redis 滑动窗口实现
│   │
│   ├── circuitbreaker/
│   │   ├── circuit.go       ← 熔断器实现
│   │   └── registry.go      ← 熔断器注册表
│   │
│   ├── http_client.go       ← 实际的 HTTP 调用（带超时、重试）
│   ├── redis.go             ← Redis 连接
│   └── audit.go             ← 审计日志
│
├── config/
│   └── config.go            ← 配置加载（yaml → struct）
│
└── pkg/                     ← 可复用的公共代码
    ├── jsonrpc/
    │   ├── message.go       ← JSON-RPC 2.0 消息结构体
    │   └── error.go         ← 预定义的错误码
    └── mcp/
        ├── tools.go         ← MCP 工具相关的类型定义
        ├── resources.go     ← MCP 资源相关的类型定义
        └── initialize.go    ← MCP 初始化相关的类型定义
```

### 14.3 `pkg/jsonrpc/message.go`——JSON-RPC 基础消息

这是最基本的数据结构，所有层都需要引用：

```go
package jsonrpc

type Request struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Result  interface{} `json:"result,omitempty"`
    Error   *Error      `json:"error,omitempty"`
}

type Error struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type Notification struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// 标准错误码
const (
    ErrParse           = -32700
    ErrInvalidRequest  = -32600
    ErrMethodNotFound  = -32601
    ErrInvalidParams   = -32602
    ErrInternal        = -32603
    ErrAuthFailed      = -32001
    ErrRateLimit       = -32002
    ErrServiceDown     = -32005
)
```

### 14.4 `internal/interface/mcp/handler.go`——总入口

这是 Agent 请求最先到达的地方。**只做路由，不做业务逻辑**：

```go
package mcp

func NewMCPHandler(sm *session.Manager, uc *application.UseCases) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /mcp", func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        req := parseJSONRPC(body)

        sessionID := r.Header.Get("Mcp-Session-Id")

        var result interface{}
        var err error

        switch req.Method {
        case "initialize":
            result, sessionID, err = uc.Initialize.Execute(sessionID, req.Params)
            if sessionID != "" {
                w.Header().Set("Mcp-Session-Id", sessionID)
            }

        case "tools/list":
            result, err = uc.ListTools.Execute(sessionID)

        case "tools/call":
            result, err = uc.CallTool.Execute(sessionID, req.Params)

        default:
            err = jsonrpc.NewError(jsonrpc.ErrMethodNotFound, "unknown method")
        }

        if err != nil {
            writeError(w, req.ID, err)
        } else {
            writeResult(w, req.ID, result)
        }
    })
    return mux
}
```

### 14.5 `internal/application/call_tool.go`——用例编排

这是 orchestrator（编排者），**调各种 repo 和 domain 逻辑，不关心具体实现**：

```go
package application

type CallToolUseCase struct {
    sessionRepo  SessionRepository      // interface
    toolRepo     ToolRepository         // interface
    paramMapper  *domain.ParamMapper    // domain 逻辑
    urlBuilder   *domain.URLTemplateBuilder
    httpClient   *infra.HTTPClient      // infrastructure
    cbRegistry   *circuitbreaker.Registry
    rateLimiter  ratelimit.RateLimiter  // interface
    auditLogger  AuditLogger            // interface
}

func (uc *CallToolUseCase) Execute(
    ctx context.Context, sessionID string, params map[string]interface{},
) (*ToolResult, error) {

    // ① 校验 session
    session, err := uc.sessionRepo.FindByID(sessionID)
    if err != nil {
        return nil, jsonrpc.NewError(jsonrpc.ErrAuthFailed, "invalid session")
    }

    // ② 解析请求
    callReq := parseToolCallParams(params)

    // ③ 查工具定义
    tool, err := uc.toolRepo.FindByName(callReq.Name)
    if err != nil {
        return nil, jsonrpc.NewError(jsonrpc.ErrInvalidParams, "tool not found")
    }

    // ④ 权限检查
    if !uc.toolRepo.CheckPermission(session.ClientID, tool.ID) {
        return nil, jsonrpc.NewError(jsonrpc.ErrAuthFailed, "no permission")
    }

    // ⑤ 限流检查
    if err := uc.rateLimiter.Check(session.ClientID, tool.Name); err != nil {
        return nil, err
    }

    // ⑥ 熔断检查
    cb := uc.cbRegistry.Get(tool.ProjectID)
    if !cb.Allow() {
        return nil, jsonrpc.NewError(jsonrpc.ErrServiceDown, "circuit breaker open")
    }

    // ⑦ 参数映射（domain 纯逻辑）
    mapped := uc.paramMapper.Map(tool.Params, callReq.Arguments)

    // ⑧ 构建 URL（domain 纯逻辑）
    url := uc.urlBuilder.Build(tool.URLTemplate, mapped.PathParams, mapped.QueryParams)

    // ⑨ 发 HTTP 请求（infrastructure）
    resp, err := uc.httpClient.Do(ctx, &infra.HTTPRequest{
        Method: tool.HTTPMethod,
        URL:    tool.BaseURL + url,
        Headers: mapped.Headers,
        Body:   mapped.Body,
        Timeout: tool.TimeoutMs,
    })

    // ⑩ 记录熔断结果
    if err != nil {
        cb.RecordFailure()
        return nil, jsonrpc.NewError(jsonrpc.ErrServiceDown, "downstream error")
    }
    cb.RecordSuccess()

    // ⑪ 处理响应（domain 纯逻辑）
    result := uc.respHandler.Handle(resp)

    // ⑫ 审计日志（异步）
    go uc.auditLogger.Log(&AuditEntry{
        ClientID:   session.ClientID,
        ToolName:   tool.Name,
        Args:       callReq.Arguments,
        Response:   result,
        Latency:    time.Since(start),
    })

    return result, nil
}
```

### 14.6 接口/实现的分离方式——由调用方定义接口

```
application/call_tool.go  ← 调用方，定义它需要的接口
    type ToolRepository interface {
        FindByName(name string) (*domain.Tool, error)
        CheckPermission(clientID string, toolID int64) (bool, error)
    }

infrastructure/db/tool_repo.go  ← 实现方
    type MySQLToolRepo struct { db *sql.DB }
    func (r *MySQLToolRepo) FindByName(name string) (*domain.Tool, error) {
        // 实际的 SQL 查询
    }
```

好处：application 层不依赖任何具体的数据库实现，换数据库只需在 main.go 换一个实现。

### 14.7 main.go 组装一切

这是项目里**唯一**知道所有具体实现的地方：

```go
func main() {
    // 加载配置
    cfg := config.Load("config.yaml")

    // 基础设施
    mysqlDB := infra.NewMySQL(cfg.Database)
    redisClient := infra.NewRedis(cfg.Redis)

    // Repo 实现
    toolRepo := db.NewMySQLToolRepo(mysqlDB)
    clientRepo := db.NewMySQLClientRepo(mysqlDB)
    sessionRepo := mcp.NewSessionRepo(redisClient)

    // 限流器
    rateLimiter := ratelimit.NewRedisRateLimiter(redisClient)
    rateLimiter.LoadRules(mysqlDB)

    // 熔断器
    cbRegistry := circuitbreaker.NewRegistry()

    // Domain 层（纯逻辑，无需依赖）
    paramMapper := domain.NewParamMapper()
    urlBuilder := domain.NewURLTemplateBuilder()
    respHandler := domain.NewResponseHandler()

    // HTTP 客户端
    httpClient := infra.NewHTTPClient()

    // 审计日志
    auditLogger := infra.NewAuditLogger(cfg.Kafka)

    // 应用层
    useCases := &application.UseCases{
        Initialize: application.NewInitializeUseCase(sessionRepo, clientRepo),
        ListTools:  application.NewListToolsUseCase(sessionRepo, toolRepo),
        CallTool:   application.NewCallToolUseCase(
            sessionRepo, toolRepo, paramMapper, urlBuilder, respHandler,
            httpClient, rateLimiter, cbRegistry, auditLogger,
        ),
    }

    // 接口层
    mcpHandler := mcp.NewMCPHandler(sessionRepo, useCases)
    adminHandler := admin.NewAdminHandler(toolRepo, clientRepo, projectRepo)

    // HTTP 服务器
    mux := http.NewServeMux()
    mux.Handle("/mcp", mcpHandler)
    mux.Handle("/admin/", adminHandler)

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### 14.8 推荐的写代码顺序

**不要一口气建所有文件**，按这个顺序写，每步都能验证：

| 步骤 | 内容 | 验证方式 |
|------|------|---------|
| 第 1 步 | `pkg/jsonrpc/message.go`（数据结构） | 跑单元测试 |
| 第 2 步 | `internal/config/config.go`（配置加载） | 读 config.yaml 验证 |
| 第 3 步 | `internal/interface/mcp/handler.go`（只实现 initialize 路由） | 用 curl 发 POST 验证 |
| 第 4 步 | `internal/application/initialize.go`（实现握手） | 连上 Cursor/Claude Desktop |
| 第 5 步 | `internal/infrastructure/db/`（数据库 + client repo） | CRUD 测试 |
| 第 6 步 | 跑通 initialize 全套流程 | **里程碑：Agent 能连上 Gateway** |
| 第 7 步 | 加 tools/list 和 tools/call | Agent 能看到工具、调用工具 |
| 第 8 步 | 加 admin API | 管理员能通过 REST 管理 |
