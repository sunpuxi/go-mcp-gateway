import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Select, InputNumber, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ApiOutlined, ReloadOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Tool, ToolForm, ParamRule, Project, getTools, createTool, updateTool, deleteTool, getProjects } from '../api'

const httpColors: Record<string, string> = {
  GET: 'green', POST: 'blue', PUT: 'orange', DELETE: 'red', PATCH: 'purple',
}

function Tools() {
  const [tools, setTools] = useState<Tool[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Tool | null>(null)
  const [saving, setSaving] = useState(false)
  const [form] = Form.useForm()

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [toolsData, projectsData] = await Promise.all([getTools(), getProjects()])
      setTools(toolsData)
      setProjects(projectsData)
    } catch (e: unknown) {
      toast.error('加载数据失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const projectOptions = projects
    .filter(p => p.status === 1)
    .map(p => ({ value: p.project_id, label: p.name }))

  const getProjectName = (projectId: string) => {
    return projects.find(p => p.project_id === projectId)?.name || projectId
  }

  const columns = [
    {
      title: '工具名称', dataIndex: 'name', key: 'name', width: 140,
      render: (v: string) => (
        <Space>
          <ApiOutlined style={{ color: '#1677ff' }} />
          <strong>{v}</strong>
        </Space>
      ),
    },
    {
      title: '标题', dataIndex: 'title', key: 'title', width: 160,
      ellipsis: { showTitle: false },
      render: (v: string) => <Tooltip title={v}><span>{v}</span></Tooltip>,
    },
    {
      title: '所属项目', dataIndex: 'project_id', key: 'project_id', width: 110,
      render: (v: string) => <Tag color="purple">{getProjectName(v)}</Tag>,
    },
    {
      title: '方法', dataIndex: 'http_method', key: 'http_method', width: 90, align: 'center' as const,
      render: (m: string) => (
        <Tag color={httpColors[m] || 'default'} style={{ fontWeight: 600, minWidth: 48, textAlign: 'center' }}>
          {m}
        </Tag>
      ),
    },
    {
      title: 'URL 模板', dataIndex: 'url_template', key: 'url_template',
      render: (v: string) => <code style={{ fontSize: 13, background: '#f5f5f5', padding: '2px 6px', borderRadius: 4 }}>{v}</code>,
    },
    {
      title: '参数', key: 'params', width: 70, align: 'center' as const,
      render: (_: unknown, r: Tool) => (
        <Tag color={r.params && r.params.length > 0 ? 'blue' : 'default'}>{r.params ? r.params.length : 0} 个</Tag>
      ),
    },
    {
      title: '超时', dataIndex: 'timeout_ms', key: 'timeout_ms', width: 90, align: 'center' as const,
      render: (v: number) => <span style={{ color: '#666' }}>{v}ms</span>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80, align: 'center' as const,
      render: (s: number) => (
        <Tag color={s === 1 ? 'success' : 'error'} style={{ minWidth: 48, textAlign: 'center' }}>
          {s === 1 ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作', key: 'action', width: 100, align: 'center' as const,
      render: (_: unknown, record: Tool) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button type="primary" size="small" ghost icon={<EditOutlined />} onClick={() => openEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="确定删除？"
            description={`将删除工具「${record.name}」及其授权信息`}
            onConfirm={() => handleDelete(record.tool_id)}
            okText="确认"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Tooltip title="删除">
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const openCreate = () => {
    setEditing(null)
    form.setFieldsValue({
      name: '', title: '', description: '', project_id: projectOptions[0]?.value || '',
      http_method: 'GET', url_template: '', timeout_ms: 5000, status: true, params: [],
    })
    setOpen(true)
  }

  const openEdit = (t: Tool) => {
    setEditing(t)
    form.setFieldsValue({ ...t, status: t.status === 1 })
    setOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    setSaving(true)
    try {
      const payload: ToolForm = {
        project_id: values.project_id,
        name: values.name,
        title: values.title,
        description: values.description || '',
        http_method: values.http_method,
        url_template: values.url_template,
        timeout_ms: values.timeout_ms,
        params: values.params || [],
        status: values.status ? 1 : 0,
      }
      if (editing) {
        await updateTool(editing.tool_id, payload)
        toast.success('工具更新成功')
      } else {
        await createTool(payload)
        toast.success('工具创建成功')
      }
      setOpen(false)
      await loadData()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteTool(id)
      toast.success('工具已删除')
      await loadData()
    } catch (e: unknown) {
      toast.error('删除失败: ' + (e instanceof Error ? e.message : String(e)))
    }
  }

  return (
    <div>
      <div className="page-header">
        <h2>工具管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建工具</Button>
        </Space>
      </div>

      <Card styles={{ body: { padding: 0 } }}>
        <Spin spinning={loading}>
          <Table
            rowKey="tool_id"
            columns={columns}
            dataSource={tools}
            size="middle"
            scroll={{ x: 1100 }}
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 个工具`,
            }}
            locale={{ emptyText: '暂无工具，点击右上角"新建工具"开始配置 HTTP 接口的协议转换' }}
            rowClassName={(_, index) => index % 2 === 0 ? 'ant-table-row-striped' : ''}
          />
        </Spin>
      </Card>

      <Modal
        title={editing ? '编辑工具' : '新建工具'}
        open={open}
        width={800}
        onOk={handleSave}
        onCancel={() => setOpen(false)}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Space.Compact block>
            <Form.Item name="name" label="工具名称" rules={[{ required: true, message: '请输入工具名称' }]} style={{ flex: 1 }}>
              <Input placeholder="get_user" disabled={!!editing} />
            </Form.Item>
            <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]} style={{ flex: 1 }}>
              <Input placeholder="获取用户信息" />
            </Form.Item>
          </Space.Compact>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space.Compact block>
            <Form.Item name="project_id" label="所属项目" rules={[{ required: true, message: '请选择项目' }]} style={{ flex: 1 }}>
              <Select options={projectOptions} placeholder="选择项目" />
            </Form.Item>
            <Form.Item name="http_method" label="HTTP 方法" rules={[{ required: true }]} style={{ flex: 1 }}>
              <Select options={['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => ({ value: m, label: m }))} />
            </Form.Item>
            <Form.Item name="timeout_ms" label="超时(ms)" style={{ flex: 1 }}>
              <InputNumber min={100} max={60000} style={{ width: '100%' }} />
            </Form.Item>
          </Space.Compact>
          <Form.Item name="url_template" label="URL 模板" rules={[{ required: true, message: '请输入 URL 模板' }]}>
            <Input placeholder="/users/{user_id}" />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.List name="params">
            {(fields, { add, remove }) => (
              <div style={{ marginTop: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <span style={{ fontWeight: 600, fontSize: 14 }}>参数映射规则</span>
                  <Button size="small" onClick={() => add({ name: '', type: 'string', location: 'query', required: false, default_value: '', description: '' })}>
                    + 添加参数
                  </Button>
                </div>
                {fields.length === 0 ? (
                  <div style={{ color: '#999', fontSize: 13, padding: 20, textAlign: 'center', background: '#fafafa', borderRadius: 6 }}>
                    暂无参数规则，点击「添加参数」配置 MCP 参数到 HTTP 请求的映射
                  </div>
                ) : (
                  <Table
                    className="param-table"
                    dataSource={fields.map(f => ({ ...f, key: f.key }))}
                    pagination={false}
                    size="small"
                    columns={[
                      {
                        title: '参数名', render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'name']} noStyle rules={[{ required: true, message: '必填' }]}>
                            <Input placeholder="参数名" size="small" />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '类型', width: 100, render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'type']} noStyle>
                            <Select size="small" options={['string', 'number', 'boolean', 'object'].map(v => ({ value: v, label: v }))} style={{ width: 85 }} />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '位置', width: 100, render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'location']} noStyle rules={[{ required: true, message: '必填' }]}>
                            <Select size="small" options={['path', 'query', 'body', 'header'].map(v => ({ value: v, label: v }))} style={{ width: 85 }} />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '必填', width: 60, align: 'center' as const, render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'required']} noStyle valuePropName="checked">
                            <Select size="small" options={[{ value: true, label: '是' }, { value: false, label: '否' }]} style={{ width: 55 }} />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '默认值', width: 110, render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'default_value']} noStyle>
                            <Input placeholder="默认值" size="small" />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '描述', render: (_: unknown, f: { key: number, name: number }) => (
                          <Form.Item name={[f.name, 'description']} noStyle>
                            <Input placeholder="参数说明" size="small" />
                          </Form.Item>
                        ),
                      },
                      {
                        title: '', width: 40,
                        render: (_: unknown, f: { key: number, name: number }) => (
                          <Button type="link" danger size="small" icon={<DeleteOutlined />} onClick={() => remove(f.name)} />
                        ),
                      },
                    ]}
                  />
                )}
              </div>
            )}
          </Form.List>
        </Form>
      </Modal>
    </div>
  )
}

export default Tools
