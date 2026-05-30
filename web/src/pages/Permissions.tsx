import { useState, useEffect, useCallback } from 'react'
import { Save, ShieldCheck } from 'lucide-react'
import {
  getClients, getTools, getProjects,
  getClientPermissions, updateClientPermissions,
  type Client, type Tool, type Project,
} from '../api'
import { useToast } from '../components/Toast'
import { TableSkeleton } from '../components/Skeleton'

function Permissions() {
  const [clients, setClients] = useState<Client[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedClient, setSelectedClient] = useState<string>('')
  const [draft, setDraft] = useState<number[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [permissionsLoading, setPermissionsLoading] = useState(false)
  const { toast } = useToast()

  const fetchData = useCallback(() => {
    setLoading(true)
    setError('')
    Promise.all([getClients(), getTools(), getProjects()])
      .then(([c, t, p]) => {
        setClients(c)
        setTools(t)
        setProjects(p)
        setLoading(false)
        if (c.length > 0) setSelectedClient(c[0].client_id)
      })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  useEffect(() => {
    if (!selectedClient) return
    setPermissionsLoading(true)
    getClientPermissions(selectedClient)
      .then(data => { setDraft(Array.isArray(data) ? data : []); setPermissionsLoading(false) })
      .catch(err => { setError(err.message); setPermissionsLoading(false) })
  }, [selectedClient])

  const selectClient = (clientId: string) => setSelectedClient(clientId)

  const toggleTool = (toolId: number) => {
    setDraft(prev => prev.includes(toolId) ? prev.filter(id => id !== toolId) : [...prev, toolId])
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await updateClientPermissions(selectedClient, draft)
      toast('success', '权限保存成功')
    } catch (err: any) {
      toast('error', '保存失败: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const currentClient = clients.find(c => c.client_id === selectedClient)

  const groupedTools = projects
    .map(proj => ({ ...proj, tools: tools.filter(t => t.project_id === proj.project_id) }))
    .filter(g => g.tools.length > 0)

  if (error) {
    return (
      <div>
        <div className="page-header"><h1>权限管理</h1><button className="btn btn-outline" onClick={fetchData}>重试</button></div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>加载失败：{error}</div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>权限管理</h1>
        <button className="btn btn-primary" onClick={handleSave} disabled={saving || !selectedClient}>
          <Save size={16} />{saving ? '保存中...' : '保存权限'}
        </button>
      </div>

      <div className="permission-layout">
        <div className="permission-client-list">
          <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 10, color: '#555' }}>选择客户端</div>
          {loading ? (
            Array.from({ length: 3 }, (_, i) => (
              <div key={i} style={{ padding: '10px 14px' }}><div className="skeleton" style={{ width: '80%', height: 14 }} /></div>
            ))
          ) : clients.length === 0 ? (
            <div style={{ color: '#7f8c9b', fontSize: 13, padding: '12px 0' }}>暂无客户端</div>
          ) : clients.map(c => (
            <div
              key={c.client_id}
              className={`permission-client-item ${selectedClient === c.client_id ? 'selected' : ''}`}
              onClick={() => selectClient(c.client_id)}
            >
              <ShieldCheck size={14} style={{ marginRight: 6, display: 'inline', verticalAlign: 'middle' }} />
              {c.name}
            </div>
          ))}
        </div>

        <div className="permission-tools">
          {!selectedClient ? (
            <div style={{ color: '#7f8c9b', padding: 20, textAlign: 'center' }}>请在左侧选择一个客户端</div>
          ) : permissionsLoading ? (
            <div className="table-card" style={{ padding: 16 }}><TableSkeleton rows={4} cols={1} /></div>
          ) : (
            <>
              <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 14 }}>当前客户端：{currentClient?.name}</div>
              <div className="table-card" style={{ padding: 16 }}>
                {groupedTools.length === 0 ? (
                  <div style={{ color: '#7f8c9b', textAlign: 'center', padding: 20 }}>暂无工具</div>
                ) : groupedTools.map(group => (
                  <div key={group.project_id} className="permission-group">
                    <div className="permission-group-title">📁 {group.name}</div>
                    {group.tools.map(tool => (
                      <label key={tool.tool_id} className="permission-tool-item">
                        <input type="checkbox" checked={draft.includes(tool.tool_id)} onChange={() => toggleTool(tool.tool_id)} />
                        <span style={{ fontWeight: 500 }}>{tool.name}</span>
                        <span style={{ color: '#7f8c9b', fontSize: 13 }}>{tool.title}</span>
                      </label>
                    ))}
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default Permissions
