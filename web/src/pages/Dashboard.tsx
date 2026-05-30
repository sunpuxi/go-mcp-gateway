import { useState, useEffect } from 'react'
import { FolderKanban, Wrench, KeyRound, Radio } from 'lucide-react'
import { getStats, type Stats } from '../api'
import { StatsSkeleton } from '../components/Skeleton'

const cards = [
  { key: 'projects' as const, label: '项目数', icon: FolderKanban, color: '#1976d2' },
  { key: 'tools' as const, label: '工具数', icon: Wrench, color: '#7b1fa2' },
  { key: 'clients' as const, label: '客户端数', icon: KeyRound, color: '#e65100' },
  { key: 'sessions' as const, label: '活跃 Session', icon: Radio, color: '#2e7d32' },
]

function Dashboard() {
  const [stats, setStats] = useState<Stats>({ projects: 0, tools: 0, clients: 0, sessions: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    getStats()
      .then(data => { setStats(data); setLoading(false) })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  return (
    <div>
      <div className="page-header">
        <h1>仪表盘</h1>
      </div>

      {loading ? (
        <StatsSkeleton />
      ) : error ? (
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>
          加载失败：{error}
        </div>
      ) : (
        <div className="stats-grid">
          {cards.map(card => (
            <div key={card.key} className="stat-card">
              <div className="stat-card-icon" style={{ background: `${card.color}15` }}>
                <card.icon size={22} color={card.color} />
              </div>
              <div className="stat-card-body">
                <div className="stat-value">{stats[card.key]}</div>
                <div className="stat-label">{card.label}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="table-card">
        <div style={{ padding: '16px', fontWeight: 600, borderBottom: '1px solid #eee' }}>
          网关运行状态
        </div>
        {loading ? (
          <div style={{ padding: '40px 16px', display: 'flex', justifyContent: 'center' }}>
            <span style={{ color: '#7f8c9b', fontSize: 14 }}>加载中...</span>
          </div>
        ) : (
          <div style={{ padding: '40px 16px', textAlign: 'center', color: '#666', fontSize: 14, lineHeight: 1.8 }}>
            网关运行中<br />
            <span style={{ color: '#94a3b8' }}>项目 {stats.projects} 个，工具 {stats.tools} 个，客户端 {stats.clients} 个</span>
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
