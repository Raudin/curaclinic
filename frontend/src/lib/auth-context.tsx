import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { ApiError, authApi, refreshSession } from '@/lib/api'
import { activityStore, authEvents, tokenStore } from '@/lib/auth-storage'
import { PERMISSIONS, hasScopePermission, resolvePermissions } from '@/lib/permissions'
import type { Account, TokenResponse, User } from '@/lib/api-types'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

type AuthContextValue = {
  status: AuthStatus
  user: User | null
  roles: string[]
  permissions: string[]
  isAuthenticated: boolean
  hasRole: (role: string) => boolean
  can: (permission: string) => boolean
  canAny: (...permissions: string[]) => boolean
  canAll: (...permissions: string[]) => boolean
  login: (email: string, password: string) => Promise<TokenResponse>
  register: (input: { first_name: string; last_name: string; email: string; phone?: string; password: string }) => Promise<TokenResponse>
  logout: () => Promise<void>
  changePassword: (currentPassword: string, newPassword: string) => Promise<void>
  refreshAccount: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const ROLE_PRIORITY = ['admin', 'doctor', 'patient'] as const

const DEFAULT_IDLE_MINUTES = 30

function idleTimeoutMs(): number {
  const configured = Number(import.meta.env.VITE_SESSION_IDLE_MINUTES)
  const minutes = Number.isFinite(configured) && configured > 0 ? configured : DEFAULT_IDLE_MINUTES
  return minutes * 60_000
}

export function homeForRoles(roles: string[], permissions: string[] = []): string {
  const granted = resolvePermissions(roles, permissions)
  for (const role of ROLE_PRIORITY) {
    if (!roles.includes(role)) continue
    if (role === 'admin' && hasScopePermission(granted, 'admin')) return '/admin'
    if (role === 'doctor' && hasScopePermission(granted, 'doctor')) {
      if (granted.has(PERMISSIONS.doctorAppointmentsViewOwn)) return '/daily-agenda'
      if (granted.has(PERMISSIONS.doctorScheduleView)) return '/schedule-availability'
      return '/settings'
    }
  }
  return '/'
}

type AuthState = {
  status: AuthStatus
  user: User | null
  roles: string[]
  permissions: string[]
}

const emptyState: AuthState = { status: 'loading', user: null, roles: [], permissions: [] }

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(emptyState)

  const applyTokens = useCallback((data: TokenResponse) => {
    setState({ status: 'authenticated', user: data.user, roles: data.roles, permissions: data.permissions })
  }, [])

  const applyAccount = useCallback((account: Account) => {
    setState({ status: 'authenticated', user: account.user, roles: account.roles, permissions: account.permissions })
  }, [])

  const clear = useCallback(() => {
    tokenStore.clear()
    setState({ status: 'anonymous', user: null, roles: [], permissions: [] })
  }, [])

  useEffect(() => {
    let cancelled = false

    async function bootstrap() {
      if (!tokenStore.hasSession()) {
        if (!cancelled) setState({ status: 'anonymous', user: null, roles: [], permissions: [] })
        return
      }
      try {
        const account = await authApi.me()
        if (!cancelled) applyAccount(account)
      } catch {
        if (!cancelled) clear()
      }
    }

    void bootstrap()

    const onExpired = () => {
      if (!cancelled) setState({ status: 'anonymous', user: null, roles: [], permissions: [] })
    }
    const onRefreshed = (event: Event) => {
      const detail = (event as CustomEvent<TokenResponse>).detail
      if (!cancelled) {
        setState({ status: 'authenticated', user: detail.user, roles: detail.roles, permissions: detail.permissions })
      }
    }
    authEvents.addEventListener('expired', onExpired)
    authEvents.addEventListener('refreshed', onRefreshed)

    return () => {
      cancelled = true
      authEvents.removeEventListener('expired', onExpired)
      authEvents.removeEventListener('refreshed', onRefreshed)
    }
  }, [applyAccount, clear])

  const status = state.status

  useEffect(() => {
    if (status !== 'authenticated') return

    const timeoutMs = idleTimeoutMs()
    const syncIntervalMs = Math.max(60_000, Math.min(timeoutMs / 2, 10 * 60_000))
    const sharedWriteIntervalMs = 5_000
    let timer: number | undefined
    let lastActivityAt = Date.now()
    let lastSharedWriteAt = lastActivityAt
    let lastSyncedAt = Date.now()
    let expired = false
    activityStore.touch(lastActivityAt)

    // Activity in any tab keeps the session alive in every tab.
    const latestActivity = () => Math.max(lastActivityAt, activityStore.get())

    const clearTimer = () => {
      if (timer !== undefined) window.clearTimeout(timer)
      timer = undefined
    }

    const expire = () => {
      if (expired) return
      expired = true
      clearTimer()
      const refreshToken = tokenStore.getRefresh()
      // Clear locally first so the UI signs out even if the API call fails;
      // tokenStore.clear() also invalidates in-flight refresh responses.
      clear()
      void authApi.logout(refreshToken).catch(() => undefined)
    }

    const schedule = () => {
      clearTimer()
      timer = window.setTimeout(checkIdle, Math.max(0, timeoutMs - (Date.now() - latestActivity())))
    }

    // Re-read shared activity before signing out: another tab may have been
    // active since this timer was set.
    function checkIdle() {
      if (Date.now() - latestActivity() >= timeoutMs) expire()
      else schedule()
    }

    const syncSession = () => {
      if (expired) return
      const now = Date.now()
      if (now - lastSyncedAt < syncIntervalMs) return
      lastSyncedAt = now
      void refreshSession()
    }

    const recordActivity = () => {
      if (expired) return
      lastActivityAt = Date.now()
      if (lastActivityAt - lastSharedWriteAt >= sharedWriteIntervalMs) {
        lastSharedWriteAt = lastActivityAt
        activityStore.touch(lastActivityAt)
      }
      schedule()
      syncSession()
    }

    const onVisibility = () => {
      if (expired || document.visibilityState !== 'visible') return
      // Background tabs throttle timers, so compare timestamps instead of
      // trusting the scheduled callback to have fired on time.
      if (Date.now() - latestActivity() >= timeoutMs) {
        expire()
        return
      }
      recordActivity()
      authApi.me().then(applyAccount).catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 401) clear()
      })
    }

    const onRefreshed = () => {
      lastSyncedAt = Date.now()
      recordActivity()
    }

    const activityEvents = ['pointerdown', 'keydown', 'wheel', 'touchstart'] as const
    for (const event of activityEvents) window.addEventListener(event, recordActivity, { passive: true })
    document.addEventListener('visibilitychange', onVisibility)
    authEvents.addEventListener('refreshed', onRefreshed)
    schedule()

    return () => {
      clearTimer()
      for (const event of activityEvents) window.removeEventListener(event, recordActivity)
      document.removeEventListener('visibilitychange', onVisibility)
      authEvents.removeEventListener('refreshed', onRefreshed)
    }
  }, [applyAccount, clear, status])

  useEffect(() => {
    const onStorage = () => {
      if (tokenStore.hasSession()) return
      // Another tab signed out; discard any refresh this tab has in flight.
      tokenStore.clear()
      setState({ status: 'anonymous', user: null, roles: [], permissions: [] })
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const result = await authApi.login(email, password)
    tokenStore.set(result.access_token, result.refresh_token)
    applyTokens(result)
    return result
  }, [applyTokens])

  const register = useCallback(async (input: { first_name: string; last_name: string; email: string; phone?: string; password: string }) => {
    const result = await authApi.register(input)
    tokenStore.set(result.access_token, result.refresh_token)
    applyTokens(result)
    return result
  }, [applyTokens])

  const logout = useCallback(async () => {
    try {
      await authApi.logout(tokenStore.getRefresh())
    } finally {
      clear()
    }
  }, [clear])

  const changePassword = useCallback(async (currentPassword: string, newPassword: string) => {
    await authApi.changePassword(currentPassword, newPassword)
    clear()
  }, [clear])

  const refreshAccount = useCallback(async () => {
    const account = await authApi.me()
    applyAccount(account)
  }, [applyAccount])

  const permissionSet = useMemo(
    () => resolvePermissions(state.roles, state.permissions),
    [state.permissions, state.roles],
  )

  const value = useMemo<AuthContextValue>(() => {
    const can = (permission: string) => permissionSet.has(permission)
    return {
      status: state.status,
      user: state.user,
      roles: state.roles,
      permissions: state.permissions,
      isAuthenticated: state.status === 'authenticated',
      hasRole: (role: string) => state.roles.includes(role),
      can,
      canAny: (...permissions: string[]) => permissions.some(can),
      canAll: (...permissions: string[]) => permissions.every(can),
      login,
      register,
      logout,
      changePassword,
      refreshAccount,
    }
  }, [changePassword, login, logout, permissionSet, refreshAccount, register, state])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}
