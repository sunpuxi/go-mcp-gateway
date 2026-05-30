# MCP Gateway 管理后台 — 产品需求文档

## 概述

MCP Gateway 是一个**协议转换网关**，它将 MCP 协议（Model Context Protocol）转换为对下游 HTTP 服务的调用。管理后台是给运维/管理人员使用，用于管理网关的所有配置。

## 角色与权限

| 角色 | 说明 |
|------|------|
| **管理员** | 唯一角色，使用配置文件中预置的 `admin_api_key` 登录后台 |

> 当前版本暂不涉及多管理员、RBAC 等复杂设计。

## 页面列表

共 **5 个页面**，左侧导航栏：

```
MCP Gateway 管理后台
├── 📊 仪表盘
├── 📁 项目管理
├── 🛠 工具管理
├── 🔑 客户端管理
└── 🔐 权限管理
```

---

## 页面 1：仪表盘

**用途**：概览网关运行状态。

**内容**：

| 卡片 | 数据来源 |
|------|---------|
| 项目数 | `SELECT COUNT(*) FROM projects` |
| 工具数 | `SELECT COUNT(*) FROM tools` |
| 客户端数 | `SELECT COUNT(*) FROM clients` |
| 活跃 Session 数 | 从 SessionManager 内存中读取 |
| 最近 10 个 Session 日志 | 运行日志（可选） |

布局：4 个统计卡片横向排列 + 下方活动日志列表。

---

## 页面 2：项目管理

**管理对象**：`projects` 表，即接入了哪些下游 HTTP 服务。

### 列表页

| 字段 | 说明 |
|------|------|
| Project ID | 唯一标识，如 `proj_user` |
| 名称 | 如"用户服务" |
| 基础 URL | `https://jsonplaceholder.typicode.com` |
| 状态 | 启用/禁用（开关） |
| 操作 | 编辑 / 删除 |

### 新建/编辑表单

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Project ID | 文本 | 是 | 唯一标识，创建后不可修改 |
| 名称 | 文本 | 是 | 显示名称 |
| 基础 URL | URL | 是 | 下游服务的基础地址 |
| 描述 | 多行文本 | 否 | 备注信息 |
| 状态 | 开关 | 是 | 默认启用 |

---

## 页面 3：工具管理

**管理对象**：`tools` 表，即每个 HTTP 接口的协议转换规则。**这是最核心的配置页面**。

### 列表页

| 字段 | 说明 |
|------|------|
| 工具名称 | `get_user` |
| 标题 | "获取用户信息" |
| 所属项目 | 显示 project 名称 |
| HTTP 方法 | GET / POST / PUT / DELETE |
| URL 模板 | `/users/{user_id}` |
| 参数数量 | 该工具有几个参数 |
| 状态 | 启用/禁用 |
| 操作 | 编辑 / 删除 |

### 新建/编辑表单 — 分两个区域

#### 区域 1：基础信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 工具名称 | 文本 | 是 | MCP 调用时用的 name，全局唯一 |
| 标题 | 文本 | 是 | 显示标题 |
| 描述 | 多行文本 | 否 | 对工具功能的说明 |
| 所属项目 | 下拉框 | 是 | 从 projects 表选择 |
| HTTP 方法 | 下拉框 | 是 | GET / POST / PUT / DELETE / PATCH |
| URL 模板 | 文本 | 是 | 支持 `{param}` 占位符，如 `/users/{user_id}` |
| 超时(ms) | 数字 | 否 | 默认 5000 |
| 状态 | 开关 | 是 | 默认启用 |

#### 区域 2：参数映射规则（动态列表）

这是最关键的配置。每条规则包含：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 参数名 | 文本 | 是 | MCP arguments 中的 key |
| 类型 | 下拉框 | 否 | string / number / boolean / object |
| 位置 | 下拉框 | 是 | **path**（URL 路径）/ **query**（查询参数）/ **body**（请求体 JSON）/ **header**（HTTP 头） |
| 必填 | 开关 | 否 | 是否必须 |
| 默认值 | 文本 | 否 | 非必填参数设此值 |
| 描述 | 文本 | 否 | 参数说明，会作为 MCP Schema 的 description |

操作：**新增一行** / **删除一行**，支持拖拽排序。

**示例配置效果**：

```
MCP 调用参数: { "user_id": "123", "fields": "name,email" }

参数映射规则：
  user_id → location: path   → 替换 URL 模板中的 {user_id}
  fields  → location: query  → 拼接到 URL 查询参数

最终 HTTP 请求: GET /users/123?fields=name,email
```

---

## 页面 4：客户端管理

**管理对象**：`clients` 表，即有哪些团队/应用在使用网关。

### 列表页

| 字段 | 说明 |
|------|------|
| Client ID | `cli_bigdata` |
| 名称 | "大数据项目组" |
| Key 前缀 | `sk-bigdata`（用于识别是哪个客户端） |
| 已授权工具数 | 关联的权限数量 |
| 状态 | 启用/禁用 |
| 创建时间 | |
| 操作 | 编辑 / 生成新 Key / 删除 |

### 新建/编辑表单

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Client ID | 文本 | 是 | 唯一标识，如 `cli_bigdata` |
| 名称 | 文本 | 是 | 显示名称 |
| 描述 | 多行文本 | 否 | 备注 |
| 状态 | 开关 | 是 | 默认启用 |

### 生成 API Key 的交互流程（关键）

