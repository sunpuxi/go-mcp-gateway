import { useState } from 'react'
import { Card, Button, Checkbox, Empty, Divider } from 'antd'
import toast from 'react-hot-toast'

interface Client {
  client_id: string
  name: string
}

interface Tool {
  tool_id: number
  name: string
  title: string
  project_id: string
}

interface Project {
  project_id: string
  name: string
}

const mockClients: Client[] = [
  { client_id: 'cli_bigdata', name: '大数据项目组' },
  { client_id: 'cli_payment', name: '支付项目组' },
]

const mockProjects: Project[] = [
  { project_id: 'proj_user', name: '用户服务' },
  { project_id: 'proj_post', name: '帖子服务' },
]

const mockTools: Tool[] = [
  { tool_id: 1, name: 'get_user', title: '获取用户信息', project_id: 'proj_user' },
  { tool_id: 2, name: 'get_user_posts', title: '获取用户帖子', project_id: 'proj_user' },
  { tool_id: 3, name: 'create_post', title: '创建帖子', project_id: 'proj_post' },
]

const mockPermissions: Record<string, number[]> = {
  cli_bigdata: [1, 2],
  cli_payment: [3],
}

function Permissions() {
  const [selectedClient, setSelectedClient] = useState<string>('cli_bigdata')
  const [permissions, setPermissions] = useState<Record<string, number[]>>(mockPermissions)
  const [draft, setDraft] = useState<number[]>(permissions['cli_bigdata'] || [])

  const selectClient = (clientId: string) => {
    setSelectedClient(clientId)
    setDraft(permissions[clientId] || [])
  }

  const toggleTool = (toolId: number) => {
    setDraft(prev => prev.includes(toolId) ? prev.filter(id => id !== toolId) : [...prev, toolId])
  }

  const handleSave = () => {
    setPermissions({ ...permissions, [selectedClient]: draft })
    toast.success('权限保存成功')
  }

  const groupedTools = mockProjects.map(proj => ({
    ...proj,
    tools: mockTools.filter(t => t.project_id === proj.project_id),
  }))

  const currentClient = mockClients.find(c => c.client_id === selectedClient)

  return (
    <div>
      <div className="page-header">
        <h2>权限管理</h2>
        <Button type="primary" onClick={handleSave}>保存权限</Button>
      </div>

      <div className="permission-layout">
        <div className="permission-client-list">
          <div style={{ fontWeight: 600, padding: '12px 16px 8px', color: '#555', fontSize: 13 }}>
            选择客户端
          </div>
          {mockClients.map(c => (
            <div
              key={c.client_id}
              className={`permission-client-item ${selectedClient === c.client_id ? 'selected' : ''}`}
              onClick={() => selectClient(c.client_id)}
            >
              {c.name}
            </div>
          ))}
        </div>

        <div className="permission-tools">
          <div style={{ fontWeight: 600, marginBottom: 16, fontSize: 15 }}>
            当前客户端：{currentClient?.name}
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
            {mockTools.length === 0 && <Empty description="暂无工具" />}
          </Card>
        </div>
      </div>
    </div>
  )
}

export default Permissions
