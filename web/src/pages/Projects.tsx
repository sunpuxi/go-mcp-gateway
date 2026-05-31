import { useState, useEffect, useCallback, useMemo } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin, Select, Dropdown } from 'antd'
import {
  PlusOutlined, EditOutlined, DeleteOutlined, FolderOutlined, ReloadOutlined,
  SearchOutlined, DownloadOutlined, FilterOutlined,
} from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Project, ProjectForm, getProjects, createProject, updateProject, deleteProject } from '../api'
import { appendOperationLog } from '../utils/operationLog'
import { exportToCSV } from '../utils/export'

function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Project | null>(null)
  const [saving, setSaving] = useState(false)
  const [togglingIds, setTogglingIds] = useState<Set<string>>(new Set())
  const [form] = Form.useForm()

  // 筛选状态
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  const loadProjects = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getProjects()
      setProjects(data)
    } catch (e: unknown) {
      toast.error('加载项目列表失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadProjects()
  }, [loadProjects])

  // 客户端筛选
  const filteredData = useMemo(() => {
    return projects.filter(p => {
      const matchSearch = !searchText ||
        p.name.toLowerCase().includes(searchText.toLowerCase()) ||
        p.project_id.toLowerCase().includes(searchText.toLowerCase()) ||
        (p.description && p.description.toLowerCase().includes(searchText.toLowerCase()))
      const matchStatus = statusFilter === 'all' ||
        (statusFilter === 'enabled' && p.status === 1) ||
        (statusFilter === 'disabled' && p.status === 0)
      return matchSearch && matchStatus
    })
  }, [projects, searchText, statusFilter])

  const columns = [
    {
      title: 'Project ID', dataIndex: 'project_id', key: 'project_id', width: 140,
      render: (v: string) => <Tag color="purple">{v}</Tag>,
    },
    {
      title: '项目名称', dataIndex: 'name', key: 'name', width: 160,
      render: (v: string) => <Space><FolderOutlined style={{ color: '#faad14' }} /><strong>{v}</strong></Space>,
    },
    {
      title: '基础 URL', dataIndex: 'base_url', key: 'base_url',
      ellipsis: { showTitle: false },
      responsive: ['md' as const],
      render: (v: string) => (
        <Tooltip title={v}>
          <code style={{ fontSize: 13 }}>{v}</code>
        </Tooltip>
      ),
    },
    {
      title: '描述', dataIndex: 'description', key: 'description', width: 200,
      ellipsis: { showTitle: false },
      responsive: ['lg' as const],
      render: (v: string) => (
        <Tooltip title={v}>
          <span style={{ color: '#666' }}>{v || '-'}</span>
        </Tooltip>
      ),
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100, align: 'center' as const,
      render: (s: number, record: Project) => (
        <Switch
          checked={s === 1}
          loading={togglingIds.has(record.project_id)}
          onChange={(checked) => handleToggleStatus(record.project_id, checked)}
        />
      ),
    },
    {
      title: '操作', key: 'action', width: 140, align: 'center' as const,
      render: (_: unknown, record: Project) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button type="primary" size="small" ghost icon={<EditOutlined />} onClick={() => openEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="确定删除？"
            description={`将删除项目「${record.name}」及其所有工具和授权信息`}
            onConfirm={() => handleDelete(record.project_id)}
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
    form.resetFields()
    form.setFieldsValue({ status: true })
    setOpen(true)
  }

  const openEdit = (p: Project) => {
    setEditing(p)
    form.setFieldsValue({ ...p, status: p.status === 1 })
    setOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    setSaving(true)
    try {
      const payload: ProjectForm = {
        project_id: editing?.project_id || values.project_id,
        name: values.name,
        base_url: values.base_url,
        description: values.description || '',
        status: values.status ? 1 : 0,
      }
      if (editing) {
        await updateProject(editing.project_id, payload)
        toast.success('项目更新成功')
        appendOperationLog('编辑', payload.name, `更新项目 ${editing.project_id}`)
      } else {
        await createProject(payload)
        toast.success('项目创建成功')
        appendOperationLog('新增', payload.name, `创建项目 ${payload.project_id}`)
      }
      setOpen(false)
      await loadProjects()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleToggleStatus = async (id: string, checked: boolean) => {
    setTogglingIds(prev => new Set(prev).add(id))
    try {
      await updateProject(id, { status: checked ? 1 : 0 })
      toast.success(checked ? '项目已启用' : '项目已禁用')
      appendOperationLog(checked ? '启用' : '停用', id, '切换项目状态')
      await loadProjects()
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

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id)
      toast.success('项目已删除')
      appendOperationLog('删除', id, '删除项目及其关联数据')
      await loadProjects()
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
        await deleteProject(id as string)
        appendOperationLog('删除', id as string, '批量删除项目')
        successCount++
      } catch {
        toast.error(`删除 ${id} 失败`)
      }
    }
    setSelectedRowKeys([])
    setLoading(false)
    if (successCount > 0) toast.success(`成功删除 ${successCount} 个项目`)
    await loadProjects()
  }

  // 导出
  const handleExport = (format: 'csv' | 'json') => {
    const exportColumns = [
      { title: 'Project ID', dataIndex: 'project_id' },
      { title: '项目名称', dataIndex: 'name' },
      { title: '基础URL', dataIndex: 'base_url' },
      { title: '描述', dataIndex: 'description' },
      { title: '状态', dataIndex: 'status' },
    ]
    if (format === 'csv') {
      exportToCSV(filteredData, exportColumns, 'projects')
    } else {
      // lazy import
      import('../utils/export').then(({ exportToJSON }) => {
        exportToJSON(filteredData, 'projects')
      })
    }
    toast.success(`已导出 ${filteredData.length} 条记录`)
  }

  return (
    <div>
      <div className="page-header">
        <h2>项目管理</h2>
        <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={loadProjects} loading={loading}>刷新</Button>
          <Dropdown menu={{
            items: [
              { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
              { key: 'json', label: '导出 JSON', onClick: () => handleExport('json') },
            ],
          }}>
            <Button icon={<DownloadOutlined />}>导出</Button>
          </Dropdown>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建项目</Button>
        </Space>
      </div>

      {/* 筛选栏 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索项目名称 / ID / 描述…"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={e => setSearchText(e.target.value)}
            style={{ width: 280 }}
            allowClear
          />
          <Select
            value={statusFilter}
            onChange={v => setStatusFilter(v)}
            style={{ width: 120 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'enabled', label: '已启用' },
              { value: 'disabled', label: '已禁用' },
            ]}
            prefix={<FilterOutlined />}
          />
          {selectedRowKeys.length > 0 && (
            <Popconfirm
              title="批量删除"
              description={`确定删除选中的 ${selectedRowKeys.length} 个项目？`}
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
            rowKey="project_id"
            columns={columns}
            dataSource={filteredData}
            size="middle"
            rowSelection={{
              selectedRowKeys,
              onChange: setSelectedRowKeys,
            }}
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 个项目`,
            }}
            locale={{ emptyText: '暂无项目，点击右上角"新建项目"开始接入下游服务' }}
            rowClassName={(_, index) => index % 2 === 0 ? 'ant-table-row-striped' : ''}
          />
        </Spin>
      </Card>

      <Modal
        title={editing ? '编辑项目' : '新建项目'}
        open={open}
        onOk={handleSave}
        onCancel={() => setOpen(false)}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnClose
        width={520}
        styles={{ body: { maxHeight: '60vh', overflowY: 'auto' } }}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="project_id" label="Project ID" rules={[{ required: true, message: '请输入 Project ID' }]}>
            <Input disabled={!!editing} placeholder="如 proj_user" />
          </Form.Item>
          <Form.Item name="name" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}>
            <Input placeholder="如 用户服务" />
          </Form.Item>
          <Form.Item name="base_url" label="基础 URL" rules={[{ required: true, message: '请输入基础 URL' }]}>
            <Input placeholder="https://api.example.com" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="项目用途说明" />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Projects
