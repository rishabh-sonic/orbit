import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

// We import api AFTER setting up pinia so useAuthStore() resolves
let api: typeof import('@/lib/api').api
let ApiError: typeof import('@/lib/api').ApiError

beforeEach(async () => {
  setActivePinia(createPinia())
  const mod = await import('@/lib/api')
  api = mod.api
  ApiError = mod.ApiError
  vi.mocked(fetch).mockReset()
})

function mockFetch(status: number, body: unknown) {
  vi.mocked(fetch).mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  )
}

// --- GET ---

describe('api.get', () => {
  it('returns unwrapped data on 200', async () => {
    mockFetch(200, { data: { id: '1', name: 'Alice' } })
    const result = await api.get<{ id: string; name: string }>('/users/me')
    expect(result).toEqual({ id: '1', name: 'Alice' })
  })

  it('returns raw json when no data envelope', async () => {
    mockFetch(200, { token: 'abc123' })
    const result = await api.get<{ token: string }>('/auth/check')
    expect(result.token).toBe('abc123')
  })

  it('attaches Authorization header when token present', async () => {
    const auth = useAuthStore()
    auth.setToken('my-jwt-token')
    mockFetch(200, { data: [] })
    await api.get('/posts')
    const call = vi.mocked(fetch).mock.calls[0]
    const headers = call[1]?.headers as Record<string, string>
    expect(headers['Authorization']).toBe('Bearer my-jwt-token')
  })

  it('throws ApiError on 404', async () => {
    mockFetch(404, { error: 'not found' })
    await expect(api.get('/missing')).rejects.toBeInstanceOf(ApiError)
  })

  it('throws ApiError on 500', async () => {
    mockFetch(500, { error: 'server error' })
    await expect(api.get('/broken')).rejects.toBeInstanceOf(ApiError)
  })

  it('calls logout and throws on 401', async () => {
    const auth = useAuthStore()
    auth.setToken('expired-token')
    mockFetch(401, { error: 'unauthorized' })
    await expect(api.get('/secret')).rejects.toBeInstanceOf(ApiError)
    expect(auth.token).toBeNull()
  })
})

// --- POST ---

describe('api.post', () => {
  it('sends JSON body', async () => {
    mockFetch(201, { data: { id: '42' } })
    await api.post('/posts', { title: 'Hello', content: 'World' })
    const call = vi.mocked(fetch).mock.calls[0]
    expect(call[1]?.method).toBe('POST')
    const body = JSON.parse(call[1]?.body as string)
    expect(body.title).toBe('Hello')
  })

  it('handles empty body', async () => {
    mockFetch(200, { data: { ok: true } })
    await expect(api.post('/notifications/read')).resolves.toBeDefined()
  })
})

// --- PUT ---

describe('api.put', () => {
  it('sends PUT method', async () => {
    mockFetch(200, { data: { id: '1' } })
    await api.put('/users/me', { username: 'new_alice' })
    expect(vi.mocked(fetch).mock.calls[0][1]?.method).toBe('PUT')
  })
})

// --- DELETE ---

describe('api.delete', () => {
  it('returns undefined on 204', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 204 }))
    const result = await api.delete('/posts/1')
    expect(result).toBeUndefined()
  })

  it('sends DELETE method', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 204 }))
    await api.delete('/posts/1')
    expect(vi.mocked(fetch).mock.calls[0][1]?.method).toBe('DELETE')
  })
})

// --- ApiError ---

describe('ApiError', () => {
  it('has correct status', async () => {
    mockFetch(422, { error: 'validation failed' })
    try {
      await api.get('/validate')
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      expect((e as InstanceType<typeof ApiError>).status).toBe(422)
      expect((e as InstanceType<typeof ApiError>).message).toBe('validation failed')
    }
  })
})
