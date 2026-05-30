# P2 - 可观测性 & 运营

> 目标：全链路可追踪、操作可审计、问题可定位
> 预估周期：2 周

---

## 1. 分布式追踪 (OpenTelemetry)

### 现状
无任何 tracing，问题排查靠看日志猜测。

### 方案
引入 OpenTelemetry Go SDK：
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)
```

### Span 设计
每个 `tools/call` 一个 Root Span，子 Span 包括：
```
tools/call (root)
├── auth.authenticate        // 鉴权
├── tool.lookup              // 查工具定义
├── param.map                // 参数映射
├── http.request             // 下游 HTTP 调用
│   ├── dns.resolve
│   ├── tcp.connect
│   └── http.response
└── response.transform       // 响应转换
```

### 实现要点
- 通过 Context 传递 SpanContext
- 在 Chi 中间件中从请求头提取/创建 trace context
- 导出到 Jaeger / Grafana Tempo / 云厂商 APM
- `config.yaml` 增加 `tracing` 配置段
- trace_id 注入到结构化日志中（关联日志和追踪）

---

## 2. Kafka 审计日志

### 现状
`config.yaml` 已定义 Kafka brokers 和 topic，但代码中无任何 Kafka 相关代码。

### 方案
引入 `github.com/IBM/sarama`（或 `segmentio/kafka-go`）：

### 审计事件结构
```json
{
  "event": "tool.call",
  "timestamp": "2026-05-30T17:00:00Z",
  "trace_id": "abc123",
  "client_id": "client-001",
  "client_name": "MyAgent",
  "tool_id": "tool-001",
  "tool_name": "get_user",
  "project_id": "project-001",
  "arguments": {"user_id": "***"},   // 敏感参数脱敏
  "status": "success",
  "http_status": 200,
  "duration_ms": 123,
  "error": null
}
```

### 实现要点
- 新建 `internal/infrastructure/kafka/producer.go`
- 异步发送（不阻塞主流程）
- 失败兜底：写本地文件或丢弃（配置可选）
- 敏感参数脱敏规则可配置
- 下游消费者可做：计费、异常检测、用量分析

---

## 3. 调用统计分析

### 现状
Dashboard 只有数量统计（项目数/工具数/客户端数/Session数）。

### 方案

#### 3.1 统计数据库表
```sql
CREATE TABLE tool_call_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    stat_date DATE NOT NULL,
    stat_hour TINYINT,          -- NULL 表示按天聚合
    client_id VARCHAR(64) NOT NULL,
    tool_id VARCHAR(64) NOT NULL,
    call_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    error_count INT DEFAULT 0,
    total_duration_ms BIGINT DEFAULT 0,
    UNIQUE KEY uk_stat (stat_date, stat_hour, client_id, tool_id)
);
```

#### 3.2 Dashboard 展示
- 24h 调用量趋势图（折线图）
- 成功率 / 错误率（饼图或仪表盘）
- P50 / P95 / P99 延迟趋势
- Top 调用工具排行
- Top 活跃客户端排行
- 按项目/客户端/工具维度筛选

### 实现要点
- 异步写入统计表（不阻塞主流程）
- 后台定时聚合（或直接用 Prometheus + Grafana 替代）
- Admin API 新增统计查询接口

---

## 4. 告警通知

### 方案
对接常见通知渠道，触发条件可配置：

| 告警条件 | 说明 |
|----------|------|
| 下游错误率 > 阈值 | 连续 N 分钟 5xx 比例超 X% |
| 限流命中激增 | 某客户端频繁被限流 |
| Session 数异常 | 突增或突降 |
| 下游超时率过高 | 下游服务响应变慢 |

### 通知渠道
- Webhook（通用）
- 企业微信 / 飞书 / 钉钉机器人
- 邮件

---

## 5. Admin 操作审计

### 现状
Admin 后台任何操作无记录，无法追溯"谁改了工具配置导致故障"。

### 方案
```sql
CREATE TABLE admin_audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    operator VARCHAR(128) NOT NULL,   -- admin 或未来用户名
    action VARCHAR(64) NOT NULL,      -- create/update/delete/generate_key
    resource VARCHAR(64) NOT NULL,     -- project/tool/client/permission
    resource_id VARCHAR(64),
    detail JSON,                       -- 变更内容快照
    ip VARCHAR(45),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 实现要点
- Admin 中间件记录所有写操作
- Dashboard 增加"操作日志"页面
- 支持按时间/操作人/操作类型筛选

---

## 6. 配置热加载

### 现状
支持环境变量覆盖，但修改 `config.yaml` 需要重启。

### 方案
- 使用 `fsnotify` 监听配置文件变更
- 变更后重新加载到内存（无需重启）
- 重新加载范围：`server`、`rate_limit` 等非结构性配置
- 结构性变更（如 DB DSN）保留重启生效
