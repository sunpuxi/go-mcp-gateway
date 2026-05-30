import { useState } from 'react'
import { Card, Typography, Divider, Steps, Table, Alert, Tag, Space, Collapse } from 'antd'
import {
  BookOutlined,
  ApiOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  LinkOutlined,
  RocketOutlined,
  CodeOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons'

const { Title, Paragraph, Text } = Typography

const steps = [
  {
    title: '接入下游服务',
    description: '在"项目管理"中添加你要对接的 HTTP 服务',
    content: (
      <div>
        <Paragraph>
          在左侧菜单进入<Text strong>「项目管理」</Text>，点击「新建项目」，填写：
        </Paragraph>
        <ul>
          <li><Text strong>Project ID</Text>：项目的唯一标识，如 <code>proj_user</code>（创建后不可修改）</li>
          <li><Text strong>项目名称</Text>：便于识别的中文名称，如「用户服务」</li>
          <li><Text strong>基础 URL</Text>：下游 HTTP 服务的基础地址，如 <code>https://api.example.com</code></li>
          <li><Text strong>描述</Text>：可选，对项目的备注说明</li>
        </ul>
      </div>
    ),
  },
  {
    title: '配置工具转换',
    description: '在"工具管理"中配置每个 HTTP 接口的协议转换规则',
    content: (
      <div>
        <Paragraph>
          进入<Text strong>「工具管理」</Text>，点击「新建工具」，这是最核心的配置步骤：
        </Paragraph>
        <Paragraph strong>① 基础信息</Paragraph>
        <ul>
          <li><Text strong>工具名称</Text>：MCP 调用时使用的标识，如 <code>get_user</code></li>
          <li><Text strong>标题</Text>：工具的显示名称，如「获取用户信息」</li>
          <li><Text strong>所属项目</Text>：选择该接口属于前面创建的哪个项目</li>
          <li><Text strong>HTTP 方法</Text>：GET / POST / PUT / DELETE / PATCH</li>
          <li><Text strong>URL 模板</Text>：使用 <code>{'{参数名}'}</code> 作为占位符，如 <code>/users/{'{user_id}'}</code></li>
          <li><Text strong>超时(ms)</Text>：默认 5000，按需调整</li>
        </ul>
        <Paragraph strong>② 参数映射规则（最重要）</Paragraph>
        <Paragraph>
          每条规则定义了 MCP 调用参数如何转换为 HTTP 请求的对应部分。需要逐一添加：
        </Paragraph>
        <Table
          dataSource={[
            { key: '1', param: '参数名', desc: 'MCP arguments 中的 key，如 user_id' },
            { key: '2', param: '类型', desc: 'string / number / boolean / object' },
            { key: '3', param: '位置', desc: 'path（URL 路径替换）/ query（查询参数）/ body（请求体 JSON）/ header（HTTP 请求头）' },
            { key: '4', param: '必填', desc: '参数是否必须提供' },
            { key: '5', param: '默认值', desc: '非必填参数未传时的默认值' },
            { key: '6', param: '描述', desc: '参数说明，会出现在 MCP 工具的 Schema 中' },
          ]}
          columns={[
            { title: '字段', dataIndex: 'param', key: 'param', width: 80 },
            { title: '说明', dataIndex: 'desc', key: 'desc' },
          ]}
          pagination={false}
          size="small"
          bordered
        />
        <Alert
          type="info"
          showIcon
          style={{ marginTop: 16 }}
          message="什么是参数映射？"
          description={
            <div>
              <Paragraph style={{ marginTop: 8, marginBottom: 8 }}>
                MCP 协议使用 <code>{'{ "user_id": "123", "fields": "name,email" }'}</code> 这样的 JSON 参数。网关需要知道该把这些参数放到 HTTP 请求的哪里：
              </Paragraph>
              <ul style={{ marginBottom: 0 }}>
                <li><code>user_id → path</code>：替换 URL 模板中的 <code>{'{user_id}'}</code></li>
                <li><code>fields → query</code>：附加到 URL 上变成 <code>?fields=name,email</code></li>
                <li><code>data → body</code>：序列化为 JSON 请求体</li>
                <li><code>X-Token → header</code>：设置为 HTTP 请求头</li>
              </ul>
            </div>
          }
        />
      </div>
    ),
  },
  {
    title: '创建客户端 & 分配 Key',
    description: '在"客户端管理"中为每个使用方创建身份并生成 API Key',
    content: (
      <div>
        <Paragraph>
          进入<Text strong>「客户端管理」</Text>，点击「新建客户端」：
        </Paragraph>
        <ul>
          <li><Text strong>Client ID</Text>：客户端的唯一标识，如 <code>cli_bigdata</code></li>
          <li><Text strong>名称</Text>：便于识别的名称，如「大数据项目组」</li>
        </ul>
        <Paragraph>
          创建后，点击<Text strong>「生成 Key」</Text>按钮。系统会生成类似 <code style={{ background: '#e6f4ff', padding: '2px 6px', borderRadius: 4 }}>sk-bigdata-a1b2c3d4e5f6g7h8i9j0</code> 的 API Key。
        </Paragraph>
        <Alert
          type="warning"
          showIcon
          message="注意：API Key 仅在生成时展示一次，请立即复制保存。关闭弹窗后将无法再次查看。"
        />
        <Paragraph style={{ marginTop: 16 }}>
          每个项目组分配一个客户端即可，同一个 API Key 可以被多个 Agent 实例同时使用。
        </Paragraph>
      </div>
    ),
  },
  {
    title: '配置工具权限',
    description: '在"权限管理"中指定每个客户端能访问哪些工具',
    content: (
      <div>
        <Paragraph>
          进入<Text strong>「权限管理」</Text>：
        </Paragraph>
        <ol>
          <li>左侧选择要配置的客户端</li>
          <li>右侧勾选该客户端可以使用的工具（按项目分组展示）</li>
          <li>点击<Text strong>「保存权限」</Text></li>
        </ol>
        <Paragraph>
          客户端只能看到和调用被授权的工具。未授权的工具在 <code>tools/list</code> 时不会返回。
        </Paragraph>
        <Alert
          type="success"
          showIcon
          style={{ marginTop: 12 }}
          message="权限即时生效"
          description="保存权限后，该客户端已有的 Session 无需重建，新权限立即生效。"
        />
      </div>
    ),
  },
  {
    title: 'Agent 接入调用',
    description: '将 API Key 配置到 Agent 中，通过 MCP 协议调用工具',
    content: (
      <div>
        <Paragraph>
          在 Agent 的 MCP 配置文件中添加：
        </Paragraph>
        <pre style={{ background: '#1a1a2e', color: '#e0e0e0', padding: 20, borderRadius: 8, fontSize: 14, lineHeight: 1.8, overflow: 'auto' }}>
{`{
  "mcpServers": {
    "go-mcp-gateway": {
      "url": "http://your-host:8081/sse",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer sk-bigdata-a1b2c3d4e5f6g7h8i9j0"
      }
    }
  }
}`}
        </pre>
        <ul style={{ marginTop: 16 }}>
          <li><Text strong>url</Text>：网关的 SSE 端点地址</li>
          <li><Text strong>transport</Text>：固定为 <code>sse</code></li>
          <li><Text strong>headers.Authorization</Text>：格式为 <code>Bearer {'<'}你的 API Key{'}'}</code></li>
        </ul>
        <Paragraph style={{ marginTop: 12 }}>
          配置完成后重启 Agent，Agent 会自动：
        </Paragraph>
        <ol>
          <li>建立 SSE 连接并携带 API Key 进行鉴权</li>
          <li>调用 <code>tools/list</code> 获取可用的工具列表</li>
          <li>通过 <code>tools/call</code> 调用具体工具</li>
        </ol>
      </div>
    ),
  },
]

const faqs = [
  {
    key: '1', label: '同一个 API Key 可以多人共用吗？',
    children: <Paragraph>可以。每个客户端对应一个项目组，组内的所有 Agent 实例共享同一个 API Key。</Paragraph>,
  },
  {
    key: '2', label: 'API Key 泄露了怎么办？',
    children: (
      <Paragraph>
        在「客户端管理」中点击「生成 Key」会生成新的 API Key，旧的 Key 会立即失效。建议拿到新 Key 后先更新所有 Agent 配置，再生成新 Key 使旧 Key 失效。
      </Paragraph>
    ),
  },
  {
    key: '3', label: '如何禁用某个项目组的访问？',
    children: (
      <Paragraph>
        在「客户端管理」中找到对应客户端，点击状态列的开关将其设为「禁用」。禁用后该客户端的所有 Agent 立即无法连接。
      </Paragraph>
    ),
  },
  {
    key: '4', label: '工具参数映射配置错误会怎样？',
    children: (
      <Paragraph>
        如果 MCP 调用时缺少必填参数，网关会返回参数校验错误。如果 HTTP 下游返回异常，网关会以 <code>CodeDownstreamErr (-32005)</code> 错误码返回给 Agent。
      </Paragraph>
    ),
  },
  {
    key: '5', label: '新增或修改工具后需要重启网关吗？',
    children: (
      <Paragraph>不需要。网关在每次 <code>tools/list</code> 和 <code>tools/call</code> 时实时从数据库读取最新配置，修改后立即生效。</Paragraph>
    ),
  },
  {
    key: '6', label: '网关支持哪些下游 HTTP 接口？',
    children: (
      <Paragraph>
        支持任何 RESTful HTTP 接口。支持 GET、POST、PUT、DELETE、PATCH 方法，参数可以映射到 URL Path、Query String、Request Body 和 Header。
      </Paragraph>
    ),
  },
]

function Help() {
  const [current, setCurrent] = useState(0)

  return (
    <div>
      <div className="page-header">
        <Space>
          <BookOutlined style={{ fontSize: 22, color: '#1677ff' }} />
          <h2 style={{ margin: 0 }}>使用帮助</h2>
        </Space>
      </div>

      {/* 快速开始 */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={3} style={{ marginTop: 0 }}>
          <RocketOutlined /> 快速开始
        </Title>
        <Paragraph>
          MCP Gateway 是一个<Text strong>协议转换网关</Text>，它将 AI Agent 的 MCP 协议调用翻译为对下游 HTTP 服务的调用。
          按照以下 5 步即可完成从零到可用的完整配置：
        </Paragraph>
        <Steps
          current={current}
          onChange={setCurrent}
          direction="vertical"
          size="small"
          items={steps.map(item => ({
            title: item.title,
            description: item.description,
          }))}
        />
        <Divider style={{ margin: '16px 0' }} />
        <Card type="inner" style={{ background: '#fafafa' }}>
          {steps[current].content}
        </Card>
      </Card>

      {/* 架构概要 */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={3} style={{ marginTop: 0 }}>
          <ApiOutlined /> 调用链路
        </Title>
        <Alert
          type="info"
          showIcon={false}
          style={{ marginBottom: 16 }}
          message={
            <pre style={{ margin: 0, fontSize: 13, lineHeight: 2, whiteSpace: 'pre-wrap' }}>
{`AI Agent                     MCP Gateway                     下游 HTTP 服务
  │                             │                                │
  │── GET /sse ───────────────▶ │                                │
  │   (带上 API Key)             │ ① 鉴权 & 创建 Session          │
  │                             │                                │
  │◀── tools/list ───────────── │ ② 返回有权限的工具列表           │
  │                             │                                │
  │── tools/call ─────────────▶ │ ③ 权限校验 → 参数映射           │
  │   {name:"get_user",         │ ④ HTTP GET /users/123 ───────▶│
  │    args:{user_id:"123"}}    │ ⑤ ◀────── HTTP 200 ──────────│ │
  │◀── MCP 响应 ──────────────  │ ⑥ 响应转换 → SSE 推送          │`}
            </pre>
          }
        />
        <Paragraph>
          <Text strong>核心原理</Text>：网关从数据库读取每个工具的参数映射规则，将 MCP 的 JSON 参数按规则翻译为 HTTP 的 path/query/body/header，然后转发给下游服务。
        </Paragraph>
      </Card>

      {/* 常见问题 */}
      <Card>
        <Title level={3} style={{ marginTop: 0 }}>
          <QuestionCircleOutlined /> 常见问题
        </Title>
        <Collapse items={faqs} />
      </Card>
    </div>
  )
}

export default Help
