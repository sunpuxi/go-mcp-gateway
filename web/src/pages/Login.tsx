import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Input, Button, Typography, Alert, Space } from 'antd'
import { KeyOutlined, LoginOutlined } from '@ant-design/icons'
import { useAuth } from '../context/AuthContext'
import Logo from '../components/Logo'

const { Title, Text } = Typography

function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [apiKey, setApiKey] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    if (!apiKey.trim()) {
      setError('请输入 Admin API Key')
      return
    }
    setError('')
    setLoading(true)

    // 模拟异步登录过程，同时给 UI 一点反馈时间
    await new Promise(resolve => setTimeout(resolve, 300))
    login(apiKey.trim())
    navigate('/', { replace: true })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit()
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--color-bg-layout, #f5f7fa)',
      }}
    >
      <Card
        style={{
          width: 420,
          borderRadius: 12,
          boxShadow: '0 4px 24px rgba(0,0,0,0.08)',
        }}
        styles={{ body: { padding: '40px 36px' } }}
      >
        <Space direction="vertical" size={24} style={{ width: '100%' }}>
          {/* 品牌标识 */}
          <div style={{ textAlign: 'center' }}>
            <Space direction="vertical" size={12}>
              <Logo size={48} />
              <div>
                <Title level={3} style={{ margin: 0, fontWeight: 700 }}>
                  MCP Gateway
                </Title>
                <Text type="secondary">管理后台</Text>
              </div>
            </Space>
          </div>

          {/* 输入区域 */}
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <div>
              <Text strong style={{ display: 'block', marginBottom: 8, fontSize: 13 }}>
                Admin API Key
              </Text>
              <Input.Password
                prefix={<KeyOutlined style={{ color: '#bfbfbf' }} />}
                value={apiKey}
                onChange={e => { setApiKey(e.target.value); setError('') }}
                onKeyDown={handleKeyDown}
                placeholder="请输入 admin_api_key"
                size="large"
                autoFocus
                allowClear
              />
            </div>

            {error && (
              <Alert
                message={error}
                type="error"
                showIcon
                closable
                onClose={() => setError('')}
              />
            )}

            <Button
              type="primary"
              block
              size="large"
              icon={<LoginOutlined />}
              loading={loading}
              onClick={handleSubmit}
              style={{ height: 44, fontSize: 15, fontWeight: 500 }}
            >
              登 录
            </Button>
          </Space>

          <Text
            type="secondary"
            style={{ display: 'block', textAlign: 'center', fontSize: 12 }}
          >
            请使用管理员分配的 API Key 登录管理后台
          </Text>
        </Space>
      </Card>
    </div>
  )
}

export default Login
