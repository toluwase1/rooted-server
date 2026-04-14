const API_BASE = '/admin'

// Admin uses Telegram init data for auth — passed from Mini App or stored locally
function getInitData(): string {
  return localStorage.getItem('admin_init_data') || ''
}

export function setInitData(data: string) {
  localStorage.setItem('admin_init_data', data)
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      'X-Telegram-Init-Data': getInitData(),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const adminApi = {
  getStats: () => request<any>('GET', '/stats'),
  getConfig: () => request<any>('GET', '/config'),
  updateConfig: (key: string, value: any, reason: string) =>
    request<any>('PUT', `/config/${key}`, { value, reason }),
  getConfigHistory: (key: string) => request<any>('GET', `/config/history/${key}`),
  getFeatures: () => request<any>('GET', '/features'),
  updateFeature: (key: string, data: any) => request<any>('PUT', `/features/${key}`, data),
  getReports: () => request<any>('GET', '/reports'),
  reviewReport: (id: string, action: string) => request<any>('PUT', `/reports/${id}`, { action }),
}