```
1. 管理员点击"生成 API Key"
2. 后端生成：sk-bigdata-xxxxxxxxxxxxxxxxxxxx（随机 20 位）
3. 界面弹窗展示完整 Key（仅此一次，关闭后不再显示）
   ┌────────────────────────────────────────────┐
   │  API Key 已生成                             │
   │                                            │
   │  sk-bigdata-a1b2c3d4e5f6g7h8i9j0            │
   │                                            │
   │  ⚠ 请立即复制保存，关闭后将无法再次查看       │
   │                                            │
   │  [复制] [关闭]                               │
   └────────────────────────────────────────────┘
4. 后端存 SHA256(完整 key) 到 api_key_hash 字段
```

**禁用客户端**：设置 `clients.status = 0`，该客户端的所有 Session 将失效。

---

## 页面 5：权限管理

**用途**：将工具授权给客户端。

**交互设计**：左侧客户端列表 + 右侧勾选工具。

```
┌──────────────────────────────────────────────────┐
│  选择客户端:  [大数据项目组 ▼]                      │
│                                                  │
│  ┌── 用户服务 ──────────────────────────────────┐ │
│  │  ☑ get_user     获取用户信息                   │ │
│  │  ☑ get_user_posts  获取用户帖子                 │ │
│  ├── 帖子服务 ──────────────────────────────────┐ │
│  │  ☐ create_post  创建帖子                       │ │
│  └──────────────────────────────────────────────┘ │
│                                                  │
│  [保存权限]                                       │
└──────────────────────────────────────────────────┘
```

**说明**：
- 顶部下拉选择客户端
- 工具按 **project** 分组展示（可折叠）
- 每个 tool 前面是 checkbox
- 点击保存后，先 DELETE 该 client 的所有旧权限，再 INSERT 新的
- 涉及 `client_tool_permissions` 表

---

## 后端 API 接口清单

管理后台需要调用以下后端接口，路由前缀统一为 `/admin/api`：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/api/projects` | 项目列表 |
| POST | `/admin/api/projects` | 新建项目 |
| PUT | `/admin/api/projects/:id` | 编辑项目 |
| DELETE | `/admin/api/projects/:id` | 删除项目 |
| GET | `/admin/api/tools` | 工具列表 |
| POST | `/admin/api/tools` | 新建工具 |
| PUT | `/admin/api/tools/:id` | 编辑工具 |
| DELETE | `/admin/api/tools/:id` | 删除工具 |
| GET | `/admin/api/clients` | 客户端列表 |
| POST | `/admin/api/clients` | 新建客户端 |
| POST | `/admin/api/clients/:id/api-key` | 生成新 API Key（返回明文 Key） |
| PUT | `/admin/api/clients/:id` | 编辑客户端 |
| DELETE | `/admin/api/clients/:id` | 删除客户端 |
| GET | `/admin/api/clients/:id/permissions` | 查某个 client 的权限 |
| PUT | `/admin/api/clients/:id/permissions` | 更新权限（传 tool_id 数组） |
| GET | `/admin/api/stats` | 仪表盘统计数据 |

**所有 `/admin/api/*` 接口都需要校验 `admin_api_key`**，方式同 MCP 鉴权：

```
Authorization: Bearer <admin_api_key>
```

`admin_api_key` 配置在当前 `config/config.yaml` 中：

```yaml
admin:
  api_key: "admin_change_me"
```

---

## 数据库表结构（供前端参考字段类型）

### projects

| 字段 | 类型 | 说明 |
|------|------|------|
| project_id | VARCHAR(64) PK | 唯一标识 |
| name | VARCHAR(128) | 项目名称 |
| base_url | VARCHAR(512) | 基础 URL |
| description | TEXT | 描述 |
| status | TINYINT | 1=启用, 0=禁用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### tools

| 字段 | 类型 | 说明 |
|------|------|------|
| tool_id | BIGINT PK AUTO_INCREMENT | 工具 ID |
| project_id | VARCHAR(64) FK | 所属项目 |
| name | VARCHAR(128) UNIQUE | 工具名称 |
| title | VARCHAR(256) | 标题 |
| description | TEXT | 描述 |
| http_method | VARCHAR(10) | HTTP 方法 |
| url_template | VARCHAR(512) | URL 模板 |
| timeout_ms | INT DEFAULT 5000 | 超时 |
| params | JSON | 参数映射规则数组 |
| status | TINYINT | 1=启用, 0=禁用 |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### clients

| 字段 | 类型 | 说明 |
|------|------|------|
| client_id | VARCHAR(64) PK | 唯一标识 |
| name | VARCHAR(128) | 名称 |
| api_key_hash | VARCHAR(64) | SHA256 加密后的 Key |
| api_key_prefix | VARCHAR(32) | Key 前缀（展示用） |
| description | TEXT | 备注 |
| status | TINYINT | 1=启用, 0=禁用 |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### client_tool_permissions

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| client_id | VARCHAR(64) | 关联 clients |
| tool_id | BIGINT | 关联 tools |
| created_at | DATETIME | |

---

## 技术约束

1. **前端框架不限**，建议 React / Vue 均可
2. 所有 API 采用 JSON 格式
3. 分页：列表接口支持 `?page=1&size=20` 参数
4. 工具的参数映射规则（`params` 字段）在 DB 中存储为 JSON 格式
5. API Key 仅在生成时展示一次明文，后端只存 SHA256 hash
