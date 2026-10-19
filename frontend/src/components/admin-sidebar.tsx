import { Activity, CalendarDays, ClipboardList, KeyRound, LayoutDashboard, Settings, Users } from 'lucide-react'
import { AppSidebar } from '@/components/app-sidebar'
import type { SidebarLink } from '@/components/app-sidebar'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'

type AdminSidebarProps = {
  pathname: string
  onNavigate: (path: string) => void
}

export function AdminSidebar({ pathname, onNavigate }: AdminSidebarProps) {
  const { can, canAny, canAll } = useAuth()
  const adminLinks: SidebarLink[] = [
    { label: 'Overview', path: '/admin', icon: LayoutDashboard, matches: (value: string) => value === '/admin' || value === '/admin/' },
  ]
  if (can(PERMISSIONS.adminUsersView)) adminLinks.push({ label: 'Patients', path: '/admin/patients', icon: Users })
  if (canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)) {
    adminLinks.push({ label: 'Doctors', path: '/admin/doctors', icon: Activity })
  }
  if (can(PERMISSIONS.adminAppointmentsView)) {
    adminLinks.push({ label: 'Appointments', path: '/admin/appointments', icon: CalendarDays })
  }
  if (can(PERMISSIONS.adminWorkingHours) && canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)) {
    adminLinks.push({ label: 'Schedule & Availability', path: '/admin/schedule', icon: ClipboardList })
  }
  if (canAll(PERMISSIONS.adminRolesManage, PERMISSIONS.adminPermissionsManage)) {
    adminLinks.push({ label: 'Roles & access', path: '/admin/roles', icon: KeyRound })
  }

  return <AppSidebar
    pathname={pathname}
    onNavigate={onNavigate}
    links={adminLinks}
    footerLink={{ label: 'Settings', path: '/admin/settings', icon: Settings }}
    homePath="/admin"
  />
}
