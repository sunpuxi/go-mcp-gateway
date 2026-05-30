import { useState } from 'react'
import { Card, Row, Col, Statistic, Table, Empty } from 'antd'
import {
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  ApiOutlined,
} from '@ant-design/icons'

interface Stats {
  projects: number
  tools: number
  clients: number
  sessions: number
}

const sessions = [
  { key: '1', client: '大数据项目组', status: '活跃', connected: '3 分钟前', toolCalled: 'get_user' },
  { key: '2', client: '支付项目组', status: '活跃', connected: '12 分钟前', toolCalled: 'create_post' },
]

const sessionColumns = [
  { title: '客户端', dataIndex: 'client', key: 'client' },
  { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <span style={{ color: '#52c41a' }}>● {s}</span> },
  { title: '连接时间', dataIndex: 'connected', key: 'connected' },
  { title: '最近调用', dataIndex: 'toolCalled', key: 'toolCalled' },
]

function Dashboard() {
  const [stats] = useState<Stats>({
    projects: 2,
    tools: 3,
    clients: 2,
    sessions: 0,
  })

  return (
    <div>
      <div className="page-header">
        <h2>仪表盘</h2>
      </div>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic title="项目数" value={stats.projects} prefix={<FolderOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="工具数" value={stats.tools} prefix={<ToolOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="客户端数" value={stats.clients} prefix={<KeyOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="活跃 Session" value={stats.sessions} prefix={<ApiOutlined />} />
          </Card>
        </Col>
      </Row>

      <Card title="网关运行状态">
        {sessions.length > 0 ? (
          <Table columns={sessionColumns} dataSource={sessions} pagination={false} size="small" />
        ) : (
          <Empty description="暂无活跃连接，连接数据库后将展示实时统计信息" />
        )}
      </Card>
    </div>
  )
}

export default Dashboard
