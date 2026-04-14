const API_BASE = (import.meta.env.VITE_API_URL || '') + '/admin'
const AUTH_BASE = (import.meta.env.VITE_API_URL || '') + '/admin/auth'

function getToken(): string {
  return localStorage.getItem('admin_token') || ''
}

export function setToken(token: string) {
  localStorage.setItem('admin_token', token)
}

export function clearToken() {
  localStorage.removeItem('admin_token')
}

export function isLoggedIn(): boolean {
  return !!getToken()
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    clearToken()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const adminApi = {
  // Auth
  login: async (email: string, password: string) => {
    const res = await fetch(`${AUTH_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'Login failed' }))
      throw new Error(err.error)
    }
    return res.json()
  },
  getMe: () => request<any>('GET', '/../admin/auth/me'),

  // Stats
  getStats: () => request<any>('GET', '/stats'),

  // Config
  getConfig: () => request<any>('GET', '/config'),
  updateConfig: (key: string, value: any, reason: string) =>
    request<any>('PUT', `/config/${key}`, { value, reason }),
  getConfigHistory: (key: string) => request<any>('GET', `/config/history/${key}`),

  // Features
  getFeatures: () => request<any>('GET', '/features'),
  updateFeature: (key: string, data: any) => request<any>('PUT', `/features/${key}`, data),

  // Reports
  getReports: (status = 'pending') => request<any>('GET', `/reports?status=${status}`),
  reviewReport: (id: string, action: string) => request<any>('PUT', `/reports/${id}`, { action }),

  // Users
  listUsers: (search = '', status = '', limit = 50, offset = 0) =>
    request<any>('GET', `/users?search=${search}&status=${status}&limit=${limit}&offset=${offset}`),
  getUser: (id: string) => request<any>('GET', `/users/${id}`),
  updateUserStatus: (id: string, status: string) =>
    request<any>('PUT', `/users/${id}/status`, { status }),

  // Logs
  getLogs: (limit = 100, range_ = '1h', severity = '', search = '') =>
    request<any>('GET', `/logs?limit=${limit}&range=${range_}&severity=${severity}&search=${search}`),
  getLogStats: (range_ = '24h') =>
    request<any>('GET', `/logs/stats?range=${range_}`),
}
