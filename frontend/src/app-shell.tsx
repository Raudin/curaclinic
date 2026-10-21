import { Suspense, lazy, type ReactNode } from 'react'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { TooltipProvider } from '@/components/ui/tooltip'
import { PatientDashboard } from '@/components/patient-dashboard'
import { PatientSidebar } from '@/components/patient-sidebar'
import { AdminSidebar } from '@/components/admin-sidebar'
import { ClinicianSidebar } from '@/components/clinician-sidebar'
import { AppTopbar } from '@/components/app-topbar'
import { MyAppointmentsPage } from '@/components/my-appointments-page'
import { ClinicianPages } from '@/components/clinician-pages'
import { AdminPages } from '@/components/admin-pages'
import { SettingsPage } from '@/components/settings-page'
import { homeForRoles, useAuth } from '@/lib/auth-context'
import { PERMISSIONS, hasScopePermission, resolvePermissions } from '@/lib/permissions'
import { Button } from '@/components/ui/button'

const BookAppointmentPage = lazy(() => import('@/components/book-appointment-page').then(module => ({ default: module.BookAppointmentPage })))

export function FullPageLoader({ label = 'Preparing your clinic…' }: { label?: string }) {
  return <div className="flex min-h-svh items-center justify-center bg-[#f7f8f7] px-6 text-center">
    <div>
      <div className="mx-auto size-8 animate-pulse rounded-xl bg-neutral-900" aria-hidden="true" />
      <p className="mt-4 text-sm text-neutral-500">{label}</p>
    </div>
  </div>
}

export function PublicRouteGate() {
  return <FullPageLoader label="Taking you to your dashboard…" />
}

function AccessDenied({ onHome }: { onHome: () => void }) {
  return <div className="mx-auto flex max-w-lg flex-col items-center px-4 py-24 text-center">
    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-400">403</p>
    <h1 className="mt-2 text-2xl font-semibold tracking-tight">You don’t have access to this page</h1>
    <p className="mt-2 text-sm text-muted-foreground">Your clinic role doesn’t include this area. Return to your dashboard to continue.</p>
    <Button className="mt-5" onClick={onHome}>Go to my dashboard</Button>
  </div>
}

function roleVariant(roles: string[], permissions: string[]): 'patient' | 'admin' | 'clinician' {
  const granted = resolvePermissions(roles, permissions)
  if (roles.includes('admin') && hasScopePermission(granted, 'admin')) return 'admin'
  if (roles.includes('doctor') && hasScopePermission(granted, 'doctor')) return 'clinician'
  return 'patient'
}

export function AuthenticatedShell({ pathname, onNavigate }: { pathname: string; onNavigate: (path: string) => void }) {
  const { roles, permissions, can, canAny, canAll } = useAuth()
  const variant = roleVariant(roles, permissions)
  const isAdminArea = pathname.startsWith('/admin')
  const isClinicianArea = pathname === '/daily-agenda' || pathname === '/doctor-appointments' || pathname === '/schedule-availability'
  const isSharedSettings = pathname === '/settings'
  const isRescheduleForClinician = variant === 'clinician' && pathname === '/book-appointment' && new URLSearchParams(window.location.search).has('reschedule')
  const isPatientArea = !isAdminArea && !isClinicianArea && !isSharedSettings && !isRescheduleForClinician

  if (isAdminArea && variant !== 'admin') return <AccessDenied onHome={() => onNavigate(homeForRoles(roles, permissions))} />
  if (isClinicianArea && variant !== 'clinician') return <AccessDenied onHome={() => onNavigate(homeForRoles(roles, permissions))} />
  if (isPatientArea && variant !== 'patient') return <AccessDenied onHome={() => onNavigate(homeForRoles(roles, permissions))} />
  if ((pathname === '/daily-agenda' || pathname === '/doctor-appointments') && !can(PERMISSIONS.doctorAppointmentsViewOwn)) return <AccessDenied onHome={() => onNavigate(homeForRoles(roles, permissions))} />
  if (pathname === '/schedule-availability' && !can(PERMISSIONS.doctorScheduleView)) return <AccessDenied onHome={() => onNavigate(homeForRoles(roles, permissions))} />
  if (pathname === '/admin/roles' && !canAll(PERMISSIONS.adminRolesManage, PERMISSIONS.adminPermissionsManage)) return <AccessDenied onHome={() => onNavigate('/admin')} />
  if (pathname === '/admin/schedule' && !(can(PERMISSIONS.adminWorkingHours) && canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView))) return <AccessDenied onHome={() => onNavigate('/admin')} />
  if (pathname === '/admin/patients' && !can(PERMISSIONS.adminUsersView)) return <AccessDenied onHome={() => onNavigate('/admin')} />
  if (pathname === '/admin/doctors' && !canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)) return <AccessDenied onHome={() => onNavigate('/admin')} />
  if (pathname === '/admin/appointments' && !can(PERMISSIONS.adminAppointmentsView)) return <AccessDenied onHome={() => onNavigate('/admin')} />

  let content: ReactNode
  if (isAdminArea && pathname !== '/admin/settings') {
    content = <AdminPages pathname={pathname} onNavigate={onNavigate} />
  } else if (isClinicianArea) {
    content = <ClinicianPages pathname={pathname} onNavigate={onNavigate} />
  } else if (pathname === '/admin/settings') {
    content = <SettingsPage variant="admin" onNavigate={onNavigate} />
  } else if (pathname === '/book-appointment') {
    content = <Suspense fallback={<div className="p-8 text-sm text-neutral-500">Loading appointment booking…</div>}><BookAppointmentPage onNavigate={onNavigate} /></Suspense>
  } else if (pathname === '/my-appointments') {
    content = <MyAppointmentsPage onNavigate={onNavigate} />
  } else if (pathname === '/settings') {
    content = <SettingsPage variant={variant === 'clinician' ? 'clinician' : variant === 'admin' ? 'admin' : 'patient'} onNavigate={onNavigate} />
  } else if (pathname === '/') {
    content = <PatientDashboard onNavigate={onNavigate} />
  } else {
    content = <div className="mx-auto max-w-lg px-4 py-24 text-center">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-400">404</p>
      <h1 className="mt-2 text-2xl font-semibold tracking-tight">Page not found</h1>
      <Button className="mt-5" onClick={() => onNavigate(homeForRoles(roles, permissions))}>Return home</Button>
    </div>
  }

  return <TooltipProvider>
    <SidebarProvider defaultOpen className="min-h-screen bg-[#f7f8f7] text-neutral-900">
      {variant === 'admin' ? <AdminSidebar pathname={pathname} onNavigate={onNavigate} /> : variant === 'clinician' ? <ClinicianSidebar pathname={pathname} onNavigate={onNavigate} /> : <PatientSidebar pathname={pathname} onNavigate={onNavigate} />}
      <SidebarInset className="h-svh min-h-0 min-w-0 overflow-hidden bg-[#f7f8f7]">
        <AppTopbar variant={variant} pathname={pathname} onNavigate={onNavigate} />
        <div className="h-[72px] shrink-0" aria-hidden="true" />
        <div className="patient-page-scroll">{content}</div>
      </SidebarInset>
    </SidebarProvider>
  </TooltipProvider>
}
