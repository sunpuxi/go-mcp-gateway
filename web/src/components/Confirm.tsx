import { X } from 'lucide-react'

interface ConfirmProps {
  title?: string
  message: string
  onConfirm: () => void
  onCancel: () => void
  confirmText?: string
  danger?: boolean
}

export function Confirm({ title = '确认操作', message, onConfirm, onCancel, confirmText = '确定', danger = false }: ConfirmProps) {
  return (
    <div className="form-overlay" onClick={onCancel}>
      <div className="form-panel" style={{ width: 400 }} onClick={e => e.stopPropagation()}>
        <div className="form-panel-header">
          {title}
          <button className="btn-icon" onClick={onCancel}><X size={16} /></button>
        </div>
        <div className="form-panel-body" style={{ textAlign: 'center', padding: '24px' }}>
          <p style={{ fontSize: 15, color: '#555', lineHeight: 1.6 }}>{message}</p>
        </div>
        <div className="form-panel-footer">
          <button className="btn btn-outline" onClick={onCancel}>取消</button>
          <button
            className={`btn ${danger ? 'btn-danger-solid' : 'btn-primary'}`}
            onClick={onConfirm}
            autoFocus
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
