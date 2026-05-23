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
