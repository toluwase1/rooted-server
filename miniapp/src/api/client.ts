const API_BASE = (import.meta.env.VITE_API_URL || '') + '/api'

function getInitData(): string {
  return window.Telegram?.WebApp?.initData || ''
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Telegram-Init-Data': getInitData(),
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }

  return res.json()
}

export const api = {
  // User & Profile
  getMe: () => request<any>('GET', '/me'),
  createProfile: (data: any) => request<any>('POST', '/profile', data),
  updateProfile: (data: any) => request<any>('PUT', '/profile', data),
  getProfile: (id: string) => request<any>('GET', `/profile/${id}`),
  pauseProfile: () => request<any>('POST', '/profile/pause'),
  resumeProfile: () => request<any>('POST', '/profile/resume'),
  deleteAccount: () => request<any>('DELETE', '/account'),

  // Photos
  uploadPhoto: async (file: File): Promise<any> => {
    const formData = new FormData()
    formData.append('photo', file)

    const res = await fetch(`${API_BASE}/photos`, {
      method: 'POST',
      headers: { 'X-Telegram-Init-Data': getInitData() },
      body: formData,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'Upload failed' }))
      throw new Error(err.error || `HTTP ${res.status}`)
    }
    return res.json()
  },
  deletePhoto: (photoId: string) => request<any>('DELETE', `/photos/${photoId}`),
  reorderPhotos: (photoIds: string[]) => request<any>('PUT', '/photos/reorder', { photo_ids: photoIds }),
  uploadVerificationSelfie: async (file: File): Promise<any> => {
    const formData = new FormData()
    formData.append('selfie', file)
    const res = await fetch(`${API_BASE}/verify`, {
      method: 'POST',
      headers: { 'X-Telegram-Init-Data': getInitData() },
      body: formData,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'Verification failed' }))
      throw new Error(err.error || `HTTP ${res.status}`)
    }
    return res.json()
  },

  // Matching
  getCircle: () => request<any>('GET', '/circle'),
  getExplore: (offset = 0, filters?: Record<string, string>) => {
    const params = new URLSearchParams({ offset: String(offset), ...filters })
    return request<any>('GET', `/explore?${params}`)
  },
  swipe: (candidateId: string, action: 'like' | 'pass', comment?: string, likedElement?: string) =>
    request<any>('POST', '/swipe', { candidate_id: candidateId, action, comment, liked_element: likedElement }),
  getMatches: () => request<any>('GET', '/matches'),
  getLikes: () => request<any>('GET', '/likes'),
  unmatch: (matchId: string) => request<any>('POST', `/unmatch/${matchId}`),

  // Chat
  getConversations: () => request<any>('GET', '/conversations'),
  getMessages: (convId: string, limit = 50, offset = 0) =>
    request<any>('GET', `/messages/${convId}?limit=${limit}&offset=${offset}`),

  // Payments
  getPlans: () => request<any>('GET', '/plans'),
  getSubscription: () => request<any>('GET', '/subscription'),
  getTransactions: () => request<any>('GET', '/transactions'),

  // Safety
  blockUser: (userId: string) => request<any>('POST', `/block/${userId}`),
  unblockUser: (userId: string) => request<any>('DELETE', `/block/${userId}`),
  report: (reportedId: string, category: string, description: string) =>
    request<any>('POST', '/report', { reported_id: reportedId, category, description }),
}

// Extend Window for Telegram
declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        initData: string
        initDataUnsafe: any
        ready: () => void
        expand: () => void
        close: () => void
        MainButton: {
          text: string
          show: () => void
          hide: () => void
          onClick: (cb: () => void) => void
          offClick: (cb: () => void) => void
          showProgress: (leaveActive?: boolean) => void
          hideProgress: () => void
        }
        BackButton: {
          show: () => void
          hide: () => void
          onClick: (cb: () => void) => void
          offClick: (cb: () => void) => void
        }
        HapticFeedback: {
          impactOccurred: (style: 'light' | 'medium' | 'heavy') => void
          notificationOccurred: (type: 'error' | 'success' | 'warning') => void
        }
        themeParams: Record<string, string>
        colorScheme: 'light' | 'dark'
      }
    }
  }
}
