# MCP Gateway

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

一个 **MCP（Model Context Protocol）协议转换网关**，将 AI Agent 的 MCP 工具调用翻译为标准 HTTP 请求，转发到下游服务。

---

## 解决的问题

AI Agent 遵循 MCP 协议发起工具调用，但你的下游服务是 RESTful HTTP 接口。  
MCP Gateway 站在两者之间，完成**协议桥接**：

```
AI Agent                    MCP Gateway                    下游 HTTP 服务
  │                            │                               │
  │── tools/call ────────────▶ │  查数据库获取映射规则            │
  │   {name: "get_user",       │  参数映射 + URL 拼接            │
  │    args: {user_id: 123}}   │──────────────────────────────▶│  GET /users/123
  │                            │◀──────────────────────────────│  HTTP 200
  │◀── MCP 响应 ─────────────  │  响应格式转换                   │
```

---

## 核心特性

- **协议转换** — MCP JSON-RPC ↔ HTTP RESTful
- **参数映射引擎** — 支持 MCP 参数映射到 HTTP Path / Query / Body / Header
- **API Key 鉴权** — 连接时校验，Session 级权限，支持多租户
- **工具权限控制** — 细粒度控制每个 Client 能调用哪些 Tool
- **管理后台** — 可视化配置项目、工具、客户端、权限
- **实时生效** — 配置变更无需重启，数据库读取即时生效

---

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/sunpuxi/go-mcp-gateway.git
cd go-mcp-gateway
```

### 2. 初始化数据库

执行 `doc/sql/` 下的 SQL 脚本：

```bash
mysql -u root -p < doc/sql/init_table.sql
mysql -u root -p < doc/sql/seed_data.sql
```

### 3. 修改配置

编辑 `config/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8081

admin:
  api_key: "your_admin_key"        # 管理后台登录密钥

database:
  dsn: "user:password@tcp(127.0.0.1:3306)/go-mco-gateway?charset=utf8mb4&parseTime=True"
```

### 4. 启动网关

```bash
go build -o mcp-gateway.exe .
./mcp-gateway
```

### 5. 启动管理后台

```bash
cd web
npm install
npm run dev
```

管理后台默认运行在 `http://localhost:3000`。

---

## 项目结构

```
go-mcp-gateway/
├── main.go                          # 入口，依赖注入
├── config/
│   ├── config.go                    # 配置加载
│   └── config.yaml                  # 配置文件
├── pkg/
│   ├── jsonrpc/                     # JSON-RPC 2.0 协议定义
│   └── mcp/                         # MCP 消息类型
├── internal/
│   ├── application/
│   │   ├── command/                 # 用例输入/输出 DTO
│   │   ├── query/                   # 查询 DTO
│   │   └── service/                 # 应用层服务（MCPService）
│   ├── domain/
│   │   ├── entity/                  # 领域实体
│   │   ├── mapper/                  # 参数映射引擎
│   │   ├── repository/              # 仓储接口
│   │   └── service/                 # 领域服务（鉴权、会话、Schema、响应）
│   ├── infrastructure/
│   │   ├── db/                      # MySQL 实现
│   │   └── http/                    # HTTP 客户端
│   └── interface/
│       └── mcp/                     # MCP HTTP Handler
├── web/                             # 管理后台前端（React + Ant Design）
├── doc/
│   ├── sql/                         # 建表 & 测试数据
│   └── front-page-design/           # 后台界面设计文档
└── mcp-gateway.exe                  # 编译产物
```

分层架构：

```
interface（传输层）→ application（用例层）→ domain（领域层）→ infrastructure（基础设施层）
```

---

## Agent 配置

在 Agent 的 MCP 配置文件中添加：

```json
{
  "mcpServers": {
    "go-mcp-gateway": {
      "url": "http://your-host:8081/sse",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer <你的 API Key>"
      }
    }
  }
}
```

> API Key 在管理后台的「客户端管理」中生成。

---

## 错误码

| 错误码 | 含义 |
|--------|------|
| -32001 | 鉴权失败 |
| -32002 | 请求过于频繁（预留） |
| -32005 | 下游服务异常 |

标准 JSON-RPC 错误码：

| 错误码 | 含义 |
|--------|------|
| -32700 | JSON 解析错误 |
| -32600 | 无效请求 |
| -32601 | 方法不存在 |
| -32602 | 无效参数 |
| -32603 | 内部错误 |

---

## 数据库表

| 表名 | 说明 |
|------|------|
| `projects` | 下游 HTTP 服务注册 |
| `tools` | 工具定义 & 参数映射规则 |
| `clients` | API 客户端（项目组） |
| `client_tool_permissions` | 客户端工具授权 |

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端语言 | Go 1.25 |
| Web 框架 | chi/v5 |
| 数据库 | MySQL + sqlx |
| 前端框架 | React 18 + TypeScript |
| UI 组件库 | Ant Design 5 |
| 构建工具 | Vite 5 |

---

## License

MIT
