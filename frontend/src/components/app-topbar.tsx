import { useState } from 'react'
import { Settings } from 'lucide-react'
import { SidebarTrigger } from '@/components/ui/sidebar'
import { Breadcrumb } from '@/components/arc/breadcrumb/breadcrumb'
import { UserMenu } from '@/components/arc/user-menu/user-menu'
import { NotificationCenter } from '@/components/arc/notification-center/notification-center'
import type { NotificationItem } from '@/components/arc/notification-center/notification-center'
import { useAuth } from '@/lib/auth-context'
import { friendlyRole } from '@/lib/permissions'

type AppTopbarProps = {
  variant: 'patient' | 'admin' | 'clinician'
  pathname: string
  onNavigate: (path: string) => void
}

const titles: Record<string, string> = {
  '/': 'Overview',
  '/book-appointment': 'Book appointment',
  '/my-appointments': 'My appointments',
  '/settings': 'Settings',
  '/daily-agenda': 'Daily agenda',
  '/schedule-availability': 'Schedule & availability',
  '/admin': 'Overview',
  '/admin/patients': 'Patients',
  '/admin/doctors': 'Doctors',
  '/admin/appointments': 'Appointments',
  '/admin/schedule': 'Schedule & availability',
  '/admin/roles': 'Roles & access',
  '/admin/settings': 'Settings',
}

export function AppTopbar({ variant, pathname, onNavigate }: AppTopbarProps) {
  const { user, roles, logout } = useAuth()
  const [notifications] = useState<NotificationItem[]>([])
  const rootPath = variant === 'admin' ? '/admin' : variant === 'clinician' ? '/daily-agenda' : '/'
  const settingsPath = variant === 'admin' ? '/admin/settings' : '/settings'
  const title = titles[pathname] ?? (pathname.startsWith('/admin') ? 'Administration' : 'Overview')
  const primaryRole = roles[0] ?? 'patient'
  const displayUser = {
    name: user ? `${user.first_name} ${user.last_name}`.trim() : 'CuraClinic user',
    email: user?.email ?? '',
    plan: friendlyRole(primaryRole),
  }

  async function signOut() {
    await logout()
    onNavigate('/login')
  }

  return <header data-patient-topbar className="patient-topbar border-b border-border/70 bg-background/90 px-4 backdrop-blur-xl sm:px-7 lg:px-10">
    <div className="flex h-[72px] items-center justify-between gap-3">
      <div className="flex min-w-0 items-center gap-3"><SidebarTrigger /><Breadcrumb items={pathname === rootPath ? [{ label: title }] : [{ label: variant === 'admin' ? 'Admin' : variant === 'clinician' ? 'Clinic' : 'Overview', onClick: () => onNavigate(rootPath) }, { label: title }]} /></div>
      <div className="flex shrink-0 items-center gap-1 sm:gap-2">
        <NotificationCenter
          notifications={notifications}
          onReadChange={() => undefined}
          onDismiss={() => undefined}
        />
        <UserMenu
          user={displayUser}
          showName
          showTheme={false}
          items={[
            { label: 'Settings', icon: <Settings size={15} />, onSelect: () => onNavigate(settingsPath) },
          ]}
          onSignOut={signOut}
        />
      </div>
    </div>
  </header>
}
