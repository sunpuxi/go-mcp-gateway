/// <reference types="vite/client" />
import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Button, Result, Typography, Collapse } from 'antd'
import { ReloadOutlined, HomeOutlined } from '@ant-design/icons'
import { Navigate } from 'react-router-dom'

const { Paragraph, Text } = Typography

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  redirectHome: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null, redirectHome: false }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('[ErrorBoundary]', error, errorInfo)
    this.setState({ errorInfo })
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  handleGoHome = () => {
    this.setState({ redirectHome: true })
  }

  render() {
    if (this.state.redirectHome) {
      return <Navigate to="/" replace />
    }

    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: 400,
          padding: 24,
        }}>
          <Result
            status="error"
            title="页面出现异常"
            subTitle="抱歉，页面遇到了一个意外错误。您可以尝试重试或返回首页。"
            extra={[
              <Button
                key="retry"
                type="primary"
                icon={<ReloadOutlined />}
                onClick={this.handleRetry}
              >
                重试
              </Button>,
              <Button
                key="home"
                icon={<HomeOutlined />}
                onClick={this.handleGoHome}
              >
                返回首页
              </Button>,
            ]}
          >
            {import.meta.env.DEV && this.state.error && (
              <Collapse
                size="small"
                style={{ maxWidth: 640, margin: '0 auto', textAlign: 'left' }}
                items={[
                  {
                    key: 'details',
                    label: '错误详情（开发模式）',
                    children: (
                      <div>
                        <Paragraph>
                          <Text type="danger" code>{this.state.error.message}</Text>
                        </Paragraph>
                        {this.state.errorInfo && (
                          <pre style={{
                            fontSize: 11,
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-all',
                            maxHeight: 200,
                            overflow: 'auto',
                            background: '#f5f5f5',
                            padding: 8,
                            borderRadius: 4,
                          }}>
                            {this.state.errorInfo.componentStack}
                          </pre>
                        )}
                      </div>
                    ),
                  },
                ]}
              />
            )}
          </Result>
        </div>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
