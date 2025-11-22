import { useAuthStore } from '@/stores/auth'

const BASE = '/api'

class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  // Lazily get token — avoids circular deps at module load time
  let token: string | null = null
  try {
    const auth = useAuthStore()
    token = auth.token
  } catch {
    // store not yet initialized
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, { ...options, headers })

  if (res.status === 401) {
    try {
      const auth = useAuthStore()
      auth.logout()
    } catch { /* ignore */ }
    throw new ApiError(401, 'Unauthorized')
  }

  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      msg = body.error ?? body.message ?? msg
    } catch { /* ignore */ }
    throw new ApiError(res.status, msg)
  }

  if (res.status === 204) return undefined as T

  const json = await res.json()
  // Unwrap the `data` envelope our Go API always returns
  return (json.data !== undefined ? json.data : json) as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body !== undefined ? JSON.stringify(body) : undefined }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T = void>(path: string) =>
    request<T>(path, { method: 'DELETE' }),

  /** Upload a file via multipart/form-data (no Content-Type override). */
  upload: async <T>(path: string, formData: FormData): Promise<T> => {
    let token: string | null = null
    try { token = useAuthStore().token } catch { /* */ }
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`
    const res = await fetch(`${BASE}${path}`, { method: 'POST', body: formData, headers })
    if (!res.ok) throw new ApiError(res.status, `HTTP ${res.status}`)
    const json = await res.json()
    return (json.data ?? json) as T
  },
}

export { ApiError }
