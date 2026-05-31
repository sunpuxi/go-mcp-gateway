import { useState, Suspense, lazy, useEffect, useCallback } from 'react'
import { Routes, Route, useNavigate, useLocation, Navigate } from 'react-router-dom'
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Space, Typography, Grid, Spin } from 'antd'
import {
  DashboardOutlined,
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  BookOutlined,
  SearchOutlined,
  GithubOutlined,
  UserOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SunOutlined,
  MoonOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuth } from './context/AuthContext'
import { useTheme } from './context/ThemeContext'
import ErrorBoundary from './components/ErrorBoundary'
import GlobalSearch from './components/GlobalSearch'
import Logo from './components/Logo'

// 懒加载页面组件
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Projects = lazy(() => import('./pages/Projects'))
const Tools = lazy(() => import('./pages/Tools'))
const Clients = lazy(() => import('./pages/Clients'))
const Permissions = lazy(() => import('./pages/Permissions'))
const Help = lazy(() => import('./pages/Help'))
const Login = lazy(() => import('./pages/Login'))

const { Header, Sider, Content } = Layout
const { useBreakpoint } = Grid

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/projects', icon: <FolderOutlined />, label: '项目管理' },
  { key: '/tools', icon: <ToolOutlined />, label: '工具管理' },
  { key: '/clients', icon: <KeyOutlined />, label: '客户端管理' },
  { key: '/permissions', icon: <SafetyCertificateOutlined />, label: '权限管理' },
  { key: '/help', icon: <BookOutlined />, label: '使用帮助' },
]

const pathLabels: Record<string, string> = {
  '/': '仪表盘',
  '/projects': '项目管理',
  '/tools': '工具管理',
  '/clients': '客户端管理',
  '/permissions': '权限管理',
  '/help': '使用帮助',
}

/** 页面加载占位 */
function PageFallback() {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
      <Spin size="large" tip="加载中…" />
    </div>
  )
}

/** 需要认证才能访问的布局 */
function AuthenticatedLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { logout } = useAuth()
  const { isDark, toggleTheme } = useTheme()
  const [collapsed, setCollapsed] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)
  const screens = useBreakpoint()

  // Ctrl+K 全局搜索快捷键
  const handleGlobalKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
      e.preventDefault()
      setSearchOpen(prev => !prev)
    }
  }, [])

  useEffect(() => {
    window.addEventListener('keydown', handleGlobalKeyDown)
    return () => window.removeEventListener('keydown', handleGlobalKeyDown)
  }, [handleGlobalKeyDown])

  const selectedKey = '/' + location.pathname.split('/')[1]
  const currentLabel = pathLabels[selectedKey] || ''

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  const userMenuItems = [
    { key: 'docs', label: '使用文档', icon: <BookOutlined /> },
    { type: 'divider' as const },
    { key: 'logout', label: '退出登录', icon: <LogoutOutlined />, danger: true },
  ]

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      handleLogout()
    } else if (key === 'docs') {
      navigate('/help')
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        width={220}
        theme="dark"
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        breakpoint="lg"
        trigger={null}
      >
        <div className="logo">
          <Logo size={collapsed ? 28 : 34} />
          {!collapsed && <span>MCP Gateway</span>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ borderInlineEnd: 'none' }}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
            boxShadow: '0 1px 4px rgba(0,0,0,0.04)',
            position: 'sticky',
            top: 0,
            zIndex: 10,
          }}
        >
          <Space>
            {!screens.lg ? (
              <MenuFoldOutlined
                onClick={() => setCollapsed(false)}
                style={{ fontSize: 18, color: '#666', cursor: 'pointer' }}
              />
            ) : collapsed ? (
              <MenuUnfoldOutlined
                onClick={() => setCollapsed(false)}
                style={{ fontSize: 18, color: '#666', cursor: 'pointer' }}
              />
            ) : (
              <MenuFoldOutlined
                onClick={() => setCollapsed(true)}
                style={{ fontSize: 18, color: '#666', cursor: 'pointer' }}
              />
            )}
            <Breadcrumb
              items={[
                { title: 'Home' },
                { title: currentLabel },
              ]}
            />
          </Space>

          <Space size={16}>
            {/* 全局搜索 */}
            <span
              onClick={() => setSearchOpen(true)}
              style={{ fontSize: 18, color: '#666', cursor: 'pointer', lineHeight: 0 }}
              title="全局搜索 (Ctrl+K)"
            >
              <SearchOutlined />
            </span>
            {/* 主题切换 */}
            <span
              onClick={toggleTheme}
              style={{ fontSize: 18, color: '#666', cursor: 'pointer', lineHeight: 0 }}
              title={isDark ? '切换亮色模式' : '切换暗色模式'}
            >
              {isDark ? <SunOutlined /> : <MoonOutlined />}
            </span>

            <a
              href="https://github.com/sunpuxi/go-mcp-gateway"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: '#666', fontSize: 18, lineHeight: 0 }}
              className="gh-link"
            >
              <GithubOutlined />
            </a>
            <Dropdown
              menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
            >
              <Space style={{ cursor: 'pointer', color: '#333' }}>
                <Avatar size={28} icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
                <Typography.Text style={{ color: '#333' }}>管理员</Typography.Text>
              </Space>
            </Dropdown>
          </Space>
        </Header>

        <Content style={{ margin: 24 }}>
          <ErrorBoundary>
            <Suspense fallback={<PageFallback />}>
              <div key={location.pathname} className="page-fade-enter">
                <Routes>
                <Route path="/" element={<Dashboard />} />
                <Route path="/projects" element={<Projects />} />
                <Route path="/tools" element={<Tools />} />
                <Route path="/clients" element={<Clients />} />
                <Route path="/permissions" element={<Permissions />} />
                <Route path="/help" element={<Help />} />
                {/* 404 */}
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
              </div>
            </Suspense>
          </ErrorBoundary>
        </Content>
      </Layout>

      {/* 全局搜索 */}
      <GlobalSearch open={searchOpen} onClose={() => setSearchOpen(false)} />
    </Layout>
  )
}

function App() {
  const { isAuthenticated } = useAuth()

  return (
    <Routes>
      <Route
        path="/login"
        element={
          isAuthenticated ? <Navigate to="/" replace /> : (
            <Suspense fallback={<PageFallback />}>
              <Login />
            </Suspense>
          )
        }
      />
      <Route
        path="/*"
        element={
          isAuthenticated ? <AuthenticatedLayout /> : <Navigate to="/login" replace />
        }
      />
    </Routes>
  )
}

export default App
