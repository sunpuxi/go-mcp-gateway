import { useState, useEffect, useCallback, useMemo } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin, Select, Dropdown } from 'antd'
import {
  PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, TeamOutlined, ReloadOutlined,
  SearchOutlined, DownloadOutlined, FilterOutlined,
} from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Client, ClientForm, getClients, createClient, updateClient, deleteClient, generateApiKey } from '../api'
import { appendOperationLog } from '../utils/operationLog'
import { exportToCSV } from '../utils/export'

function Clients() {
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState(false)
  const [openForm, setOpenForm] = useState(false)
  const [openKeyModal, setOpenKeyModal] = useState(false)
  const [newApiKey, setNewApiKey] = useState('')
  const [editing, setEditing] = useState<Client | null>(null)
  const [saving, setSaving] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [togglingIds, setTogglingIds] = useState<Set<string>>(new Set())
  const [form] = Form.useForm()

  // 筛选状态
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  const loadClients = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getClients()
      setClients(data)
    } catch (e: unknown) {
      toast.error('加载客户端列表失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadClients()
  }, [loadClients])

  // 客户端筛选
  const filteredData = useMemo(() => {
    return clients.filter(c => {
      const matchSearch = !searchText ||
        c.name.toLowerCase().includes(searchText.toLowerCase()) ||
        c.client_id.toLowerCase().includes(searchText.toLowerCase()) ||
        (c.description && c.description.toLowerCase().includes(searchText.toLowerCase()))
      const matchStatus = statusFilter === 'all' ||
        (statusFilter === 'enabled' && c.status === 1) ||
        (statusFilter === 'disabled' && c.status === 0)
      return matchSearch && matchStatus
    })
  }, [clients, searchText, statusFilter])

  const columns = [
    {
      title: 'Client ID', dataIndex: 'client_id', key: 'client_id', width: 140,
      render: (v: string) => <Tag color="purple" style={{ fontSize: 13 }}>{v}</Tag>,
    },
    {
      title: '客户端名称', dataIndex: 'name', key: 'name', width: 180,
      render: (v: string) => (
        <Space>
          <TeamOutlined style={{ color: '#1677ff' }} />
          <strong>{v}</strong>
        </Space>
      ),
    },
    {
      title: 'API Key', dataIndex: 'api_key_prefix', key: 'api_key_prefix', width: 180,
      responsive: ['md' as const],
      render: (v: string) => (
        <Tooltip title="完整 Key 仅在生成时展示一次">
          <code style={{ fontSize: 13, background: '#f5f5f5', padding: '3px 8px', borderRadius: 4 }}>
            <span style={{ fontWeight: 600, color: '#1677ff' }}>{v || '未生成'}</span>
            {v && <span style={{ color: '#999' }}>-********</span>}
          </code>
        </Tooltip>
      ),
    },
    {
      title: '描述', dataIndex: 'description', key: 'description', width: 240,
      ellipsis: { showTitle: false },
      responsive: ['lg' as const],
      render: (v: string) => (
        <Tooltip title={v}>
          <span style={{ color: '#666' }}>{v || '-'}</span>
        </Tooltip>
      ),
    },
    {
      title: '已授权工具', dataIndex: 'tool_count', key: 'tool_count', width: 100, align: 'center' as const,
      render: (v: number) => (
        <Tag color={v > 0 ? 'blue' : 'default'} style={{ minWidth: 32, textAlign: 'center' }}>
          {v} 个
        </Tag>
      ),
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80, align: 'center' as const,
      render: (s: number, record: Client) => (
        <Switch
          checked={s === 1}
          loading={togglingIds.has(record.client_id)}
          onChange={(checked) => handleToggleStatus(record.client_id, checked)}
        />
      ),
    },
    {
      title: '操作', key: 'action', width: 200, align: 'center' as const,
      render: (_: unknown, record: Client) => (
        <Space size="small" wrap>
          <Tooltip title="生成新 API Key">
            <Button size="small" type="primary" icon={<KeyOutlined />} loading={generating} onClick={() => handleGenerateKey(record)}>
              Key
            </Button>
          </Tooltip>
          <Tooltip title="编辑">
            <Button type="primary" size="small" ghost icon={<EditOutlined />} onClick={() => openEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="确定删除？"
            description={`将删除客户端「${record.name}」`}
            onConfirm={() => handleDelete(record.client_id)}
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
    setOpenForm(true)
  }

  const openEdit = (c: Client) => {
    setEditing(c)
    form.setFieldsValue({ ...c, status: c.status === 1 })
    setOpenForm(true)
  }

  const handleGenerateKey = async (client: Client) => {
    setGenerating(true)
    try {
      const res = await generateApiKey(client.client_id)
      appendOperationLog('生成密钥', client.name || client.client_id, '生成新 API Key')
      setNewApiKey(res.api_key)
      setOpenKeyModal(true)
      await loadClients()
    } catch (e: unknown) {
      toast.error('生成 Key 失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setGenerating(false)
    }
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    setSaving(true)
    try {
      const payload: ClientForm = {
        client_id: editing?.client_id || values.client_id,
        name: values.name,
        description: values.description || '',
        status: values.status ? 1 : 0,
      }
      if (editing) {
        await updateClient(editing.client_id, payload)
        toast.success('客户端更新成功')
        appendOperationLog('编辑', payload.name, `更新客户端 ${editing.client_id}`)
      } else {
        await createClient(payload)
        toast.success('客户端创建成功')
        appendOperationLog('新增', payload.name, `创建客户端 ${payload.client_id}`)
      }
      setOpenForm(false)
      await loadClients()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleToggleStatus = async (id: string, checked: boolean) => {
    setTogglingIds(prev => new Set(prev).add(id))
    try {
      await updateClient(id, { status: checked ? 1 : 0 })
      toast.success(checked ? '客户端已启用' : '客户端已禁用')
      appendOperationLog(checked ? '启用' : '停用', id, '切换客户端状态')
      await loadClients()
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
      await deleteClient(id)
      toast.success('客户端已删除')
      appendOperationLog('删除', id, '删除客户端及其关联数据')
      await loadClients()
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
        await deleteClient(id as string)
        appendOperationLog('删除', id as string, '批量删除客户端')
        successCount++
      } catch {
        toast.error(`删除 ${id} 失败`)
      }
    }
    setSelectedRowKeys([])
    setLoading(false)
    if (successCount > 0) toast.success(`成功删除 ${successCount} 个客户端`)
    await loadClients()
  }

  const handleExport = (format: 'csv' | 'json') => {
    const exportColumns = [
      { title: 'Client ID', dataIndex: 'client_id' },
      { title: '客户端名称', dataIndex: 'name' },
      { title: 'API Key前缀', dataIndex: 'api_key_prefix' },
      { title: '描述', dataIndex: 'description' },
      { title: '已授权工具', dataIndex: 'tool_count' },
      { title: '状态', dataIndex: 'status' },
    ]
    if (format === 'csv') {
      exportToCSV(filteredData, exportColumns, 'clients')
    } else {
      import('../utils/export').then(({ exportToJSON }) => {
        exportToJSON(filteredData, 'clients')
      })
    }
    toast.success(`已导出 ${filteredData.length} 条记录`)
  }

  return (
    <div>
      <div className="page-header">
        <h2>客户端管理</h2>
        <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={loadClients} loading={loading}>刷新</Button>
          <Dropdown menu={{
            items: [
              { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
              { key: 'json', label: '导出 JSON', onClick: () => handleExport('json') },
            ],
          }}>
            <Button icon={<DownloadOutlined />}>导出</Button>
          </Dropdown>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建客户端</Button>
        </Space>
      </div>

      {/* 筛选栏 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索客户端名称 / ID / 描述…"
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
              description={`确定删除选中的 ${selectedRowKeys.length} 个客户端？`}
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
            rowKey="client_id"
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
              showTotal: (total) => `共 ${total} 个客户端`,
            }}
            locale={{ emptyText: '暂无客户端，点击右上角"新建客户端"创建一个项目组' }}
            rowClassName={(_, index) => index % 2 === 0 ? 'ant-table-row-striped' : ''}
          />
        </Spin>
      </Card>

      <Modal
        title={editing ? '编辑客户端' : '新建客户端'}
        open={openForm}
        onOk={handleSave}
        onCancel={() => setOpenForm(false)}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnClose
        width={480}
        styles={{ body: { maxHeight: '60vh', overflowY: 'auto' } }}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="client_id" label="Client ID" rules={[{ required: true, message: '请输入 Client ID' }]}>
            <Input disabled={!!editing} placeholder="如 cli_bigdata" />
          </Form.Item>
          <Form.Item name="name" label="客户端名称" rules={[{ required: true, message: '请输入客户端名称' }]}>
            <Input placeholder="如 大数据项目组" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="客户端用途说明" />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="API Key 已生成"
        open={openKeyModal}
        onCancel={() => setOpenKeyModal(false)}
        width={480}
        footer={[
          <Button key="copy" type="primary" size="large" block onClick={() => { navigator.clipboard.writeText(newApiKey); toast.success('已复制到剪贴板') }}>
            复制 Key 并关闭
          </Button>,
        ]}
      >
        <div style={{ padding: '8px 0' }}>
          <div className="api-key-display">{newApiKey}</div>
          <div className="api-key-warning">⚠ 请立即复制保存，关闭后将无法再次查看</div>
        </div>
      </Modal>
    </div>
  )
}

export default Clients
