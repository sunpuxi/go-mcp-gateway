import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, TeamOutlined, ReloadOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Client, ClientForm, getClients, createClient, updateClient, deleteClient, generateApiKey } from '../api'

function Clients() {
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState(false)
  const [openForm, setOpenForm] = useState(false)
  const [openKeyModal, setOpenKeyModal] = useState(false)
  const [newApiKey, setNewApiKey] = useState('')
  const [editing, setEditing] = useState<Client | null>(null)
  const [saving, setSaving] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [form] = Form.useForm()

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
      render: (s: number) => (
        <Tag color={s === 1 ? 'success' : 'error'} style={{ minWidth: 48, textAlign: 'center' }}>
          {s === 1 ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作', key: 'action', width: 200, align: 'center' as const,
      render: (_: unknown, record: Client) => (
        <Space size="small">
          <Tooltip title="生成新 API Key">
            <Button size="small" type="primary" icon={<KeyOutlined />} loading={generating} onClick={() => handleGenerateKey(record)}>
              生成 Key
            </Button>
          </Tooltip>
          <Tooltip title="编辑">
            <Button size="small" ghost icon={<EditOutlined />} onClick={() => openEdit(record)} />
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
      } else {
        await createClient(payload)
        toast.success('客户端创建成功')
      }
      setOpenForm(false)
      await loadClients()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteClient(id)
      toast.success('客户端已删除')
      await loadClients()
    } catch (e: unknown) {
      toast.error('删除失败: ' + (e instanceof Error ? e.message : String(e)))
    }
  }

  return (
    <div>
      <div className="page-header">
        <h2>客户端管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={loadClients} loading={loading}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建客户端</Button>
        </Space>
      </div>

      <Card styles={{ body: { padding: 0 } }}>
        <Spin spinning={loading}>
          <Table
            rowKey="client_id"
            columns={columns}
            dataSource={clients}
            size="middle"
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
