import { useCallback, useEffect, useState } from 'react'
import { Activity, CalendarDays, Clock3, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState, ErrorState, LoadingState, PageHeader, StatusBadge } from '@/components/shared/page'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import { auditApi, usersApi } from '@/lib/api-admin'
import type { Appointment, AuditLog, Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, formatDate, formatTimeRange, zonedDayBounds } from '@/lib/time'

const CLINIC_TIMEZONE = import.meta.env.VITE_CLINIC_TIMEZONE || DEFAULT_TIMEZONE

export function AdminDashboardPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { can, canAny, canAll } = useAuth()
  const canViewUsers = can(PERMISSIONS.adminUsersView)
  const canViewDoctors = canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)
  const canViewAppointments = can(PERMISSIONS.adminAppointmentsView)
  const canViewAudit = can(PERMISSIONS.adminAuditView)
  const canManageAccess = canAll(PERMISSIONS.adminRolesManage, PERMISSIONS.adminPermissionsManage)
  const hasOverview = canViewUsers || canViewDoctors || canViewAppointments || canViewAudit
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [patients, setPatients] = useState(0)
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [upcoming, setUpcoming] = useState<Appointment[]>([])
  const [upcomingTotal, setUpcomingTotal] = useState(0)
  const [today, setToday] = useState<Appointment[]>([])
  const [todayTotal, setTodayTotal] = useState(0)
  const [audit, setAudit] = useState<AuditLog[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const bounds = zonedDayBounds(new Date(), CLINIC_TIMEZONE)
      const [patientPage, doctorList, upcomingPage, todayPage, auditPage] = await Promise.all([
        canViewUsers ? usersApi.list({ role: 'patient', limit: 1 }) : Promise.resolve(null),
        canViewDoctors ? doctorsApi.list() : Promise.resolve(null),
        canViewAppointments ? appointmentsApi.list({ scope: 'upcoming', limit: 5 }) : Promise.resolve(null),
        canViewAppointments ? appointmentsApi.list({ scope: 'all', from: bounds.from, to: bounds.to, order: 'asc', limit: 10 }) : Promise.resolve(null),
        canViewAudit ? auditApi.list({ limit: 5 }) : Promise.resolve(null),
      ])
      setPatients(patientPage?.pagination.total ?? 0)
      setDoctors(doctorList ?? [])
      setUpcoming(upcomingPage?.data ?? [])
      setUpcomingTotal(upcomingPage?.pagination.total ?? 0)
      setToday(todayPage?.data ?? [])
      setTodayTotal(todayPage?.pagination.total ?? 0)
      setAudit(auditPage?.data ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load the clinic overview.')
    } finally {
      setLoading(false)
    }
  }, [canViewAppointments, canViewAudit, canViewDoctors, canViewUsers])

  useEffect(() => { void load() }, [load])

  const cards = [
    ...(canViewUsers ? [{ label: 'Patients', value: patients, icon: Users, action: () => onNavigate('/admin/patients') }] : []),
    ...(canViewDoctors ? [{ label: 'Doctors', value: doctors.length, icon: Activity, action: () => onNavigate('/admin/doctors') }] : []),
    ...(canViewAppointments ? [
      { label: 'Upcoming visits', value: upcomingTotal, icon: CalendarDays, action: () => onNavigate('/admin/appointments') },
      { label: 'Today', value: todayTotal, icon: Clock3, action: () => onNavigate('/admin/appointments') },
    ] : []),
  ]

  return <main className="mx-auto w-full max-w-[1300px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow="Administration"
      title="Clinic overview"
      description="A live snapshot of patients, doctors and the appointment diary."
      actions={canViewAppointments ? <Button className="gap-2" onClick={() => onNavigate('/admin/appointments')}><CalendarDays size={15} aria-hidden="true" /> Open appointments</Button> : undefined}
    />
    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading clinic overview…" /> : null}
    {!loading && !error && <>
      {!hasOverview && <EmptyState title="No overview access" description="Your role does not include access to clinic overview data." />}
      {cards.length > 0 && <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map(card => <button key={card.label} type="button" onClick={card.action} className="rounded-2xl border border-border bg-card p-4 text-left shadow-sm transition hover:border-neutral-400">
          <span className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground"><card.icon size={16} aria-hidden="true" /></span>
          <p className="mt-3 text-2xl font-semibold tracking-tight">{card.value}</p>
          <p className="text-[10px] font-medium uppercase tracking-[0.1em] text-muted-foreground">{card.label}</p>
        </button>)}
      </div>}

      {canViewAppointments && <div className="mt-5 grid gap-5 xl:grid-cols-[minmax(0,1.4fr)_minmax(300px,0.8fr)]">
        <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
          <div className="mb-3 flex items-center justify-between"><div><h2 className="text-base font-semibold">Today’s appointments</h2><p className="text-[11px] text-muted-foreground">{formatDate(new Date().toISOString(), CLINIC_TIMEZONE)}</p></div><Button variant="outline" size="sm" onClick={() => onNavigate('/admin/appointments')}>View all</Button></div>
          {today.length === 0 ? <EmptyState title="Nothing booked today" description="The clinic diary is clear for today." /> : <div className="divide-y divide-border/70">
            {today.map(appointment => <div key={appointment.id} className="flex items-center gap-3 py-2.5 first:pt-0 last:pb-0">
              <div className="w-20 shrink-0 text-[11px] font-semibold">{formatTimeRange(appointment.start_time, appointment.end_time, CLINIC_TIMEZONE)}</div>
              <div className="min-w-0 flex-1"><p className="truncate text-xs font-semibold">{appointment.patient.first_name} {appointment.patient.last_name}</p><p className="truncate text-[10px] text-muted-foreground">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name} · {appointment.reason}</p></div>
              <StatusBadge status={appointment.status} />
            </div>)}
          </div>}
        </section>

        <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
          <h2 className="text-base font-semibold">Next up</h2>
          <p className="text-[11px] text-muted-foreground">The next five confirmed visits.</p>
          {upcoming.length === 0 ? <EmptyState title="No upcoming visits" /> : <div className="mt-3 space-y-2">
            {upcoming.map(appointment => <div key={appointment.id} className="flex items-center gap-2 rounded-xl bg-muted/50 px-3 py-2">
              <div className="min-w-0 flex-1"><p className="truncate text-[11px] font-semibold">{appointment.patient.first_name} {appointment.patient.last_name}</p><p className="truncate text-[10px] text-muted-foreground">{formatDate(appointment.start_time, CLINIC_TIMEZONE, { weekday: 'short', year: undefined })} · {formatTimeRange(appointment.start_time, appointment.end_time, CLINIC_TIMEZONE)}</p></div>
              <span className="shrink-0 text-[9px] text-muted-foreground">Dr. {appointment.doctor.last_name}</span>
            </div>)}
          </div>}
        </section>
      </div>}

      {canViewAudit && audit.length > 0 && <section className="mt-5 rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
        <div className="mb-3 flex items-center justify-between"><h2 className="text-base font-semibold">Recent activity</h2>{canManageAccess && <Button variant="outline" size="sm" onClick={() => onNavigate('/admin/roles')}>Roles & access</Button>}</div>
        <div className="divide-y divide-border/70">{audit.map(entry => <div key={entry.id} className="flex items-center justify-between gap-3 py-2 text-[11px] first:pt-0 last:pb-0"><span className="font-medium">{entry.action}</span><span className="text-muted-foreground">{new Date(entry.created_at).toLocaleString()}</span></div>)}</div>
      </section>}
    </>}
  </main>
}

