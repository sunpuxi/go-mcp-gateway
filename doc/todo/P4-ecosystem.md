# P4 - 生态 & 体验

> 目标：多协议支持、可扩展架构、开发者体验
> 预估周期：长远规划

---

## 1. gRPC 下游支持

### 现状
仅支持 HTTP 下游，无法直接调用 gRPC 服务。

### 方案
工具 `url_template` 扩展支持 `grpc://` 协议：
```json
{
  "http_method": "GRPC",
  "url_template": "grpc://user-service/UserService/GetUser",
  "grpc_reflection": true     // 是否通过 reflection 获取 proto 定义
}
```

### 实现要点
- 新建 `internal/infrastructure/grpc/grpc_client.go`
- 通过 gRPC Reflection 自动获取方法签名 → 生成 JSON Schema
- 支持 proto 文件导入（预编译场景）
- JSON ↔ Protobuf 自动序列化/反序列化

---

## 2. 响应缓存

### 场景
对幂等 GET 类工具调用，短时间内重复请求可直接返回缓存结果。

### 方案
工具配置增加：
```json
{
  "cache": {
    "enabled": true,
    "ttl_seconds": 60,
    "key_template": "{{.client_id}}:{{hash .arguments}}"
  }
}
```

### 实现要点
- 缓存后端：Redis（分布式）/ 内存（单实例）
- Cache Key 由 client_id + 参数 hash 组成
- 缓存命中跳过下游 HTTP 调用
- Admin 提供缓存清除接口（手动失效）

---

## 3. 插件系统 (Wasm)

### 目标
允许用户编写自定义中间件，在不修改网关代码的情况下扩展功能。

### 方案
- 使用 **Wasm (WebAssembly)** 作为沙箱执行环境
- 插件类型：
  - **前置钩子** (Pre-Request)：修改请求参数、增加 Header
  - **后置钩子** (Post-Response)：转换响应内容、提取数据

### 插件接口 (WASI)
```go
// 用户编写 → 编译为 .wasm
//export pre_request
func preRequest(inputPtr, inputLen int32) int64 {
    input := readMemory(inputPtr, inputLen)
    // 修改 input
    return writeResult(modifiedInput)
}
```

### 实现要点
- 引入 `github.com/tetratelabs/wazero`（纯 Go Wasm 运行时）
- 工具配置关联插件 ID
- Admin 提供插件上传/管理界面
- 沙箱约束：内存限制、执行超时、禁止网络/文件访问

---

## 4. 多租户隔离

### 场景
一个 Gateway 实例服务多个团队/业务线，需要数据和资源隔离。

### 方案
增加 `namespace` 概念：
```sql
ALTER TABLE projects ADD COLUMN namespace VARCHAR(64) DEFAULT 'default';
ALTER TABLE clients ADD COLUMN namespace VARCHAR(64) DEFAULT 'default';
```
- Admin API 按 namespace 隔离
- 每个 namespace 有独立的 admin API key
- Session 鉴权时校验 namespace 匹配

---

## 5. 请求/响应录制 & 回放

### 场景
- 测试工具配置变更后行为是否一致
- 故障复现

### 方案
- **录制**：工具调用时保存请求/响应到存储
- **回放**：在沙箱模式下调取历史录制数据，不真实调用下游
- Dashboard 提供录制管理和回放对比

---

## 6. A/B 测试 & 灰度发布

### 场景
工具升级（如切换下游版本）时逐步放量验证。

### 方案
```json
{
  "canary": {
    "enabled": true,
    "variants": [
      {"tool_id": "tool-v1", "weight": 90},
      {"tool_id": "tool-v2", "weight": 10}
    ]
  }
}
```
- 按 client_id hash 分配流量
- Admin 实时调整权重
- 对比不同变体的成功率/延迟

---

## 7. API 文档自动生成

### 方案
- 基于项目注册信息 + 工具定义，自动生成 OpenAPI 3.0 文档
- 暴露 `GET /admin/api/docs/openapi.json`
- Dashboard 内嵌 Swagger UI / Redoc 渲染

---

## 8. CLI 管理工具

### 方案
提供 `mcp-gateway` CLI 工具，支持通过命令行管理：

```bash
mcp-gateway tool list
mcp-gateway tool create --file tool-config.yaml
mcp-gateway tool export --all > tools-backup.yaml
mcp-gateway client gen-key --client-id client-001
mcp-gateway stats show --last 24h
```

### 实现要点
- 使用 `cobra` 库构建 CLI
- 复用 Admin API（CLI 本质是 Admin API 客户端）
- 支持配置文件（`~/.mcp-gateway.yaml`）存储 server 地址和 API Key

---

## 9. 压测 & 性能基准

### 方案
- 内置 benchmark 工具
- `mcp-gateway bench --concurrency 100 --duration 60s`
- 输出 QPS、P50/P95/P99 延迟、错误率
- 可视化压测报告

---

## 10. 开发者体验优化

| 项目 | 说明 |
|------|------|
| SDK 封装 | 提供 Go/Python/JS SDK，简化 Agent 对接 |
| Mock Server | 内置 Mock 模式，开发阶段不依赖下游 |
| 调试模式 | 详细展示每次调用的参数映射过程 |
| 示例项目 | 提供完整的示例项目（Agent + Gateway + 下游） |
| 单元测试覆盖 | 核心链路单元测试 > 80% |
| 集成测试 | docker-compose 一键启动完整环境 + 自动化测试 |
