# P3 - 功能增强

> 目标：协议完善、管理体验提升
> 预估周期：按需

---

## 1. MCP Resources 支持

### 现状
`ServerCapabilities.Resources` 是空 struct，Agent 无法获取资源。

### 方案
MCP 协议完整支持 Resources：
- `resources/list` — 返回所有可用资源
- `resources/read` — 读取指定资源内容
- `resources/templates/list` — 资源模板（支持 URI 模板参数）

### 使用场景
- 将下游 HTTP 服务的 OpenAPI/Swagger 文档暴露为资源
- 将工具使用说明文档暴露为资源
- 将项目配置信息暴露为资源

### 实现要点
- `tools` 表扩展 `resource` 类型（工具 OR 资源）
- 资源也走参数映射 + HTTP 转发机制
- JSON-RPC 方法路由增加 `resources/*` 处理

---

## 2. 工具变更通知

### 现状
`ServerCapabilities.Tools.ListChanged = false`，Agent 不知道工具何时变更。

### 方案
- 管理员在后台修改工具后，Gateway 推送 SSE 事件：
  ```
  event: message
  data: {"jsonrpc":"2.0","method":"notifications/tools/list_changed"}
  ```
- Agent 收到后自动重新调用 `tools/list` 更新本地工具列表
- `ServerCapabilities.Tools.ListChanged = true`

### 实现要点
- Admin Service 在 Create/Update/Delete Tool 后通知 SessionManager
- SessionManager 遍历所有活跃 Session，推送变更事件
- 支持按项目/客户端灰度推送

---

## 3. WebSocket 传输支持

### 现状
仅 SSE（GET /sse + POST /messages），半双工。

### 方案
MCP 2025 规范新增 WebSocket 传输：
```
GET /ws → Upgrade → WebSocket 连接
双方通过 WebSocket 帧交换 JSON-RPC 消息
```

### 优势
- 全双工，无需额外的 POST /messages 端点
- 减少连接数（SSE 每次 tools/call 需要新 HTTP 请求）
- 更好的移动端支持

### 实现要点
- 使用 `gorilla/websocket` 或 `nhooyr.io/websocket`
- 鉴权通过连接时的 `Authorization` 头
- 复用现有 JSON-RPC 消息路由逻辑
- SSE 和 WebSocket 并存（通过配置选择）

---

## 4. 管理后台多用户 & RBAC

### 现状
只支持单一 `admin_api_key`，无多用户体系。

### 方案

#### 4.1 用户表
```sql
CREATE TABLE admin_users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'viewer',  -- superadmin/admin/viewer
    status VARCHAR(16) DEFAULT 'active',
    last_login_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 4.2 RBAC 角色
| 角色 | 权限 |
|------|------|
| `superadmin` | 全部（包括用户管理） |
| `admin` | 项目/工具/客户端 CRUD |
| `viewer` | 只读查看 |

### 实现要点
- 登录接口：`POST /admin/api/login` → JWT Token
- Admin 鉴权中间件支持 Bearer Token（JWT）和 API Key 两种方式
- Dashboard 增加用户管理页面（仅 superadmin 可见）
- JWT 过期/刷新机制

---

## 5. 工具配置导入导出 & 版本管理

### 方案

#### 5.1 GitOps 导入导出
```bash
# 导出所有工具
GET /admin/api/tools/export → JSON/YAML 文件

# 导入工具
POST /admin/api/tools/import ← JSON/YAML 文件
```
- 导出格式：完整的工具定义 JSON（含参数规则）
- 导入支持：创建新工具 / 更新已有工具（按 tool_id 幂等）
- 配合 Git 实现 Infrastructure as Code

#### 5.2 配置变更历史
```sql
CREATE TABLE tool_revisions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    tool_id VARCHAR(64) NOT NULL,
    version INT NOT NULL,
    config JSON NOT NULL,        -- 完整工具配置快照
    changed_by VARCHAR(128),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tool_version (tool_id, version)
);
```
- 每次更新工具自动创建新版本
- Dashboard 支持查看历史版本 & 回滚到指定版本

---

## 6. 响应转换模板

### 现状
下游 HTTP 响应直接透传给 Agent，无法定制。

### 方案
工具配置增加 `response_template` 字段：
```json
{
  "response_template": {
    "type": "gotemplate",    // gotemplate / jsonpath / jq
    "template": "用户 {{.name}} 的邮箱是 {{.email}}"
  }
}
```

### 使用场景
- 下游返回大量字段，只需提取部分
- 下游字段名与 MCP 约定不一致，需要重命名
- 聚合多个字段计算新值

### 实现要点
- 支持 Go template 语法
- 支持 JsonPath 提取
- 模板渲染异常时优雅降级（透传原始响应）

---

## 7. 工具依赖 & 编排

### 方案
工具间可以定义依赖关系：
```json
{
  "depends_on": [
    {
      "tool_id": "user.login",
      "output_mapping": {
        "token": "$.data.access_token"
      }
    }
  ]
}
```
- 调用工具前自动调用依赖工具
- 依赖工具的输出自动注入到当前工具的参数/Header 中
- 支持依赖链（A → B → C）

---

## 8. 下游服务发现

### 方案
支持从服务注册中心动态获取下游服务列表：
- Consul
- Nacos
- Kubernetes Service (通过 API Server)

### 实现要点
- 新建 `internal/infrastructure/discovery/` 目录
- 定义 `ServiceDiscoverer` 接口
- 项目 `base_url` 支持 `discovery://service-name` 格式
- 定期刷新服务实例列表，自动摘除不健康实例
