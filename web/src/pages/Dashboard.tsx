import { useState, useEffect } from 'react'
import { getStats, type Stats } from '../api'

function Dashboard() {
  const [stats, setStats] = useState<Stats>({ projects: 0, tools: 0, clients: 0, sessions: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    getStats()
      .then(data => {
        setStats(data)
        setLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) {
    return (
      <div>
        <div className="page-header"><h1>仪表盘</h1></div>
        <div style={{ textAlign: 'center', padding: 60, color: '#7f8c9b' }}>加载中...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <div className="page-header"><h1>仪表盘</h1></div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>
          加载失败：{error}
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>仪表盘</h1>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-value">{stats.projects}</div>
          <div className="stat-label">项目数</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.tools}</div>
          <div className="stat-label">工具数</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.clients}</div>
          <div className="stat-label">客户端数</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.sessions}</div>
          <div className="stat-label">活跃 Session</div>
        </div>
      </div>

      <div className="table-card">
        <div style={{ padding: '16px', fontWeight: 600, borderBottom: '1px solid #eee' }}>
          网关运行状态
        </div>
        <div style={{ padding: '40px 16px', textAlign: 'center', color: '#7f8c9b', fontSize: 14 }}>
          网关运行中 — 项目 {stats.projects} 个，工具 {stats.tools} 个，客户端 {stats.clients} 个
        </div>
      </div>
    </div>
  )
}

export default Dashboard
