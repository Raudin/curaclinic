const ACCESS_TOKEN_KEY = 'curaclinic.access_token'
const REFRESH_TOKEN_KEY = 'curaclinic.refresh_token'

export const tokenStore = {
  getAccess(): string | null {
    return window.localStorage.getItem(ACCESS_TOKEN_KEY)
  },
  getRefresh(): string | null {
    return window.localStorage.getItem(REFRESH_TOKEN_KEY)
  },
  set(accessToken: string, refreshToken: string) {
    window.localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  clear() {
    window.localStorage.removeItem(ACCESS_TOKEN_KEY)
    window.localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
  hasSession(): boolean {
    return Boolean(window.localStorage.getItem(ACCESS_TOKEN_KEY) || window.localStorage.getItem(REFRESH_TOKEN_KEY))
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
