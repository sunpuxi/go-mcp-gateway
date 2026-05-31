import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, theme, App as AntApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Toaster } from 'react-hot-toast'
import { AuthProvider } from './context/AuthContext'
import { ThemeProvider, useTheme } from './context/ThemeContext'
import App from './App'
import './styles/global.css'
import './styles/layout.css'
import './styles/pages.css'
import './styles/components.css'

/** 内层组件：需要访问 ThemeContext 来配置主题 */
function ThemedApp() {
  const { isDark } = useTheme()

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 8,
          colorBgLayout: isDark ? '#141414' : '#f5f7fa',
          fontFamily: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif`,
        },
        components: {
          Layout: {
            headerBg: isDark ? '#141414' : '#ffffff',
            headerHeight: 56,
            siderBg: '#001529',
          },
          Menu: {
            darkItemBg: '#001529',
            darkItemSelectedBg: '#1677ff',
            darkItemHoverBg: 'rgba(255,255,255,0.08)',
            itemBorderRadius: 6,
          },
          Card: {
            paddingLG: 24,
          },
          Table: {
            headerBg: isDark ? '#1f1f1f' : '#fafbfc',
            rowHoverBg: isDark ? '#111d2c' : '#f0f5ff',
          },
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <App />
        </BrowserRouter>
        <Toaster
          position="top-right"
          toastOptions={{
            style: {
              background: isDark ? '#2a2a2a' : '#fff',
              color: isDark ? '#e8e8e8' : '#333',
              border: isDark ? '1px solid #303030' : '1px solid #f0f0f0',
            },
          }}
        />
      </AntApp>
    </ConfigProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AuthProvider>
      <ThemeProvider>
        <ThemedApp />
      </ThemeProvider>
    </AuthProvider>
  </React.StrictMode>,
)
