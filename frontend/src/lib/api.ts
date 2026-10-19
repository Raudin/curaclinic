import { emitSessionExpired, emitSessionRefreshed, tokenStore } from './auth-storage'
import type { Account, ApiErrorBody, Page, TokenResponse } from './api-types'

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api').replace(/\/$/, '')

export class ApiError extends Error {
  readonly status: number
  readonly code?: string
  readonly details?: Record<string, string>
  readonly requestId?: string

  constructor(message: string, status: number, code?: string, details?: Record<string, string>, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
    this.requestId = requestId
  }
}

type QueryValue = string | number | boolean | undefined | null

export function toQuery(params?: Record<string, QueryValue>): string {
  if (!params) return ''
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    search.set(key, String(value))
  }
  const query = search.toString()
  return query ? `?${query}` : ''
}

async function readError(response: Response): Promise<ApiError> {
  let body: ApiErrorBody | null = null
  try {
    body = (await response.json()) as ApiErrorBody
  } catch {
    body = null
  }
  const payload = body?.error
  const message = payload?.message || `The clinic API returned an error (${response.status}).`
  return new ApiError(message, response.status, payload?.code, payload?.details, payload?.request_id)
}

let refreshPromise: Promise<boolean> | null = null

async function refreshAccessToken(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const refreshToken = tokenStore.getRefresh()
      if (!refreshToken) return false
      try {
        const response = await fetch(`${apiBaseUrl}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        })
        if (!response.ok) {
          tokenStore.clear()
          emitSessionExpired()
          return false
        }
        const payload = (await response.json()) as { data: TokenResponse }
        tokenStore.set(payload.data.access_token, payload.data.refresh_token)
        emitSessionRefreshed({
          access_token: payload.data.access_token,
          refresh_token: payload.data.refresh_token,
          user: payload.data.user,
          roles: payload.data.roles,
          permissions: payload.data.permissions,
        })
        return true
      } catch {
        return false
      }
    })().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

const AUTH_PATHS = ['auth/login', 'auth/register', 'auth/refresh']

export async function requestRaw<T>(path: string, init: RequestInit = {}, retry = true): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const accessToken = tokenStore.getAccess()
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)

  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}/${path.replace(/^\//, '')}`, { ...init, headers })
  } catch {
    throw new ApiError('Unable to reach the clinic API. Make sure the backend is running.', 0)
  }

  if (response.status === 401 && retry && tokenStore.getRefresh() && !AUTH_PATHS.some(item => path.startsWith(item))) {
    const refreshed = await refreshAccessToken()
    if (refreshed) return requestRaw<T>(path, init, false)
  }

  if (!response.ok) throw await readError(response)
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const payload = await requestRaw<{ data: T } | T>(path, init)
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return (payload as { data: T }).data
  }
  return payload as T
}

export function apiList<T>(path: string, init: RequestInit = {}): Promise<Page<T>> {
  return requestRaw<Page<T>>(path, init)
}

export const authApi = {
  login(email: string, password: string) {
    return apiRequest<TokenResponse>('auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },
  register(input: { first_name: string; last_name: string; email: string; phone?: string; password: string }) {
    return apiRequest<TokenResponse>('auth/register', {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  logout(refreshToken: string | null) {
    if (!refreshToken) return Promise.resolve()
    return apiRequest<void>('auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  },
  me() {
    return apiRequest<Account>('auth/me')
  },
  changePassword(currentPassword: string, newPassword: string) {
    return apiRequest<void>('auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    })
  },
}
