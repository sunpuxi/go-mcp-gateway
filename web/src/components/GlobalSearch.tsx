import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { Input, List, Tag, Typography, Empty } from 'antd'
import { useNavigate } from 'react-router-dom'
import {
  FolderOutlined, ToolOutlined, KeyOutlined, SearchOutlined,
  EnterOutlined, ArrowUpOutlined, ArrowDownOutlined, CloseOutlined,
} from '@ant-design/icons'
import { getProjects, getTools, getClients, type Project, type Tool, type Client } from '../api'

const { Text } = Typography

interface SearchResult {
  id: string
  type: 'project' | 'tool' | 'client'
  title: string
  subtitle: string
  path: string
  tags: string[]
}

interface GlobalSearchProps {
  open: boolean
  onClose: () => void
}

function GlobalSearch({ open, onClose }: GlobalSearchProps) {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [clients, setClients] = useState<Client[]>([])
  const [activeIndex, setActiveIndex] = useState(0)
  const inputRef = useRef<any>(null)

  // 打开时加载数据
  useEffect(() => {
    if (open) {
      setQuery('')
      setActiveIndex(0)
      Promise.all([
        getProjects().catch(() => []),
        getTools().catch(() => []),
        getClients().catch(() => []),
      ]).then(([p, t, c]) => {
        setProjects(Array.isArray(p) ? p : [])
        setTools(Array.isArray(t) ? t : [])
        setClients(Array.isArray(c) ? c : [])
      })
      // 聚焦输入框
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  // 搜索结果
  const results = useMemo<SearchResult[]>(() => {
    if (!query.trim()) return []

    const q = query.toLowerCase()
    const items: SearchResult[] = []

    projects
      .filter(p => p.name.toLowerCase().includes(q) || p.project_id.toLowerCase().includes(q))
      .slice(0, 3)
      .forEach(p => {
        items.push({
          id: p.project_id,
          type: 'project',
          title: p.name,
          subtitle: p.project_id,
          path: '/projects',
          tags: ['项目', p.status === 1 ? '启用' : '禁用'],
        })
      })

    tools
      .filter(t => t.name.toLowerCase().includes(q) || t.title.toLowerCase().includes(q))
      .slice(0, 5)
      .forEach(t => {
        items.push({
          id: `tool_${t.tool_id}`,
          type: 'tool',
          title: t.title || t.name,
          subtitle: `${t.http_method} ${t.url_template}`,
          path: '/tools',
          tags: ['工具', t.http_method, t.status === 1 ? '启用' : '禁用'],
        })
      })

    clients
      .filter(c => c.name.toLowerCase().includes(q) || c.client_id.toLowerCase().includes(q))
      .slice(0, 3)
      .forEach(c => {
        items.push({
          id: c.client_id,
          type: 'client',
          title: c.name,
          subtitle: c.client_id,
          path: '/clients',
          tags: ['客户端', c.status === 1 ? '启用' : '禁用'],
        })
      })

    return items
  }, [query, projects, tools, clients])

  // 键盘导航
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActiveIndex(i => Math.min(i + 1, results.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActiveIndex(i => Math.max(i - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (results[activeIndex]) {
        handleSelect(results[activeIndex])
      }
    } else if (e.key === 'Escape') {
      onClose()
    }
  }, [results, activeIndex, onClose])

  const handleSelect = (item: SearchResult) => {
    navigate(item.path)
    onClose()
  }

  if (!open) return null

  const typeConfig = {
    project: { icon: <FolderOutlined />, color: '#faad14', label: '项目' },
    tool: { icon: <ToolOutlined />, color: '#1677ff', label: '工具' },
    client: { icon: <KeyOutlined />, color: '#722ed1', label: '客户端' },
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        background: 'rgba(0,0,0,0.45)',
        display: 'flex',
        justifyContent: 'center',
        paddingTop: '12vh',
      }}
      onClick={onClose}
    >
      <div
        style={{
          width: 560,
          maxWidth: '90vw',
          background: '#fff',
          borderRadius: 12,
          boxShadow: '0 16px 48px rgba(0,0,0,0.18)',
          overflow: 'hidden',
          maxHeight: '70vh',
          display: 'flex',
          flexDirection: 'column',
        }}
        onClick={e => e.stopPropagation()}
      >
        {/* 搜索栏 */}
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #f0f0f0' }}>
          <Input
            ref={inputRef}
            placeholder="搜索项目、工具、客户端…"
            prefix={<SearchOutlined style={{ color: '#bfbfbf' }} />}
            suffix={
              <Text type="secondary" style={{ fontSize: 11 }}>
                <Tag style={{ fontSize: 10, lineHeight: '18px' }}><CloseOutlined /> 关闭</Tag>
              </Text>
            }
            value={query}
            onChange={e => { setQuery(e.target.value); setActiveIndex(0) }}
            onKeyDown={handleKeyDown}
            size="large"
            variant="borderless"
            autoFocus
          />
        </div>

        {/* 结果列表 */}
        <div style={{ flex: 1, overflowY: 'auto' }}>
          {query.trim() === '' ? (
            <div style={{ padding: '32px 20px', textAlign: 'center' }}>
              <Text type="secondary">
                输入关键字搜索项目、工具或客户端
              </Text>
              <div style={{ marginTop: 16, display: 'flex', gap: 16, justifyContent: 'center' }}>
                <Tag color="gold"><ArrowUpOutlined /> <ArrowDownOutlined /> 导航</Tag>
                <Tag color="blue"><EnterOutlined /> 选择</Tag>
                <Tag><CloseOutlined /> 关闭</Tag>
              </div>
            </div>
          ) : results.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="未找到匹配结果"
              style={{ padding: 32 }}
            />
          ) : (
            <List
              dataSource={results}
              renderItem={(item, index) => {
                const cfg = typeConfig[item.type]
                return (
                  <div
                    key={item.id}
                    style={{
                      padding: '12px 20px',
                      cursor: 'pointer',
                      background: index === activeIndex ? '#f0f5ff' : 'transparent',
                      borderLeft: index === activeIndex ? '3px solid #1677ff' : '3px solid transparent',
                      transition: 'background 0.1s',
                    }}
                    onClick={() => handleSelect(item)}
                    onMouseEnter={() => setActiveIndex(index)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span style={{ color: cfg.color, fontSize: 16 }}>{cfg.icon}</span>
                      <div style={{ flex: 1 }}>
                        <Text strong style={{ fontSize: 14 }}>{item.title}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{item.subtitle}</Text>
                      </div>
                      <div style={{ display: 'flex', gap: 4 }}>
                        <Tag color={cfg.color === '#faad14' ? 'gold' : cfg.color === '#1677ff' ? 'blue' : 'purple'} style={{ fontSize: 11 }}>
                          {cfg.label}
                        </Tag>
                        {item.tags.slice(1).map(tag => (
                          <Tag key={tag} style={{ fontSize: 11 }}>{tag}</Tag>
                        ))}
                      </div>
                    </div>
                  </div>
                )
              }}
            />
          )}
        </div>

        {/* 底部提示 */}
        <div style={{
          padding: '8px 20px',
          borderTop: '1px solid #f0f0f0',
          display: 'flex',
          justifyContent: 'space-between',
        }}>
          <Text type="secondary" style={{ fontSize: 11 }}>
            共 {results.length} 条结果
          </Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            <kbd>↑↓</kbd> 导航 <kbd>Enter</kbd> 选择 <kbd>Esc</kbd> 关闭
          </Text>
        </div>
      </div>
    </div>
  )
}

export default GlobalSearch
