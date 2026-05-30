import { useState } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, FolderOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'

interface Project {
  project_id: string
  name: string
  base_url: string
  description: string
  status: number
}

const mockProjects: Project[] = [
  { project_id: 'proj_user', name: '用户服务', base_url: 'https://jsonplaceholder.typicode.com', description: '用户相关的 HTTP 服务', status: 1 },
  { project_id: 'proj_post', name: '帖子服务', base_url: 'https://jsonplaceholder.typicode.com', description: '帖子相关的 HTTP 服务', status: 1 },
]

function Projects() {
  const [projects, setProjects] = useState<Project[]>(mockProjects)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Project | null>(null)
  const [form] = Form.useForm()

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
      render: (v: string) => (
        <Tooltip title={v}>
          <code style={{ fontSize: 13 }}>{v}</code>
        </Tooltip>
      ),
    },
    {
      title: '描述', dataIndex: 'description', key: 'description', width: 200,
      ellipsis: { showTitle: false },
      render: (v: string) => (
        <Tooltip title={v}>
          <span style={{ color: '#666' }}>{v || '-'}</span>
        </Tooltip>
      ),
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100, align: 'center',
      render: (s: number) => (
        <Tag color={s === 1 ? 'success' : 'error'} style={{ minWidth: 48, textAlign: 'center' }}>
          {s === 1 ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作', key: 'action', width: 140, align: 'center',
      render: (_: unknown, record: Project) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button type="primary" size="small" ghost icon={<EditOutlined />} onClick={() => openEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="确定删除？"
            description={`将删除项目「${record.name}」及其所有工具`}
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
    setOpen(true)
  }

  const openEdit = (p: Project) => {
    setEditing(p)
    form.setFieldsValue(p)
    setOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    if (editing) {
      setProjects(projects.map(p => p.project_id === editing.project_id ? { ...p, ...values } : p))
      toast.success('项目更新成功')
    } else {
      setProjects([...projects, { ...values, status: 1 }])
      toast.success('项目创建成功')
    }
    setOpen(false)
  }

  const handleDelete = (id: string) => {
    setProjects(projects.filter(p => p.project_id !== id))
    toast.success('项目已删除')
  }

  return (
    <div>
      <div className="page-header">
        <h2>项目管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建项目</Button>
      </div>

      <Card styles={{ body: { padding: 0 } }}>
        <Table
          rowKey="project_id"
          columns={columns}
          dataSource={projects}
          size="middle"
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个项目`,
          }}
          locale={{ emptyText: '暂无项目，点击右上角"新建项目"开始接入下游服务' }}
          rowClassName={(_, index) => index % 2 === 0 ? 'ant-table-row-striped' : ''}
        />
      </Card>

      <Modal
        title={editing ? '编辑项目' : '新建项目'}
        open={open}
        onOk={handleSave}
        onCancel={() => setOpen(false)}
        okText="保存"
        cancelText="取消"
        destroyOnClose
        width={520}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="project_id" label="Project ID" rules={[{ required: true }]}>
            <Input disabled={!!editing} placeholder="如 proj_user" />
          </Form.Item>
          <Form.Item name="name" label="项目名称" rules={[{ required: true }]}>
            <Input placeholder="如 用户服务" />
          </Form.Item>
          <Form.Item name="base_url" label="基础 URL" rules={[{ required: true }]}>
            <Input placeholder="https://api.example.com" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="项目用途说明" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Projects
