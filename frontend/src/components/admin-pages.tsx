import { AdminDashboardPage } from '@/components/admin/admin-dashboard-page'
import { AdminPatientsPage } from '@/components/admin/admin-patients-page'
import { AdminDoctorsPage } from '@/components/admin/admin-doctors-page'
import { AdminAppointmentsPage } from '@/components/admin/admin-appointments-page'
import { AdminSchedulePage } from '@/components/admin/admin-schedule-page'
import { AdminRolesPage } from '@/components/admin/admin-roles-page'
import { Button } from '@/components/ui/button'

export function AdminPages({ pathname, onNavigate }: { pathname: string; onNavigate: (path: string) => void }) {
  switch (pathname) {
    case '/admin':
      return <AdminDashboardPage onNavigate={onNavigate} />
    case '/admin/patients':
      return <AdminPatientsPage />
    case '/admin/doctors':
      return <AdminDoctorsPage onNavigate={onNavigate} />
    case '/admin/appointments':
      return <AdminAppointmentsPage />
    case '/admin/schedule':
      return <AdminSchedulePage />
    case '/admin/roles':
      return <AdminRolesPage />
    default:
      return <div className="mx-auto max-w-lg px-4 py-24 text-center">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-400">404</p>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight">Admin page not found</h1>
        <Button className="mt-5" onClick={() => onNavigate('/admin')}>Back to overview</Button>
      </div>
  }
}
