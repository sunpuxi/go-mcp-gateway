import { Routes, Route, NavLink, Navigate } from 'react-router-dom'
import { LayoutDashboard, FolderKanban, Wrench, KeyRound, ShieldCheck } from 'lucide-react'
import { AuthProvider, useAuth } from './context/AuthContext'
import { ToastProvider } from './components/Toast'
import Dashboard from './pages/Dashboard'
import Projects from './pages/Projects'
import Tools from './pages/Tools'
import Clients from './pages/Clients'
import Permissions from './pages/Permissions'
import Login from './pages/Login'

const navItems = [
  { path: '/', label: '仪表盘', icon: LayoutDashboard, end: true },
  { path: '/projects', label: '项目管理', icon: FolderKanban },
  { path: '/tools', label: '工具管理', icon: Wrench },
  { path: '/clients', label: '客户端管理', icon: KeyRound },
  { path: '/permissions', label: '权限管理', icon: ShieldCheck },
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
              <item.icon size={18} />
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
              display: 'flex',
              alignItems: 'center',
              gap: 8,
            }}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
            退出登录
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
      <ToastProvider>
        <AppLayout />
      </ToastProvider>
    </AuthProvider>
  )
}

export default App
