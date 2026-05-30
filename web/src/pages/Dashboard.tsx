import { useState, useEffect, useRef, useCallback } from 'react'
import { Card, Row, Col, Statistic, Table, Tag, Empty, Skeleton } from 'antd'
import {
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Stats, SessionInfo, getStats } from '../api'

const sessionColumns = [
  { title: 'Session ID', dataIndex: 'id', key: 'id', width: 180, ellipsis: true,
    render: (v: string) => <code style={{ fontSize: 12 }}>{v.substring(0, 8)}...</code> },
  { title: '客户端', dataIndex: 'client_id', key: 'client_id', width: 140 },
  { title: '协议版本', dataIndex: 'protocol_version', key: 'protocol_version', width: 100,
    render: (v: string) => v || '-' },
  { title: '状态', dataIndex: 'initialized', key: 'initialized', width: 100, align: 'center' as const,
    render: (v: boolean) => v
      ? <Tag color="success">已初始化</Tag>
      : <Tag color="processing">已连接</Tag>
  },
  { title: '连接时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
]

function Dashboard() {
  const [stats, setStats] = useState<Stats>({ projects: 0, tools: 0, clients: 0, sessions: 0, session_list: [] })
  const [loading, setLoading] = useState(true)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const load = useCallback(async () => {
    try {
      const data = await getStats()
      setStats(data)
    } catch (e: unknown) {
      toast.error('加载统计数据失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    timerRef.current = setInterval(load, 5000)
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [load])

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
          <Table columns={sessionColumns} dataSource={stats.session_list} rowKey="id" pagination={false} size="small" />
        ) : (
          <Empty description="暂无活跃连接，MCP 客户端连接后将在此展示实时 Session 信息" />
        )}
      </Card>
    </div>
  )
}

export default Dashboard
