import { apiClient } from './client'

// ---------- Types ----------

export interface Project {
  project_id: string
  name: string
  base_url: string
  description: string
  status: number
  created_at?: string
  updated_at?: string
}

export interface ProjectForm {
  project_id: string
  name: string
  base_url: string
  description: string
  status: number
}

export interface ParamRule {
  name: string
  type: string
  location: string
  required: boolean
  default_value: string
  description: string
}

export interface RetryConfig {
  max_retries: number
  backoff_type: string
  retry_on_status: number[]
  retry_on_methods: string[]
}

export interface RateLimitConfig {
  max_requests: number
  window_seconds: number
}

export interface Tool {
  tool_id: number
  project_id: string
  name: string
  title: string
  description: string
  http_method: string
  url_template: string
  timeout_ms: number
  params: ParamRule[]
  retry_config?: RetryConfig | null
  rate_limit_config?: RateLimitConfig | null
  status: number
  created_at?: string
  updated_at?: string
}

export interface ToolForm {
  project_id: string
  name: string
  title: string
  description: string
  http_method: string
  url_template: string
  timeout_ms: number
  params: ParamRule[]
  retry_config?: RetryConfig | null
  rate_limit_config?: RateLimitConfig | null
  status: number
}

export interface Client {
  client_id: string
  name: string
  api_key_prefix: string
  api_key_hash?: string
  description: string
  status: number
  tool_count?: number
  created_at?: string
  updated_at?: string
}

export interface ClientForm {
  client_id: string
  name: string
  description: string
  status: number
}

export interface SessionInfo {
  id: string
  client_id: string
  protocol_version: string
  initialized: boolean
  created_at: string
}

export interface Stats {
  projects: number
  tools: number
  clients: number
  sessions: number
  session_list: SessionInfo[]
}

export interface ApiKeyResponse {
  api_key: string
}

export interface ListResponse<T> {
  data: T[]
  total?: number
}

// ---------- Projects ----------

export function getProjects(): Promise<Project[]> {
  return apiClient.get<ListResponse<Project>>('/projects').then(r => r.data || r as unknown as Project[])
}

export function createProject(data: ProjectForm): Promise<Project> {
  return apiClient.post<Project>('/projects', data)
}

export function updateProject(id: string, data: Partial<ProjectForm>): Promise<Project> {
  return apiClient.put<Project>(`/projects/${id}`, data)
}

export function deleteProject(id: string): Promise<void> {
  return apiClient.del<void>(`/projects/${id}`)
}

// ---------- Tools ----------

export function getTools(): Promise<Tool[]> {
  return apiClient.get<ListResponse<Tool>>('/tools').then(r => r.data || r as unknown as Tool[])
}

export function createTool(data: ToolForm): Promise<Tool> {
  return apiClient.post<Tool>('/tools', data)
}

export function updateTool(id: number, data: Partial<ToolForm>): Promise<Tool> {
  return apiClient.put<Tool>(`/tools/${id}`, data)
}

export function deleteTool(id: number): Promise<void> {
  return apiClient.del<void>(`/tools/${id}`)
}

// ---------- Clients ----------

export function getClients(): Promise<Client[]> {
  return apiClient.get<ListResponse<Client>>('/clients').then(r => r.data || r as unknown as Client[])
}

export function createClient(data: ClientForm): Promise<Client> {
  return apiClient.post<Client>('/clients', data)
}

export function updateClient(id: string, data: Partial<ClientForm>): Promise<Client> {
  return apiClient.put<Client>(`/clients/${id}`, data)
}

export function deleteClient(id: string): Promise<void> {
  return apiClient.del<void>(`/clients/${id}`)
}

export function generateApiKey(id: string): Promise<ApiKeyResponse> {
  return apiClient.post<ApiKeyResponse>(`/clients/${id}/api-key`)
}

// ---------- Permissions ----------

export function getClientPermissions(clientId: string): Promise<number[]> {
  return apiClient.get<{ tool_ids: number[] }>(`/clients/${clientId}/permissions`).then(r => r.tool_ids || (r as unknown as number[]))
}

export function updateClientPermissions(clientId: string, toolIds: number[]): Promise<void> {
  return apiClient.put<void>(`/clients/${clientId}/permissions`, { tool_ids: toolIds })
}

// ---------- Stats ----------

export function getStats(): Promise<Stats> {
  return apiClient.get<Stats>('/stats')
}
