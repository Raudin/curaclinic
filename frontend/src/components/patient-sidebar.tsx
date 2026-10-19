import { CalendarDays, CalendarPlus, LayoutDashboard, Settings } from 'lucide-react'
import { AppSidebar } from '@/components/app-sidebar'
import type { SidebarLink } from '@/components/app-sidebar'

type PatientSidebarProps = {
  pathname: string
  onNavigate: (path: string) => void
}

const patientLinks: SidebarLink[] = [
  { label: 'Overview', path: '/', icon: LayoutDashboard, matches: (pathname: string) => !['/book-appointment', '/my-appointments', '/settings'].includes(pathname) },
  { label: 'Book appointment', path: '/book-appointment', icon: CalendarPlus },
  { label: 'My appointments', path: '/my-appointments', icon: CalendarDays },
]

export function PatientSidebar({ pathname, onNavigate }: PatientSidebarProps) {
  return <AppSidebar pathname={pathname} onNavigate={onNavigate} links={patientLinks} footerLink={{ label: 'Settings', path: '/settings', icon: Settings }} homePath="/" />
}
