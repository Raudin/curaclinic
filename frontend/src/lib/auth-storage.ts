const ACCESS_TOKEN_KEY = 'curaclinic.access_token'
const REFRESH_TOKEN_KEY = 'curaclinic.refresh_token'
const LAST_ACTIVITY_KEY = 'curaclinic.last_activity'

// Incremented whenever tokens are written or cleared so an in-flight refresh
// response cannot resurrect a session the user (or the idle timer) just signed
// out of, and so a stale refresh cannot overwrite a newer login.
let sessionEpoch = 0

export const tokenStore = {
  getAccess(): string | null {
    return window.localStorage.getItem(ACCESS_TOKEN_KEY)
  },
  getRefresh(): string | null {
    return window.localStorage.getItem(REFRESH_TOKEN_KEY)
  },
  set(accessToken: string, refreshToken: string) {
    sessionEpoch += 1
    window.localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  clear() {
    sessionEpoch += 1
    window.localStorage.removeItem(ACCESS_TOKEN_KEY)
    window.localStorage.removeItem(REFRESH_TOKEN_KEY)
    window.localStorage.removeItem(LAST_ACTIVITY_KEY)
  },
  epoch(): number {
    return sessionEpoch
  },
  hasSession(): boolean {
    return Boolean(window.localStorage.getItem(ACCESS_TOKEN_KEY) || window.localStorage.getItem(REFRESH_TOKEN_KEY))
  },
}

// Last user activity across every tab, so an idle background tab does not sign
// out a user who is working in another tab.
export const activityStore = {
  get(): number {
    const value = Number(window.localStorage.getItem(LAST_ACTIVITY_KEY))
    return Number.isFinite(value) ? value : 0
  },
  touch(at = Date.now()) {
    window.localStorage.setItem(LAST_ACTIVITY_KEY, String(at))
  },
}

type TokenRefreshDetail = {
  access_token: string
  refresh_token: string
  user: import('./api-types').User
  roles: string[]
  permissions: string[]
}

export type AuthEventMap = {
  expired: Event
  refreshed: CustomEvent<TokenRefreshDetail>
}

export const authEvents = new EventTarget()

export function emitSessionExpired() {
  authEvents.dispatchEvent(new Event('expired'))
}

export function emitSessionRefreshed(detail: TokenRefreshDetail) {
  authEvents.dispatchEvent(new CustomEvent('refreshed', { detail }))
}
