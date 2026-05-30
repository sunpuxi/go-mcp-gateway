import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Space, Popconfirm, Tag, Tooltip, Card, Switch, Spin } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, FolderOutlined, ReloadOutlined } from '@ant-design/icons'
import toast from 'react-hot-toast'
import { Project, ProjectForm, getProjects, createProject, updateProject, deleteProject } from '../api'

function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Project | null>(null)
  const [saving, setSaving] = useState(false)
  const [form] = Form.useForm()

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
      title: '状态', dataIndex: 'status', key: 'status', width: 100, align: 'center' as const,
      render: (s: number) => (
        <Tag color={s === 1 ? 'success' : 'error'} style={{ minWidth: 48, textAlign: 'center' }}>
          {s === 1 ? '启用' : '禁用'}
        </Tag>
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
      } else {
        await createProject(payload)
        toast.success('项目创建成功')
      }
      setOpen(false)
      await loadProjects()
    } catch (e: unknown) {
      toast.error('保存失败: ' + (e instanceof Error ? e.message : String(e)))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id)
      toast.success('项目已删除')
      await loadProjects()
    } catch (e: unknown) {
      toast.error('删除失败: ' + (e instanceof Error ? e.message : String(e)))
    }
  }

  return (
    <div>
      <div className="page-header">
        <h2>项目管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={loadProjects} loading={loading}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建项目</Button>
        </Space>
      </div>

      <Card styles={{ body: { padding: 0 } }}>
        <Spin spinning={loading}>
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
