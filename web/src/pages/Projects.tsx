import { useState, useEffect, useCallback } from 'react'
import { getProjects, createProject, updateProject, deleteProject, type Project, type ProjectForm } from '../api'

function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Project | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<ProjectForm>({ project_id: '', name: '', base_url: '', description: '', status: 1 })

  const fetchProjects = useCallback(() => {
    setLoading(true)
    setError('')
    getProjects()
      .then(data => { setProjects(data); setLoading(false) })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  useEffect(() => { fetchProjects() }, [fetchProjects])

  const openCreate = () => {
    setEditing(null)
    setForm({ project_id: '', name: '', base_url: '', description: '', status: 1 })
    setShowForm(true)
  }

  const openEdit = (p: Project) => {
    setEditing(p)
    setForm({ project_id: p.project_id, name: p.name, base_url: p.base_url, description: p.description, status: p.status })
    setShowForm(true)
  }

  const handleSave = async () => {
    if (!form.project_id.trim() || !form.name.trim() || !form.base_url.trim()) {
      alert('请填写必填字段')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        const { project_id, ...data } = form
        const updated = await updateProject(editing.project_id, data)
        setProjects(projects.map(p => p.project_id === editing.project_id ? updated : p))
      } else {
        const created = await createProject(form)
        setProjects([...projects, created])
      }
      setShowForm(false)
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除该项目？')) return
    try {
      await deleteProject(id)
      setProjects(projects.filter(p => p.project_id !== id))
    } catch (err: any) {
      alert('删除失败: ' + err.message)
    }
  }

  const toggleStatus = async (p: Project) => {
    try {
      const updated = await updateProject(p.project_id, { status: p.status === 1 ? 0 : 1 })
      setProjects(projects.map(x => x.project_id === p.project_id ? updated : x))
    } catch (err: any) {
      alert('操作失败: ' + err.message)
    }
  }

  if (loading) {
    return <div><div className="page-header"><h1>项目管理</h1></div><div style={{ textAlign: 'center', padding: 60, color: '#7f8c9b' }}>加载中...</div></div>
  }

  if (error) {
    return (
      <div>
        <div className="page-header"><h1>项目管理</h1><button className="btn btn-outline" onClick={fetchProjects}>重试</button></div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>加载失败：{error}</div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>项目管理</h1>
        <button className="btn btn-primary" onClick={openCreate}>+ 新建项目</button>
      </div>

      <div className="table-card">
        <table>
          <thead>
            <tr>
              <th>Project ID</th>
              <th>名称</th>
              <th>基础 URL</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {projects.length === 0 ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: 40, color: '#7f8c9b' }}>暂无项目</td></tr>
            ) : projects.map(p => (
              <tr key={p.project_id}>
                <td><strong>{p.project_id}</strong></td>
                <td>{p.name}</td>
                <td style={{ maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{p.base_url}</td>
                <td>
                  <span className={`status-badge ${p.status === 1 ? 'active' : 'inactive'}`}>
                    {p.status === 1 ? '启用' : '禁用'}
                  </span>
                </td>
                <td>
                  <button className="btn-icon" onClick={() => toggleStatus(p)} title="切换状态">⚡</button>
                  <button className="btn-icon" onClick={() => openEdit(p)} title="编辑">✏️</button>
                  <button className="btn-icon" onClick={() => handleDelete(p.project_id)} title="删除">🗑️</button>
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
              {editing ? '编辑项目' : '新建项目'}
              <button className="btn-icon" onClick={() => setShowForm(false)}>✕</button>
            </div>
            <div className="form-panel-body">
              <div className="form-group">
                <label>Project ID <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.project_id} onChange={e => setForm({ ...form, project_id: e.target.value })} disabled={!!editing} placeholder="如 proj_user" />
              </div>
              <div className="form-group">
                <label>名称 <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="如 用户服务" />
              </div>
              <div className="form-group">
                <label>基础 URL <span style={{ color: '#e53935' }}>*</span></label>
                <input value={form.base_url} onChange={e => setForm({ ...form, base_url: e.target.value })} placeholder="https://example.com" />
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
    </div>
  )
}

export default Projects
