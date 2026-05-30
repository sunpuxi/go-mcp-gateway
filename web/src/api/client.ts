const API_BASE = '/admin/api'

class ApiClient {
  private apiKey: string | null = null

  setApiKey(key: string) {
    this.apiKey = key
    localStorage.setItem('admin_api_key', key)
  }

  getApiKey(): string | null {
    return this.apiKey || localStorage.getItem('admin_api_key')
  }

  clearApiKey() {
    this.apiKey = null
    localStorage.removeItem('admin_api_key')
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const apiKey = this.getApiKey()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    }
    if (apiKey) {
      headers['Authorization'] = `Bearer ${apiKey}`
    }

    const resp = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    })

    if (!resp.ok) {
      if (resp.status === 401 || resp.status === 403) {
        this.clearApiKey()
        window.location.reload()
        throw new Error('认证失败，请重新登录')
      }
      const error = await resp.json().catch(() => ({ message: resp.statusText }))
      throw new Error(error.message || `HTTP ${resp.status}`)
    }

    return resp.json()
  }

  get<T>(path: string): Promise<T> {
    return this.request<T>(path)
  }

  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  }

  put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'PUT',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  }

  del<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'DELETE' })
  }
}

export const apiClient = new ApiClient()
