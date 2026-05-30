import { useState } from 'react'
import { useAuth } from '../context/AuthContext'

function Login() {
  const { login } = useAuth()
  const [apiKey, setApiKey] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!apiKey.trim()) {
      setError('请输入 API Key')
      return
    }
    setError('')
    login(apiKey.trim())
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: '#f5f6fa',
    }}>
      <div style={{
        background: '#fff',
        borderRadius: 12,
        padding: '40px 36px',
        width: 400,
        boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
      }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#1a1a2e', marginBottom: 8 }}>
            MCP Gateway
          </h1>
          <p style={{ fontSize: 14, color: '#7f8c9b' }}>管理后台登录</p>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Admin API Key</label>
            <input
              type="password"
              value={apiKey}
              onChange={e => setApiKey(e.target.value)}
              placeholder="请输入 admin_api_key"
              autoFocus
            />
          </div>

          {error && (
            <div style={{
              color: '#c62828',
              fontSize: 13,
              marginBottom: 16,
              background: '#fbe9e7',
              padding: '8px 12px',
              borderRadius: 6,
            }}>
              {error}
            </div>
          )}

          <button
            type="submit"
            className="btn btn-primary"
            style={{ width: '100%', justifyContent: 'center', padding: '10px 0', fontSize: 14 }}
          >
            登录
          </button>
        </form>
      </div>
    </div>
  )
}

export default Login
