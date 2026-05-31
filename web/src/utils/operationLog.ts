export interface OperationRecord {
  id: number
  action: string
  target: string
  detail: string
  timestamp: string
}

const STORAGE_KEY = 'mcp_gateway_ops_log'
const MAX_ENTRIES = 20

let nextId = Date.now()

/** 从 sessionStorage 读取操作日志 */
export function getOperationLogs(): OperationRecord[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    return JSON.parse(raw) as OperationRecord[]
  } catch {
    return []
  }
}

/** 追加一条操作日志 */
export function appendOperationLog(
  action: string,
  target: string,
  detail: string,
): void {
  try {
    const logs = getOperationLogs()
    logs.unshift({
      id: nextId++,
      action,
      target,
      detail,
      timestamp: new Date().toLocaleString('zh-CN'),
    })
    // 只保留最近 N 条
    if (logs.length > MAX_ENTRIES) {
      logs.length = MAX_ENTRIES
    }
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(logs))
  } catch {
    // 静默失败，不影响主流程
  }
}
