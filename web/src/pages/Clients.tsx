import { useState, useEffect, useCallback } from 'react'
import { getClients, createClient, updateClient, deleteClient, generateApiKey, type Client, type ClientForm } from '../api'

function Clients() {
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [showKeyModal, setShowKeyModal] = useState(false)
  const [newApiKey, setNewApiKey] = useState('')
  const [editing, setEditing] = useState<Client | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<ClientForm>({ client_id: '', name: '', description: '', status: 1 })

  const fetchClients = useCallback(() => {
    setLoading(true)
    setError('')
    getClients()
      .then(data => { setClients(data); setLoading(false) })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  useEffect(() => { fetchClients() }, [fetchClients])

  const openCreate = () => {
    setEditing(null)
    setForm({ client_id: '', name: '', description: '', status: 1 })
    setShowForm(true)
  }

  const openEdit = (c: Client) => {
    setEditing(c)
    setForm({ client_id: c.client_id, name: c.name, description: c.description, status: c.status })
    setShowForm(true)
  }

  const handleGenerateKey = async (client: Client) => {
    try {
      const result = await generateApiKey(client.client_id)
      setNewApiKey(result.api_key || '')
      setShowKeyModal(true)
      fetchClients()
    } catch (err: any) {
      alert('生成 API Key 失败: ' + err.message)
    }
  }

  const handleSave = async () => {
    if (!form.client_id.trim() || !form.name.trim()) {
      alert('请填写必填字段')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        const { client_id, ...data } = form
        const updated = await updateClient(editing.client_id, data)
        setClients(clients.map(c => c.client_id === editing.client_id ? updated : c))
      } else {
        const created = await createClient(form)
        setClients([...clients, created])
      }
      setShowForm(false)
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除该客户端？')) return
    try {
      await deleteClient(id)
      setClients(clients.filter(c => c.client_id !== id))
    } catch (err: any) {
      alert('删除失败: ' + err.message)
    }
  }

  const toggleStatus = async (c: Client) => {
    try {
      const updated = await updateClient(c.client_id, { status: c.status === 1 ? 0 : 1 })
      setClients(clients.map(x => x.client_id === c.client_id ? updated : x))
    } catch (err: any) {
      alert('操作失败: ' + err.message)
    }
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(newApiKey).then(() => {
      alert('已复制到剪贴板')
    })
  }

  if (loading) {
    return <div><div className="page-header"><h1>客户端管理</h1></div><div style={{ textAlign: 'center', padding: 60, color: '#7f8c9b' }}>加载中...</div></div>
  }

  if (error) {
    return (
      <div>
        <div className="page-header"><h1>客户端管理</h1><button className="btn btn-outline" onClick={fetchClients}>重试</button></div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>加载失败：{error}</div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>客户端管理</h1>
        <button className="btn btn-primary" onClick={openCreate}>+ 新建客户端</button>
      </div>

      <div className="table-card">
        <table>
          <thead>
            <tr>
              <th>Client ID</th>
              <th>名称</th>
              <th>Key 前缀</th>
              <th>已授权工具</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {clients.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: 40, color: '#7f8c9b' }}>暂无客户端</td></tr>
            ) : clients.map(c => (
              <tr key={c.client_id}>
                <td><strong>{c.client_id}</strong></td>
                <td>{c.name}</td>
                <td><code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 4, fontSize: 13 }}>{c.api_key_prefix || 'sk-***'}-***</code></td>
                <td>{c.tool_count != null ? c.tool_count : '-'}</td>
                <td><span className={`status-badge ${c.status === 1 ? 'active' : 'inactive'}`}>{c.status === 1 ? '启用' : '禁用'}</span></td>
                <td style={{ fontSize: 13, color: '#7f8c9b' }}>{c.created_at ? new Date(c.created_at).toLocaleString('zh-CN') : '-'}</td>
                <td>
                  <button className="btn btn-sm btn-primary" style={{ marginRight: 4 }} onClick={() => handleGenerateKey(c)}>生成 Key</button>
                  <button className="btn-icon" onClick={() => toggleStatus(c)} title="切换状态">⚡</button>
                  <button className="btn-icon" onClick={() => openEdit(c)} title="编辑">✏️</button>
                  <button className="btn-icon" onClick={() => handleDelete(c.client_id)} title="删除">🗑️</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showForm && (
        <div className="form-overlay" onClick={() => setShowForm(false)}>
          <div className="form-panel" onClick={e => e.stopPropagation()}>
            <div className="form-panel-header">
              {editing ? '编辑客户端' : '新建客户端'}
              <button className="btn-icon" onClick={() => setShowForm(false)}>✕</button>
            </div>
            <div className="form-panel-body">
              <div className="form-group">
                <label>Client ID <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.client_id} onChange={e => setForm({ ...form, client_id: e.target.value })} disabled={!!editing} placeholder="如 cli_bigdata" />
              </div>
              <div className="form-group">
                <label>名称 <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="如 大数据项目组" />
              </div>
              <div className="form-group">
                <label>描述</label>
                <textarea value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} />
              </div>
              <div className="form-group">
                <label>状态</label>
                <select value={form.status} onChange={e => setForm({ ...form, status: Number(e.target.value) })}>
                  <option value={1}>启用</option>
                  <option value={0}>禁用</option>
                </select>
              </div>
            </div>
            <div className="form-panel-footer">
              <button className="btn btn-outline" onClick={() => setShowForm(false)}>取消</button>
              <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}

      {showKeyModal && (
        <div className="form-overlay" onClick={() => setShowKeyModal(false)}>
          <div className="form-panel" style={{ width: 480 }} onClick={e => e.stopPropagation()}>
            <div className="form-panel-header">
              API Key 已生成
              <button className="btn-icon" onClick={() => setShowKeyModal(false)}>✕</button>
            </div>
            <div className="form-panel-body">
              <div className="api-key-modal">
                <div className="api-key-display">{newApiKey}</div>
                <div className="api-key-warning">⚠ 请立即复制保存，关闭后将无法再次查看</div>
                <div className="form-panel-footer" style={{ justifyContent: 'center', border: 'none', padding: 0 }}>
                  <button className="btn btn-primary" onClick={copyToClipboard}>📋 复制 Key</button>
                  <button className="btn btn-outline" onClick={() => setShowKeyModal(false)}>关闭</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Clients
