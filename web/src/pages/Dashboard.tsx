import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Table, Empty, Skeleton } from 'antd'
import {
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Stats, getStats } from '../api'

const sessionColumns = [
  { title: '客户端', dataIndex: 'client', key: 'client' },
  { title: '状态', dataIndex: 'status', key: 'status', render: (s: string) => <span style={{ color: '#52c41a' }}>● {s}</span> },
  { title: '连接时间', dataIndex: 'connected', key: 'connected' },
  { title: '最近调用', dataIndex: 'toolCalled', key: 'toolCalled' },
]

function Dashboard() {
  const [stats, setStats] = useState<Stats>({ projects: 0, tools: 0, clients: 0, sessions: 0 })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getStats()
        setStats(data)
      } catch (e: unknown) {
        toast.error('加载统计数据失败: ' + (e instanceof Error ? e.message : String(e)))
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return (
      <div>
        <div className="page-header"><h2>仪表盘</h2></div>
        <Row gutter={16} style={{ marginBottom: 24 }}>
          {[1, 2, 3, 4].map(i => (
            <Col span={6} key={i}>
              <Card><Skeleton active paragraph={{ rows: 1 }} /></Card>
            </Col>
          ))}
        </Row>
      </div>
    )
  }

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
        {stats.sessions > 0 ? (
          <Table columns={sessionColumns} dataSource={[]} pagination={false} size="small" />
        ) : (
          <Empty description="暂无活跃连接，MCP 客户端连接后将在此展示实时 Session 信息" />
        )}
      </Card>
    </div>
  )
}

export default Dashboard
