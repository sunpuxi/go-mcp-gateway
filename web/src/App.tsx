import { Routes, Route, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  DashboardOutlined,
  FolderOutlined,
  ToolOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  BookOutlined,
  GithubOutlined,
} from '@ant-design/icons'
import Dashboard from './pages/Dashboard'
import Projects from './pages/Projects'
import Tools from './pages/Tools'
import Clients from './pages/Clients'
import Permissions from './pages/Permissions'
import Help from './pages/Help'
import Logo from './components/Logo'

const { Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/projects', icon: <FolderOutlined />, label: '项目管理' },
  { key: '/tools', icon: <ToolOutlined />, label: '工具管理' },
  { key: '/clients', icon: <KeyOutlined />, label: '客户端管理' },
  { key: '/permissions', icon: <SafetyCertificateOutlined />, label: '权限管理' },
  { key: '/help', icon: <BookOutlined />, label: '使用帮助' },
]

function App() {
  const navigate = useNavigate()
  const location = useLocation()

  const selectedKey = '/' + location.pathname.split('/')[1]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={220} theme="dark">
        <div className="logo">
          <Logo size={34} />
          <span>MCP Gateway</span>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ flex: 1 }}
        />
        <div style={{
          padding: '12px 16px',
          borderTop: '1px solid rgba(255,255,255,0.1)',
        }}>
          <a
            href="https://github.com/sunpuxi/go-mcp-gateway"
            target="_blank"
            rel="noopener noreferrer"
            style={{
              color: 'rgba(255,255,255,0.55)',
              textDecoration: 'none',
              fontSize: 14,
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              transition: 'color 0.2s',
            }}
            onMouseEnter={e => (e.currentTarget.style.color = '#fff')}
            onMouseLeave={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.55)')}
          >
            <GithubOutlined style={{ fontSize: 16 }} />
            源码仓库
          </a>
        </div>
      </Sider>
      <Layout>
        <Content style={{ margin: 24 }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/projects" element={<Projects />} />
            <Route path="/tools" element={<Tools />} />
            <Route path="/clients" element={<Clients />} />
            <Route path="/permissions" element={<Permissions />} />
            <Route path="/help" element={<Help />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  )
}

export default App
