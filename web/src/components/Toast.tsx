import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'

type ToastType = 'success' | 'error' | 'info' | 'warning'

interface ToastItem {
  id: number
  type: ToastType
  message: string
}

interface ToastContextType {
  toast: (type: ToastType, message: string) => void
}

const ToastContext = createContext<ToastContextType>({ toast: () => {} })

export function useToast() {
  return useContext(ToastContext)
}

let toastId = 0

const typeConfig: Record<ToastType, { bg: string; border: string; color: string }> = {
  success: { bg: '#e8f5e9', border: '#4caf50', color: '#2e7d32' },
  error: { bg: '#fbe9e7', border: '#e53935', color: '#c62828' },
  info: { bg: '#e3f2fd', border: '#1976d2', color: '#1565c0' },
  warning: { bg: '#fff3e0', border: '#ff9800', color: '#e65100' },
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const toast = useCallback((type: ToastType, message: string) => {
    const id = ++toastId
    setToasts(prev => [...prev, { id, type, message }])
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 3000)
  }, [])

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="toast-container">
        {toasts.map(t => {
          const cfg = typeConfig[t.type]
          return (
            <div
              key={t.id}
              className="toast-item"
              style={{ background: cfg.bg, borderColor: cfg.border, color: cfg.color }}
              onClick={() => setToasts(prev => prev.filter(x => x.id !== t.id))}
            >
              {t.message}
            </div>
          )
        })}
      </div>
    </ToastContext.Provider>
  )
}
