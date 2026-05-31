import { useState, useEffect, useRef, useCallback } from 'react'
import {
  Card, Row, Col, Statistic, Table, Tag, Empty, Skeleton, Tooltip,
  Badge, Timeline, Typography, Space, Button,
} from 'antd'
import {
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ReloadOutlined,
  ClockCircleOutlined,
  CheckOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  StopOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Stats, getStats } from '../api'
import { useCountUp } from '../hooks/useCountUp'
import { getOperationLogs, type OperationRecord } from '../utils/operationLog'

const { Text } = Typography

// ── 工具函数 ──

/** 较上次变化的数据 */
interface DeltaRecord {
  projects: number
  tools: number
  clients: number
  sessions: number
}

function getPreviousStats(): DeltaRecord | null {
  try {
    const raw = localStorage.getItem('mcp_gateway_prev_stats')
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function saveCurrentStats(s: Stats) {
  try {
    localStorage.setItem('mcp_gateway_prev_stats', JSON.stringify({
      projects: s.projects,
      tools: s.tools,
      clients: s.clients,
      sessions: s.sessions,
    }))
  } catch { /* noop */ }
}

function deltaLabel(current: number, previous: number | undefined): string | null {
  if (previous === undefined || previous === null) return null
  const diff = current - previous
  if (diff > 0) return `+${diff}`
  if (diff < 0) return `${diff}`
  return null
}

function deltaColor(current: number, previous: number | undefined): string | undefined {
  if (previous === undefined || previous === null) return undefined
  if (current > previous) return '#52c41a'
  if (current < previous) return '#ff4d4f'
  return undefined
}

// ── 操作日志图标映射 ──
const actionIcons: Record<string, React.ReactNode> = {
  '新增': <PlusOutlined style={{ color: '#52c41a' }} />,
  '编辑': <EditOutlined style={{ color: '#1677ff' }} />,
  '删除': <DeleteOutlined style={{ color: '#ff4d4f' }} />,
  '启用': <PlayCircleOutlined style={{ color: '#52c41a' }} />,
  '停用': <StopOutlined style={{ color: '#fa8c16' }} />,
  '生成密钥': <KeyOutlined style={{ color: '#722ed1' }} />,
  '权限变更': <SafetyOutlined style={{ color: '#1677ff' }} />,
}

import { SafetyOutlined } from '@ant-design/icons'

// ── Session 表格列 ──

const sessionColumns = [
  {
    title: 'Session ID', dataIndex: 'id', key: 'id', width: 220,
    render: (v: string) => (
      <Tooltip title={v}>
        <code style={{ fontSize: 12, cursor: 'pointer' }}>{v.substring(0, 16)}...</code>
      </Tooltip>
    ),
  },
  { title: '客户端', dataIndex: 'client_id', key: 'client_id', width: 160 },
  {
    title: '协议版本', dataIndex: 'protocol_version', key: 'protocol_version', width: 100,
    render: (v: string) => v || '-',
  },
  {
    title: '状态', dataIndex: 'initialized', key: 'initialized', width: 110, align: 'center' as const,
    render: (v: boolean) =>
      v
        ? <Tag color="success" icon={<CheckCircleOutlined />}>已初始化</Tag>
        : <Tag color="processing" icon={<SyncOutlined spin />}>已连接</Tag>,
  },
  { title: '连接时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
]

// ── 统计卡片图标样式 ──

const iconStyle = (color: string, bg: string) => ({
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  width: 44,
  height: 44,
  fontSize: 22,
  color,
  background: bg,
  borderRadius: 10,
})

// ── 主组件 ──

function Dashboard() {
  const [stats, setStats] = useState<Stats>({
    projects: 0, tools: 0, clients: 0, sessions: 0, session_list: [],
  })
  const [loading, setLoading] = useState(true)
  const [latency, setLatency] = useState<number | null>(null)
  const [lastRefresh, setLastRefresh] = useState<string>('')
  const [prevStats, setPrevStats] = useState<DeltaRecord | null>(getPreviousStats())
  const [operationLogs, setOperationLogs] = useState<OperationRecord[]>([])
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // 计数动画
  const animProjects = useCountUp(stats.projects, 800, !loading)
  const animTools = useCountUp(stats.tools, 800, !loading)
  const animClients = useCountUp(stats.clients, 800, !loading)
  const animSessions = useCountUp(stats.sessions, 800, !loading)

  const load = useCallback(async (isManual = false) => {
    const startTime = performance.now()
    try {
      const data = await getStats()
      const elapsed = Math.round(performance.now() - startTime)
      setStats(data)
      setLatency(elapsed)
      setLastRefresh(new Date().toLocaleTimeString('zh-CN'))
      saveCurrentStats(data)
      if (isManual) {
        toast.success('数据已刷新')
      }
    } catch (e: unknown) {
      toast.error('加载统计数据失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    // 每 10 秒自动刷新
    timerRef.current = setInterval(() => load(), 10000)
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [load])

  // 操作日志周期性读取
  useEffect(() => {
    setOperationLogs(getOperationLogs())
    const interval = setInterval(() => setOperationLogs(getOperationLogs()), 3000)
    return () => clearInterval(interval)
  }, [])

  // 存储上一次数据用于对比
  useEffect(() => {
    if (!loading && stats.projects + stats.tools + stats.clients + stats.sessions > 0) {
      const prev = getPreviousStats()
      if (prev) setPrevStats(prev)
    }
  }, [loading, stats])

  // ── 加载态 ──
  if (loading) {
    return (
      <div>
        <div className="page-header"><h2>仪表盘</h2></div>
        <Row gutter={[16, 16]}>
          {[1, 2, 3, 4].map(i => (
            <Col xs={24} sm={12} lg={6} key={i}>
              <Card><Skeleton active paragraph={{ rows: 1 }} /></Card>
            </Col>
          ))}
        </Row>
      </div>
    )
  }

  // ── 统计卡片配置 ──
  const cards = [
    {
      title: '项目数', value: animProjects, raw: stats.projects,
      icon: <FolderOutlined />, color: '#1677ff', bg: '#e6f4ff',
    },
    {
      title: '工具数', value: animTools, raw: stats.tools,
      icon: <ToolOutlined />, color: '#52c41a', bg: '#f6ffed',
    },
    {
      title: '客户端数', value: animClients, raw: stats.clients,
      icon: <KeyOutlined />, color: '#fa8c16', bg: '#fff7e6',
    },
    {
      title: '活跃 Session', value: animSessions, raw: stats.sessions,
      icon: <ApiOutlined />, color: '#eb2f96', bg: '#fff0f6',
    },
  ]

  // ── 系统健康状态 ──
  const isHealthy = latency !== null && latency < 3000
  const healthStatus = isHealthy ? '正常' : latency === null ? '未知' : '延迟偏高'
  const healthColor = isHealthy ? 'success' : latency === null ? 'default' : 'warning'

  return (
    <div>
      {/* ── 页面标题 ── */}
      <div className="page-header">
        <Space>
          <h2>仪表盘</h2>
          {lastRefresh && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              <ClockCircleOutlined style={{ marginRight: 4 }} />
              最后刷新: {lastRefresh}
            </Text>
          )}
        </Space>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => load(true)}
          size="small"
        >
          刷新
        </Button>
      </div>

      {/* ── 统计卡片 ── */}
      <Row gutter={[16, 16]} className="dashboard-section">
        {cards.map((c, i) => (
          <Col xs={24} sm={12} lg={6} key={i}>
            <Card className="stat-card" styles={{ body: { padding: 20 } }}>
              <Statistic
                title={
                  <span style={{ color: '#8c8c8c', fontSize: 13 }}>{c.title}</span>
                }
                value={c.value}
                valueStyle={{ fontSize: 32, fontWeight: 700 }}
                prefix={<span style={iconStyle(c.color, c.bg)}>{c.icon}</span>}
                suffix={
                  prevStats && deltaLabel(c.raw, prevStats[c.title === '项目数' ? 'projects' : c.title === '工具数' ? 'tools' : c.title === '客户端数' ? 'clients' : 'sessions' as keyof DeltaRecord]) ? (
                    <span style={{
                      fontSize: 13,
                      fontWeight: 500,
                      color: deltaColor(c.raw, prevStats[c.title === '项目数' ? 'projects' : c.title === '工具数' ? 'tools' : c.title === '客户端数' ? 'clients' : 'sessions' as keyof DeltaRecord]),
                    }}>
                      {deltaLabel(c.raw, prevStats[c.title === '项目数' ? 'projects' : c.title === '工具数' ? 'tools' : c.title === '客户端数' ? 'clients' : 'sessions' as keyof DeltaRecord])}
                    </span>
                  ) : null
                }
              />
            </Card>
          </Col>
        ))}
      </Row>

      {/* ── 第二行：系统健康 + 操作日志 ── */}
      <Row gutter={[16, 16]} className="dashboard-section">
        {/* 系统健康状态 */}
        <Col xs={24} lg={8}>
          <Card
            title={<Space><Badge status={healthColor as 'success' | 'default' | 'warning'} /><span>系统健康状态</span></Space>}
            styles={{ header: { borderBottom: '1px solid #f0f0f0' } }}
          >
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">Gateway 状态</Text>
                <Badge
                  status={isHealthy ? 'success' : 'warning'}
                  text={healthStatus}
                />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">响应延迟</Text>
                <Text strong>
                  {latency !== null ? `${latency}ms` : '—'}
                </Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">活跃连接</Text>
                <Text strong>{stats.sessions}</Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">统计项目</Text>
                <Text strong>
                  {stats.projects} 项目 · {stats.tools} 工具 · {stats.clients} 客户端
                </Text>
              </div>
            </Space>
          </Card>
        </Col>

        {/* 操作日志 */}
        <Col xs={24} lg={16}>
          <Card
            title={<Space><ClockCircleOutlined /><span>最近操作</span></Space>}
            styles={{ header: { borderBottom: '1px solid #f0f0f0' } }}
          >
            {operationLogs.length > 0 ? (
              <Timeline
                items={operationLogs.slice(0, 8).map(op => ({
                  dot: actionIcons[op.action] || <CheckOutlined />,
                  children: (
                    <div>
                      <Space size={4}>
                        <Tag style={{ fontSize: 11 }}>{op.action}</Tag>
                        <Text strong style={{ fontSize: 13 }}>{op.target}</Text>
                      </Space>
                      <Text type="secondary" style={{ fontSize: 12, marginLeft: 8 }}>
                        {op.detail}
                      </Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {op.timestamp}
                      </Text>
                    </div>
                  ),
                }))}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无操作记录，进行项目管理操作后将在此展示"
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* ── 第三行：会话表格 ── */}
      <Card
        title={<Space><ApiOutlined /><span>网关运行状态</span></Space>}
        styles={{ header: { borderBottom: '1px solid #f0f0f0' } }}
      >
        {stats.sessions > 0 ? (
          <Table
            columns={sessionColumns}
            dataSource={stats.session_list}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: '暂无活跃连接' }}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无活跃连接，MCP 客户端连接后将在此展示实时 Session 信息"
          />
        )}
      </Card>
    </div>
  )
}

export default Dashboard
