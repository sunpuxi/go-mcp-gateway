import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, X, Wrench } from 'lucide-react'
import { getTools, createTool, updateTool, deleteTool, getProjects, type Tool, type ParamRule, type Project } from '../api'
import { useToast } from '../components/Toast'
import { Confirm } from '../components/Confirm'
import { TableSkeleton } from '../components/Skeleton'

const emptyParam = (): ParamRule => ({ name: '', type: 'string', location: 'query', required: false, default_value: '', description: '' })
const emptyForm = (projectId: string = ''): Tool => ({
  tool_id: 0, name: '', title: '', description: '', project_id: projectId,
  http_method: 'GET', url_template: '', timeout_ms: 5000, status: 1, params: [],
})

function Tools() {
  const [tools, setTools] = useState<Tool[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Tool | null>(null)
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null)
  const [form, setForm] = useState<Tool>(emptyForm())
  const { toast } = useToast()

  const fetchData = useCallback(() => {
    setLoading(true)
    setError('')
    Promise.all([getTools(), getProjects()])
      .then(([toolsData, projectsData]) => { setTools(toolsData); setProjects(projectsData); setLoading(false) })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  const openCreate = () => {
    setEditing(null)
    const f = emptyForm(projects.length > 0 ? projects[0].project_id : '')
    setForm(f)
    setShowForm(true)
  }

  const openEdit = (t: Tool) => {
    setEditing(t)
    setForm({ ...t, params: t.params?.map((p: any) => ({ ...p })) || [] })
    setShowForm(true)
  }

  const addParam = () => setForm({ ...form, params: [...form.params, emptyParam()] })
  const removeParam = (idx: number) => setForm({ ...form, params: form.params.filter((_, i) => i !== idx) })
  const updateParam = (idx: number, field: keyof ParamRule, value: string | boolean) => {
    setForm({ ...form, params: form.params.map((p, i) => i === idx ? { ...p, [field]: value } : p) })
  }

  const handleSave = async () => {
    if (!form.name.trim() || !form.title.trim() || !form.url_template.trim() || !form.project_id) {
      toast('warning', '请填写必填字段')
      return
    }
    setSaving(true)
    try {
      const { tool_id, ...data } = form
      if (editing) {
        const updated = await updateTool(editing.tool_id, data)
        setTools(tools.map(t => t.tool_id === editing.tool_id ? updated : t))
        toast('success', '工具已更新')
      } else {
        const created = await createTool(data)
        setTools([...tools, created])
        toast('success', '工具已创建')
      }
      setShowForm(false)
    } catch (err: any) {
      toast('error', '保存失败: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (confirmDelete == null) return
    try {
      await deleteTool(confirmDelete)
      setTools(tools.filter(t => t.tool_id !== confirmDelete))
      toast('success', '工具已删除')
    } catch (err: any) {
      toast('error', '删除失败: ' + err.message)
    }
    setConfirmDelete(null)
  }

  const getProjectName = (projectId: string) => projects.find(p => p.project_id === projectId)?.name || projectId

  if (error) {
    return (
      <div>
        <div className="page-header"><h1>工具管理</h1><button className="btn btn-outline" onClick={fetchData}>重试</button></div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>加载失败：{error}</div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>工具管理</h1>
        <button className="btn btn-primary" onClick={openCreate}><Plus size={16} />新建工具</button>
      </div>

      <div className="table-card">
        {loading ? <TableSkeleton rows={5} cols={8} /> : tools.length === 0 ? (
          <div className="empty-state"><Wrench size={40} /><p>暂无工具</p><p className="empty-hint">点击「新建工具」配置 MCP 协议转换规则</p></div>
        ) : (
          <table>
            <thead>
              <tr><th>工具名称</th><th>标题</th><th>所属项目</th><th>HTTP 方法</th><th>URL 模板</th><th>参数</th><th>状态</th><th>操作</th></tr>
            </thead>
            <tbody>
              {tools.map(t => (
                <tr key={t.tool_id}>
                  <td><strong>{t.name}</strong></td>
                  <td>{t.title}</td>
                  <td>{getProjectName(t.project_id)}</td>
                  <td><span style={{ fontWeight: 600 }}>{t.http_method}</span></td>
                  <td style={{ fontFamily: 'monospace', fontSize: 13, maxWidth: 180, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.url_template}</td>
                  <td>{t.params?.length || 0}</td>
                  <td><span className={`status-badge ${t.status === 1 ? 'active' : 'inactive'}`}>{t.status === 1 ? '启用' : '禁用'}</span></td>
                  <td>
                    <button className="btn-icon" onClick={() => openEdit(t)} title="编辑"><Pencil size={15} /></button>
                    <button className="btn-icon" onClick={() => setConfirmDelete(t.tool_id)} title="删除"><Trash2 size={15} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showForm && (
        <div className="form-overlay" onClick={() => setShowForm(false)}>
          <div className="form-panel" style={{ width: 720 }} onClick={e => e.stopPropagation()}>
            <div className="form-panel-header">
              {editing ? '编辑工具' : '新建工具'}
              <button className="btn-icon" onClick={() => setShowForm(false)}><X size={16} /></button>
            </div>
            <div className="form-panel-body">
              <div className="form-row">
                <div className="form-group">
                  <label>工具名称 <span style={{ color: '#e53935' }}>*</span></label>
                  <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="get_user" />
                </div>
                <div className="form-group">
                  <label>标题 <span style={{ color: '#e53935' }}>*</span></label>
                  <input value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} placeholder="获取用户信息" />
                </div>
              </div>
              <div className="form-group">
                <label>描述</label>
                <textarea value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>所属项目 <span style={{ color: '#e53935' }}>*</span></label>
                  <select value={form.project_id} onChange={e => setForm({ ...form, project_id: e.target.value })}>
                    <option value="">请选择项目</option>
                    {projects.map(p => <option key={p.project_id} value={p.project_id}>{p.name}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>HTTP 方法</label>
                  <select value={form.http_method} onChange={e => setForm({ ...form, http_method: e.target.value })}>
                    {['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => <option key={m} value={m}>{m}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>超时(ms)</label>
                  <input type="number" value={form.timeout_ms} onChange={e => setForm({ ...form, timeout_ms: Number(e.target.value) })} />
                </div>
              </div>
              <div className="form-group">
                <label>URL 模板 <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.url_template} onChange={e => setForm({ ...form, url_template: e.target.value })} placeholder="/users/{user_id}" />
              </div>
              <div className="form-group">
                <label>状态</label>
                <select value={form.status} onChange={e => setForm({ ...form, status: Number(e.target.value) })}>
                  <option value={1}>启用</option><option value={0}>禁用</option>
                </select>
              </div>

              <div style={{ marginTop: 20 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <label style={{ fontWeight: 600, fontSize: 14 }}>参数映射规则</label>
                  <button className="btn btn-sm btn-outline" onClick={addParam}><Plus size={14} />添加参数</button>
                </div>
                {form.params.length > 0 ? (
                  <table className="param-table">
                    <thead>
                      <tr><th style={{ width: '16%' }}>参数名</th><th style={{ width: '10%' }}>类型</th><th style={{ width: '14%' }}>位置</th><th style={{ width: '8%' }}>必填</th><th style={{ width: '14%' }}>默认值</th><th style={{ width: '28%' }}>描述</th><th style={{ width: '10%' }}>操作</th></tr>
                    </thead>
                    <tbody>
                      {form.params.map((p, idx) => (
                        <tr key={idx}>
                          <td><input value={p.name} onChange={e => updateParam(idx, 'name', e.target.value)} placeholder="参数名" /></td>
                          <td>
                            <select value={p.type} onChange={e => updateParam(idx, 'type', e.target.value)}>
                              <option value="string">string</option><option value="number">number</option><option value="boolean">boolean</option><option value="object">object</option>
                            </select>
                          </td>
                          <td>
                            <select value={p.location} onChange={e => updateParam(idx, 'location', e.target.value)}>
                              <option value="path">path</option><option value="query">query</option><option value="body">body</option><option value="header">header</option>
                            </select>
                          </td>
                          <td style={{ textAlign: 'center' }}><input type="checkbox" checked={p.required} onChange={e => updateParam(idx, 'required', e.target.checked)} /></td>
                          <td><input value={p.default_value} onChange={e => updateParam(idx, 'default_value', e.target.value)} /></td>
                          <td><input value={p.description} onChange={e => updateParam(idx, 'description', e.target.value)} placeholder="参数说明" /></td>
                          <td><button className="btn-icon" onClick={() => removeParam(idx)}><Trash2 size={15} /></button></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                ) : (
                  <div style={{ color: '#7f8c9b', fontSize: 13, padding: '12px 0' }}>暂无参数规则，点击「添加参数」开始配置</div>
                )}
              </div>
            </div>
            <div className="form-panel-footer">
              <button className="btn btn-outline" onClick={() => setShowForm(false)}>取消</button>
              <button className="btn btn-primary" onClick={handleSave} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
            </div>
          </div>
        </div>
      )}

      {confirmDelete != null && (
        <Confirm title="删除工具" message={`确定删除该工具吗？此操作不可撤销。`} onConfirm={handleDelete} onCancel={() => setConfirmDelete(null)} confirmText="删除" danger />
      )}
    </div>
  )
}

export default Tools
