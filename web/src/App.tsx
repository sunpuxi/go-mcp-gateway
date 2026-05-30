import { Routes, Route, NavLink, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import Dashboard from './pages/Dashboard'
import Projects from './pages/Projects'
import Tools from './pages/Tools'
import Clients from './pages/Clients'
import Permissions from './pages/Permissions'
import Login from './pages/Login'

const navItems = [
  { path: '/', label: '📊 仪表盘', end: true },
  { path: '/projects', label: '📁 项目管理' },
  { path: '/tools', label: '🛠 工具管理' },
  { path: '/clients', label: '🔑 客户端管理' },
  { path: '/permissions', label: '🔐 权限管理' },
]

function AppLayout() {
  const { isAuthenticated, logout } = useAuth()

  if (!isAuthenticated) {
    return <Login />
  }

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div className="sidebar-header">MCP Gateway</div>
        <nav className="sidebar-nav">
          {navItems.map(item => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.end}
              className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div style={{ marginTop: 'auto', padding: '16px 20px', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
          <button
            onClick={logout}
            style={{
              background: 'none',
              border: 'none',
              color: 'rgba(255,255,255,0.6)',
              cursor: 'pointer',
              fontSize: 13,
              width: '100%',
              textAlign: 'left',
              padding: 0,
            }}
          >
            🚪 退出登录
          </button>
        </div>
      </aside>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/tools" element={<Tools />} />
          <Route path="/clients" element={<Clients />} />
          <Route path="/permissions" element={<Permissions />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  )
}

function App() {
  return (
    <AuthProvider>
      <AppLayout />
    </AuthProvider>
  )
}

export default App
