import { useState } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, TeamOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'

interface Client {
  client_id: string
  name: string
  api_key_prefix: string
  description: string
  status: number
  tool_count: number
}

const mockClients: Client[] = [
  { client_id: 'cli_bigdata', name: '大数据项目组', api_key_prefix: 'sk-bigdata', description: '大数据组，用于用户数据查询', status: 1, tool_count: 2 },
  { client_id: 'cli_payment', name: '支付项目组', api_key_prefix: 'sk-payment', description: '支付组，可创建帖子数据', status: 1, tool_count: 1 },
]

function Clients() {
  const [clients, setClients] = useState<Client[]>(mockClients)
  const [openForm, setOpenForm] = useState(false)
  const [openKeyModal, setOpenKeyModal] = useState(false)
  const [newApiKey, setNewApiKey] = useState('')
  const [editing, setEditing] = useState<Client | null>(null)
  const [form] = Form.useForm()

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
            <span style={{ fontWeight: 600, color: '#1677ff' }}>{v}</span>
            <span style={{ color: '#999' }}>-********</span>
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
            <Button size="small" type="primary" icon={<KeyOutlined />} onClick={() => handleGenerateKey(record)}>
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
    setOpenForm(true)
  }

  const openEdit = (c: Client) => {
    setEditing(c)
    form.setFieldsValue(c)
    setOpenForm(true)
  }

  const handleGenerateKey = (client: Client) => {
    const random = Array.from({ length: 20 }, () => Math.random().toString(36)[2]).join('')
    setNewApiKey(`${client.api_key_prefix}-${random}`)
    setOpenKeyModal(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    if (editing) {
      setClients(clients.map(c => c.client_id === editing.client_id ? { ...c, ...values } : c))
      toast.success('客户端更新成功')
    } else {
      setClients([...clients, { ...values, api_key_prefix: `sk-${values.client_id.replace('cli_', '')}`, tool_count: 0 }])
      toast.success('客户端创建成功')
    }
    setOpenForm(false)
  }

  const handleDelete = (id: string) => {
    setClients(clients.filter(c => c.client_id !== id))
    toast.success('客户端已删除')
  }

  return (
    <div>
      <div className="page-header">
        <h2>客户端管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建客户端</Button>
      </div>

      <Card styles={{ body: { padding: 0 } }}>
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
      </Card>

      <Modal
        title={editing ? '编辑客户端' : '新建客户端'}
        open={openForm}
        onOk={handleSave}
        onCancel={() => setOpenForm(false)}
        okText="保存"
        cancelText="取消"
        destroyOnClose
        width={480}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="client_id" label="Client ID" rules={[{ required: true }]}>
            <Input disabled={!!editing} placeholder="如 cli_bigdata" />
          </Form.Item>
          <Form.Item name="name" label="客户端名称" rules={[{ required: true }]}>
            <Input placeholder="如 大数据项目组" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="客户端用途说明" />
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
