import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Power, PowerOff, X, FolderOpen } from 'lucide-react'
import { getProjects, createProject, updateProject, deleteProject, type Project, type ProjectForm } from '../api'
import { useToast } from '../components/Toast'
import { Confirm } from '../components/Confirm'
import { TableSkeleton } from '../components/Skeleton'

const emptyForm: ProjectForm = { project_id: '', name: '', base_url: '', description: '', status: 1 }

function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Project | null>(null)
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [form, setForm] = useState({ ...emptyForm })
  const { toast } = useToast()

  const fetchProjects = useCallback(() => {
    setLoading(true)
    setError('')
    getProjects()
      .then(data => { setProjects(data); setLoading(false) })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  useEffect(() => { fetchProjects() }, [fetchProjects])

  const openCreate = () => { setEditing(null); setForm({ ...emptyForm }); setShowForm(true) }
  const openEdit = (p: Project) => { setEditing(p); setForm({ project_id: p.project_id, name: p.name, base_url: p.base_url, description: p.description, status: p.status }); setShowForm(true) }

  const handleSave = async () => {
    if (!form.project_id.trim() || !form.name.trim() || !form.base_url.trim()) {
      toast('warning', '请填写必填字段')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        const { project_id, ...data } = form
        const updated = await updateProject(editing.project_id, data)
        setProjects(projects.map(p => p.project_id === editing.project_id ? updated : p))
        toast('success', '项目已更新')
      } else {
        const created = await createProject(form)
        setProjects([...projects, created])
        toast('success', '项目已创建')
      }
      setShowForm(false)
    } catch (err: any) {
      toast('error', '保存失败: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!confirmDelete) return
    try {
      await deleteProject(confirmDelete)
      setProjects(projects.filter(p => p.project_id !== confirmDelete))
      toast('success', '项目已删除')
    } catch (err: any) {
      toast('error', '删除失败: ' + err.message)
    }
    setConfirmDelete(null)
  }

  const toggleStatus = async (p: Project) => {
    try {
      const updated = await updateProject(p.project_id, { status: p.status === 1 ? 0 : 1 })
      setProjects(projects.map(x => x.project_id === p.project_id ? updated : x))
      toast('success', `项目已${p.status === 1 ? '禁用' : '启用'}`)
    } catch (err: any) {
      toast('error', '操作失败: ' + err.message)
    }
  }

  if (error) {
    return (
      <div>
        <div className="page-header">
          <h1>项目管理</h1>
          <button className="btn btn-outline" onClick={fetchProjects}>重试</button>
        </div>
        <div className="table-card" style={{ padding: '40px 16px', textAlign: 'center', color: '#c62828', fontSize: 14 }}>加载失败：{error}</div>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1>项目管理</h1>
        <button className="btn btn-primary" onClick={openCreate}><Plus size={16} />新建项目</button>
      </div>

      <div className="table-card">
        {loading ? <TableSkeleton rows={5} cols={5} /> : projects.length === 0 ? (
          <div className="empty-state">
            <FolderOpen size={40} /><p>暂无项目</p><p className="empty-hint">点击「新建项目」接入下游 HTTP 服务</p>
          </div>
        ) : (
          <table>
            <thead>
              <tr><th>Project ID</th><th>名称</th><th>基础 URL</th><th>状态</th><th>操作</th></tr>
            </thead>
            <tbody>
              {projects.map(p => (
                <tr key={p.project_id}>
                  <td><strong>{p.project_id}</strong></td>
                  <td>{p.name}</td>
                  <td style={{ maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{p.base_url}</td>
                  <td><span className={`status-badge ${p.status === 1 ? 'active' : 'inactive'}`}>{p.status === 1 ? '启用' : '禁用'}</span></td>
                  <td>
                    <button className="btn-icon" onClick={() => toggleStatus(p)} title="切换状态">
                      {p.status === 1 ? <PowerOff size={15} /> : <Power size={15} />}
                    </button>
                    <button className="btn-icon" onClick={() => openEdit(p)} title="编辑"><Pencil size={15} /></button>
                    <button className="btn-icon" onClick={() => setConfirmDelete(p.project_id)} title="删除"><Trash2 size={15} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showForm && (
        <div className="form-overlay" onClick={() => setShowForm(false)}>
          <div className="form-panel" onClick={e => e.stopPropagation()}>
            <div className="form-panel-header">
              {editing ? '编辑项目' : '新建项目'}
              <button className="btn-icon" onClick={() => setShowForm(false)}><X size={16} /></button>
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
                  <option value={1}>启用</option><option value={0}>禁用</option>
                </select>
              </div>
            </div>
            <div className="form-panel-footer">
              <button className="btn btn-outline" onClick={() => setShowForm(false)}>取消</button>
              <button className="btn btn-primary" onClick={handleSave} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
            </div>
          </div>
        </div>
      )}

      {confirmDelete && (
        <Confirm title="删除项目" message={`确定删除项目「${confirmDelete}」吗？此操作不可撤销。`} onConfirm={handleDelete} onCancel={() => setConfirmDelete(null)} confirmText="删除" danger />
      )}
    </div>
  )
}

export default Projects
