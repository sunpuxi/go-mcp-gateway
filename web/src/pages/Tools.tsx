import { useState, useEffect, useCallback, useMemo } from 'react'
import { Table, Button, Modal, Form, Input, Select, InputNumber, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin, Collapse, Dropdown } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ApiOutlined, ReloadOutlined, SearchOutlined, DownloadOutlined, FilterOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Tool, ToolForm, ParamRule, RetryConfig, RateLimitConfig, Project, getTools, createTool, updateTool, deleteTool, getProjects } from '../api'
import { appendOperationLog } from '../utils/operationLog'
import { exportToCSV } from '../utils/export'

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
  const [togglingIds, setTogglingIds] = useState<Set<number>>(new Set())
  const [form] = Form.useForm()

  // 筛选状态
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [methodFilter, setMethodFilter] = useState<string>('all')
  const [projectFilter, setProjectFilter] = useState<string>('all')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

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

  // 客户端筛选
  const filteredData = useMemo(() => {
    return tools.filter(t => {
      const matchSearch = !searchText ||
        t.name.toLowerCase().includes(searchText.toLowerCase()) ||
        t.title.toLowerCase().includes(searchText.toLowerCase()) ||
        (t.description && t.description.toLowerCase().includes(searchText.toLowerCase()))
      const matchStatus = statusFilter === 'all' ||
        (statusFilter === 'enabled' && t.status === 1) ||
        (statusFilter === 'disabled' && t.status === 0)
      const matchMethod = methodFilter === 'all' || t.http_method === methodFilter
      const matchProject = projectFilter === 'all' || t.project_id === projectFilter
      return matchSearch && matchStatus && matchMethod && matchProject
    })
  }, [tools, searchText, statusFilter, methodFilter, projectFilter])

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
      responsive: ['md' as const],
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
      responsive: ['lg' as const],
      render: (v: string) => <code style={{ fontSize: 13, background: '#f5f5f5', padding: '2px 6px', borderRadius: 4 }}>{v}</code>,
    },
    {
      title: '参数', key: 'params', width: 70, align: 'center' as const,
      render: (_: unknown, r: Tool) => (
        <Tag color={r.params && r.params.length > 0 ? 'blue' : 'default'}>{r.params ? r.params.length : 0} 个</Tag>
      ),
    },
    {
      title: '限流', key: 'rate_limit', width: 80, align: 'center' as const,
      render: (_: unknown, r: Tool) => {
        const cfg = r.rate_limit_config
        if (cfg && cfg.max_requests > 0) {
          return (
            <Tooltip title={`${cfg.max_requests} 次 / ${cfg.window_seconds}s`}>
              <Tag color="orange">{cfg.max_requests}/{cfg.window_seconds}s</Tag>
            </Tooltip>
          )
        }
        return <Tag color="default">不限</Tag>
      },
    },
    {
      title: '超时', dataIndex: 'timeout_ms', key: 'timeout_ms', width: 90, align: 'center' as const,
      responsive: ['sm' as const],
      render: (v: number) => <span style={{ color: '#666' }}>{v}ms</span>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80, align: 'center' as const,
      render: (s: number, record: Tool) => (
        <Switch
          checked={s === 1}
          loading={togglingIds.has(record.tool_id)}
          onChange={(checked) => handleToggleStatus(record.tool_id, checked)}
        />
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
      retry_enabled: false, max_retries: 3, backoff_type: 'exponential',
      retry_on_status: [502, 503, 504], retry_on_methods: ['GET'],
      rate_limit_enabled: false, max_requests: 100, window_seconds: 1,
    })
    setOpen(true)
  }

  const openEdit = (t: Tool) => {
    setEditing(t)
    const rc = t.retry_config
    const rlc = t.rate_limit_config
    form.setFieldsValue({
      ...t, status: t.status === 1,
      retry_enabled: !!(rc && rc.max_retries > 0),
      max_retries: rc?.max_retries ?? 3,
      backoff_type: rc?.backoff_type ?? 'exponential',
      retry_on_status: rc?.retry_on_status ?? [502, 503, 504],
      retry_on_methods: rc?.retry_on_methods ?? ['GET'],
      rate_limit_enabled: !!(rlc && rlc.max_requests > 0),
      max_requests: rlc?.max_requests ?? 100,
      window_seconds: rlc?.window_seconds ?? 1,
    })
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
        retry_config: values.retry_enabled ? {
          max_retries: values.max_retries ?? 3,
          backoff_type: values.backoff_type ?? 'exponential',
          retry_on_status: values.retry_on_status ?? [502, 503, 504],
          retry_on_methods: values.retry_on_methods ?? ['GET'],
        } : null,
        rate_limit_config: values.rate_limit_enabled ? {
          max_requests: values.max_requests ?? 100,
          window_seconds: values.window_seconds ?? 1,
        } : null,
        status: values.status ? 1 : 0,
      }
      if (editing) {
        await updateTool(editing.tool_id, payload)
        toast.success('工具更新成功')
        appendOperationLog('编辑', payload.name, `更新工具 #${editing.tool_id}`)
      } else {
        await createTool(payload)
        toast.success('工具创建成功')
        appendOperationLog('新增', payload.name, `创建工具 ${payload.project_id}/${payload.name}`)
      }
      setOpen(false)
      await loadData()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleToggleStatus = async (id: number, checked: boolean) => {
    const tool = tools.find(t => t.tool_id === id)
    if (!tool) return

    setTogglingIds(prev => new Set(prev).add(id))
    try {
      const payload: ToolForm = {
        project_id: tool.project_id,
        name: tool.name,
        title: tool.title,
        description: tool.description,
        http_method: tool.http_method,
        url_template: tool.url_template,
        timeout_ms: tool.timeout_ms,
        params: tool.params || [],
        retry_config: tool.retry_config,
        rate_limit_config: tool.rate_limit_config,
        status: checked ? 1 : 0,
      }
      await updateTool(id, payload)
      toast.success(checked ? '工具已启用' : '工具已禁用')
      appendOperationLog(checked ? '启用' : '停用', tool.name, `切换工具 #${id} 状态`)
      await loadData()
    } catch (e: unknown) {
      toast.error('操作失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setTogglingIds(prev => {
        const next = new Set(prev)
        next.delete(id)
        return next
      })
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteTool(id)
      toast.success('工具已删除')
      appendOperationLog('删除', `#${id}`, '删除工具定义')
      await loadData()
    } catch (e: unknown) {
      toast.error('删除失败: ' + (e instanceof Error ? e.message : String(e)))
    }
  }

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) return
    setLoading(true)
    let successCount = 0
    for (const id of selectedRowKeys) {
      try {
        await deleteTool(id as number)
        appendOperationLog('删除', `#${id}`, '批量删除工具')
        successCount++
      } catch {
        toast.error(`删除工具 #${id} 失败`)
      }
    }
    setSelectedRowKeys([])
    setLoading(false)
    if (successCount > 0) toast.success(`成功删除 ${successCount} 个工具`)
    await loadData()
  }

  const handleExport = (format: 'csv' | 'json') => {
    const exportColumns = [
      { title: '名称', dataIndex: 'name' },
      { title: '标题', dataIndex: 'title' },
      { title: '所属项目', dataIndex: 'project_id' },
      { title: 'HTTP方法', dataIndex: 'http_method' },
      { title: 'URL模板', dataIndex: 'url_template' },
      { title: '超时', dataIndex: 'timeout_ms' },
      { title: '状态', dataIndex: 'status' },
    ]
    if (format === 'csv') {
      exportToCSV(filteredData, exportColumns, 'tools')
    } else {
      import('../utils/export').then(({ exportToJSON }) => {
        exportToJSON(filteredData, 'tools')
      })
    }
    toast.success(`已导出 ${filteredData.length} 条记录`)
  }

  return (
    <div>
      <div className="page-header">
        <h2>工具管理</h2>
        <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>刷新</Button>
          <Dropdown menu={{
            items: [
              { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
              { key: 'json', label: '导出 JSON', onClick: () => handleExport('json') },
            ],
          }}>
            <Button icon={<DownloadOutlined />}>导出</Button>
          </Dropdown>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建工具</Button>
        </Space>
      </div>

      {/* 筛选栏 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索工具名称 / 标题…"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={e => setSearchText(e.target.value)}
            style={{ width: 240 }}
            allowClear
          />
          <Select
            value={projectFilter}
            onChange={v => setProjectFilter(v)}
            style={{ width: 140 }}
            placeholder="所属项目"
            options={[
              { value: 'all', label: '全部项目' },
              ...projectOptions.map(p => ({ value: p.value, label: p.label })),
            ]}
            prefix={<FilterOutlined />}
          />
          <Select
            value={methodFilter}
            onChange={v => setMethodFilter(v)}
            style={{ width: 100 }}
            options={[
              { value: 'all', label: '全部方法' },
              ...['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => ({ value: m, label: m })),
            ]}
          />
          <Select
            value={statusFilter}
            onChange={v => setStatusFilter(v)}
            style={{ width: 110 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'enabled', label: '已启用' },
              { value: 'disabled', label: '已禁用' },
            ]}
          />
          {selectedRowKeys.length > 0 && (
            <Popconfirm
              title="批量删除"
              description={`确定删除选中的 ${selectedRowKeys.length} 个工具？`}
              onConfirm={handleBatchDelete}
              okText="确认删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
            >
              <Button danger icon={<DeleteOutlined />}>
                删除选中 ({selectedRowKeys.length})
              </Button>
            </Popconfirm>
          )}
        </Space>
      </Card>

      <Card styles={{ body: { padding: 0 } }}>
        <Spin spinning={loading}>
          <Table
            rowKey="tool_id"
            columns={columns}
            dataSource={filteredData}
            size="middle"
            scroll={{ x: 1200 }}
            rowSelection={{
              selectedRowKeys,
              onChange: setSelectedRowKeys,
            }}
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
        styles={{ body: { maxHeight: '65vh', overflowY: 'auto' } }}
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

          <Collapse
            ghost
            items={[{
              key: 'retry',
              label: '重试策略（可选）',
              children: (
                <>
                  <Form.Item name="retry_enabled" label="启用重试" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item noStyle shouldUpdate={(prev, cur) => prev.retry_enabled !== cur.retry_enabled}>
                    {({ getFieldValue }) => {
                      const enabled = getFieldValue('retry_enabled')
                      if (!enabled) return null
                      return (
                        <>
                          <Space.Compact block>
                            <Form.Item name="max_retries" label="最大重试次数" style={{ flex: 1 }}>
                              <InputNumber min={1} max={10} style={{ width: '100%' }} />
                            </Form.Item>
                            <Form.Item name="backoff_type" label="退避策略" style={{ flex: 1 }}>
                              <Select options={[
                                { value: 'exponential', label: '指数退避 (1s → 2s → 4s...)' },
                                { value: 'fixed', label: '固定间隔 (每次 1s)' },
                              ]} />
                            </Form.Item>
                          </Space.Compact>
                          <Form.Item name="retry_on_status" label="触发重试的 HTTP 状态码">
                            <Select mode="tags" placeholder="输入状态码后回车添加"
                              options={[502, 503, 504, 500, 408, 429].map(s => ({ value: s, label: String(s) }))}
                            />
                          </Form.Item>
                          <Form.Item name="retry_on_methods" label="允许重试的 HTTP 方法"
                            help="默认仅 GET，POST/PUT/DELETE 需显式添加">
                            <Select mode="tags" placeholder="添加方法"
                              options={['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => ({ value: m, label: m }))}
                            />
                          </Form.Item>
                        </>
                      )
                    }}
                  </Form.Item>
                </>
              ),
            }, {
              key: 'rate_limit',
              label: '限流策略（可选）',
              children: (
                <>
                  <Form.Item name="rate_limit_enabled" label="启用限流" valuePropName="checked"
                    help="基于滑动窗口算法，限制单个工具在时间窗口内的最大请求数">
                    <Switch />
                  </Form.Item>
                  <Form.Item noStyle shouldUpdate={(prev, cur) => prev.rate_limit_enabled !== cur.rate_limit_enabled}>
                    {({ getFieldValue }) => {
                      const enabled = getFieldValue('rate_limit_enabled')
                      if (!enabled) return null
                      return (
                        <Space.Compact block>
                          <Form.Item name="max_requests" label="最大请求数" style={{ flex: 1 }}
                            help="窗口内允许的最大请求数">
                            <InputNumber min={1} max={10000} style={{ width: '100%' }} />
                          </Form.Item>
                          <Form.Item name="window_seconds" label="窗口大小(秒)" style={{ flex: 1 }}
                            help="滑动窗口的时间长度">
                            <InputNumber min={1} max={3600} style={{ width: '100%' }} />
                          </Form.Item>
                        </Space.Compact>
                      )
                    }}
                  </Form.Item>
                </>
              ),
            }]}
          />
        </Form>
      </Modal>
    </div>
  )
}

export default Tools
