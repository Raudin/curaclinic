import { useEffect, useState } from 'react'
import { FullPageLoader, AuthenticatedShell, PublicRouteGate } from '@/app-shell'
import { LoginPage } from '@/components/auth/login-page'
import { SignupPage } from '@/components/auth/signup-page'
import { homeForRoles, useAuth } from '@/lib/auth-context'

const PUBLIC_ROUTES = ['/login', '/signup']

export function AppRouter() {
  const { status, roles, permissions } = useAuth()
  const [pathname, setPathname] = useState(window.location.pathname)
  const isPublic = PUBLIC_ROUTES.includes(pathname)

  const navigate = (path: string) => {
    const [pathname] = path.split('?')
    if (path !== `${window.location.pathname}${window.location.search}`) window.history.pushState({}, '', path)
    setPathname(pathname)
  }

  useEffect(() => {
    const onPopState = () => setPathname(window.location.pathname)
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])

  useEffect(() => {
    document.querySelector<HTMLElement>('.patient-page-scroll')?.scrollTo({ top: 0, behavior: 'instant' })
  }, [pathname])

  useEffect(() => {
    if (status === 'loading') return
    if (status === 'anonymous' && !isPublic) {
      sessionStorage.setItem('curaclinic.redirect', pathname)
      navigate('/login')
      return
    }
    if (status === 'authenticated' && isPublic) {
      navigate(homeForRoles(roles, permissions))
    }
  }, [isPublic, pathname, permissions, roles, status])

  if (status === 'loading') return <FullPageLoader />
  if (status === 'anonymous') {
    if (pathname === '/signup') return <SignupPage onNavigate={navigate} />
    return <LoginPage onNavigate={navigate} />
  }
  if (isPublic) return <PublicRouteGate />
  return <AuthenticatedShell pathname={pathname} onNavigate={navigate} />
}
