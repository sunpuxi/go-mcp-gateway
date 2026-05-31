/**
 * 将对象数组导出为 CSV 文件并触发下载
 */
export function exportToCSV<T extends Record<string, any>>(
  data: T[],
  columns: { title: string; dataIndex: string }[],
  filename: string,
): void {
  if (data.length === 0) return

  // BOM 头保证 Excel 正确识别 UTF-8 中文
  const BOM = '﻿'
  const header = columns.map(c => escapeCSV(c.title)).join(',')
  const rows = data.map(row =>
    columns.map(c => {
      const val = row[c.dataIndex]
      return escapeCSV(String(val ?? ''))
    }).join(',')
  )
  const csv = BOM + [header, ...rows].join('\n')

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${filename}_${new Date().toISOString().slice(0, 10)}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function escapeCSV(val: string): string {
  if (val.includes(',') || val.includes('"') || val.includes('\n')) {
    return `"${val.replace(/"/g, '""')}"`
  }
  return val
}

/**
 * 将对象数组导出为 JSON 文件并触发下载
 */
export function exportToJSON<T>(data: T[], filename: string): void {
  const json = JSON.stringify(data, null, 2)
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${filename}_${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
