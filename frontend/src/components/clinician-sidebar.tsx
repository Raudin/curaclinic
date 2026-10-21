import { CalendarDays, CalendarRange, ClipboardList, Settings } from 'lucide-react'
import { AppSidebar } from '@/components/app-sidebar'
import type { SidebarLink } from '@/components/app-sidebar'

type ClinicianSidebarProps = {
  pathname: string
  onNavigate: (path: string) => void
}

const clinicianLinks: SidebarLink[] = [
  { label: 'Daily agenda', path: '/daily-agenda', icon: CalendarDays },
  { label: 'My appointments', path: '/doctor-appointments', icon: CalendarRange },
  { label: 'Schedule & availability', path: '/schedule-availability', icon: ClipboardList },
]

export function ClinicianSidebar({ pathname, onNavigate }: ClinicianSidebarProps) {
  return <AppSidebar
    pathname={pathname}
    onNavigate={onNavigate}
    links={clinicianLinks}
    footerLink={{ label: 'Settings', path: '/settings', icon: Settings }}
    homePath="/daily-agenda"
  />
}
