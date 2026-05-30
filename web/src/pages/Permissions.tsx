import { useState, useEffect, useCallback } from 'react'
import { Card, Button, Checkbox, Empty, Divider, Spin } from 'antd'
import toast from 'react-hot-toast'
import { Client, Tool, Project, getClients, getTools, getProjects, getClientPermissions, updateClientPermissions } from '../api'

function Permissions() {
  const [clients, setClients] = useState<Client[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedClient, setSelectedClient] = useState<string>('')
  const [draft, setDraft] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [clientsData, toolsData, projectsData] = await Promise.all([
        getClients(), getTools(), getProjects(),
      ])
      setClients(clientsData)
      setTools(toolsData)
      setProjects(projectsData)
      // 自动选中第一个客户端
      if (clientsData.length > 0 && !selectedClient) {
        setSelectedClient(clientsData[0].client_id)
      }
    } catch (e: unknown) {
      toast.error('加载数据失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [selectedClient])

  useEffect(() => {
    loadData()
  }, [loadData])

  // 切换客户端时加载其权限
  useEffect(() => {
    if (!selectedClient) return
    const loadPermissions = async () => {
      try {
        const toolIds = await getClientPermissions(selectedClient)
        setDraft(toolIds)
      } catch {
        setDraft([])
      }
    }
    loadPermissions()
  }, [selectedClient])

  const selectClient = (clientId: string) => {
    setSelectedClient(clientId)
  }

  const toggleTool = (toolId: number) => {
    setDraft(prev => prev.includes(toolId) ? prev.filter(id => id !== toolId) : [...prev, toolId])
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await updateClientPermissions(selectedClient, draft)
      toast.success('权限保存成功')
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const enabledProjects = projects.filter(p => p.status === 1)
  const enabledTools = tools.filter(t => t.status === 1)

  const groupedTools = enabledProjects.map(proj => ({
    ...proj,
    tools: enabledTools.filter(t => t.project_id === proj.project_id),
  })).filter(g => g.tools.length > 0)

  const currentClient = clients.find(c => c.client_id === selectedClient)

  return (
    <div>
      <div className="page-header">
        <h2>权限管理</h2>
        <Button type="primary" onClick={handleSave} loading={saving}>保存权限</Button>
      </div>

      <Spin spinning={loading}>
        <div className="permission-layout">
          <div className="permission-client-list">
            <div style={{ fontWeight: 600, padding: '12px 16px 8px', color: '#555', fontSize: 13 }}>
              选择客户端
            </div>
            {clients.map(c => (
              <div
                key={c.client_id}
                className={`permission-client-item ${selectedClient === c.client_id ? 'selected' : ''}`}
                onClick={() => selectClient(c.client_id)}
              >
                {c.name}
              </div>
            ))}
            {clients.length === 0 && (
              <div style={{ padding: 16, color: '#999', fontSize: 13 }}>暂无客户端</div>
            )}
          </div>

          <div className="permission-tools">
            <div style={{ fontWeight: 600, marginBottom: 16, fontSize: 15 }}>
              当前客户端：{currentClient?.name || '-'}
            </div>

            <Card>
              {groupedTools.map((group, i) => (
                <div key={group.project_id}>
                  {i > 0 && <Divider style={{ margin: '8px 0' }} />}
                  <div className="permission-group">
                    <div className="permission-group-title">📁 {group.name}</div>
                    {group.tools.map(tool => (
                      <label key={tool.tool_id} className="permission-tool-item">
                        <Checkbox
                          checked={draft.includes(tool.tool_id)}
                          onChange={() => toggleTool(tool.tool_id)}
                        />
                        <span style={{ fontWeight: 500 }}>{tool.name}</span>
                        <span style={{ color: '#999', fontSize: 13 }}>{tool.title}</span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
              {groupedTools.length === 0 && <Empty description="暂无启用的工具" />}
            </Card>
          </div>
        </div>
      </Spin>
    </div>
  )
}

export default Permissions
