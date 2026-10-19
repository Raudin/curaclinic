import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowRight, CalendarDays, CalendarPlus, Clock3, HeartPulse, Plus, Stethoscope } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { MetricCard } from '@/components/arc/metric-card/metric-card'
import { Button } from '@/components/ui/button'
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '@/components/shared/page'
import { useAuth } from '@/lib/auth-context'
import { appointmentsApi } from '@/lib/api-clinic'
import type { Appointment } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, formatShortDate, formatTime, formatTimeRange } from '@/lib/time'

const CLINIC_TIMEZONE = import.meta.env.VITE_CLINIC_TIMEZONE || DEFAULT_TIMEZONE

function greetingFor(timeZone: string): string {
  const hour = Number(new Intl.DateTimeFormat('en-GB', { timeZone, hour: '2-digit', hour12: false }).format(new Date()))
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
}

function dateParts(iso: string) {
  const weekday = new Intl.DateTimeFormat('en-GB', { timeZone: CLINIC_TIMEZONE, weekday: 'short' }).format(new Date(iso)).toUpperCase()
  const day = new Intl.DateTimeFormat('en-GB', { timeZone: CLINIC_TIMEZONE, day: 'numeric' }).format(new Date(iso))
  const month = new Intl.DateTimeFormat('en-GB', { timeZone: CLINIC_TIMEZONE, month: 'short' }).format(new Date(iso)).toUpperCase()
  return { weekday, day, month }
}

export function PatientDashboard({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { user } = useAuth()
  const [upcoming, setUpcoming] = useState<Appointment[]>([])
  const [past, setPast] = useState<Appointment[]>([])
  const [totals, setTotals] = useState({ upcoming: 0, past: 0, cancelled: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [upcomingPage, pastPage, cancelledPage] = await Promise.all([
        appointmentsApi.list({ scope: 'upcoming', limit: 5 }),
        appointmentsApi.list({ scope: 'past', limit: 3 }),
        appointmentsApi.list({ scope: 'cancelled', limit: 1 }),
      ])
      setUpcoming(upcomingPage.data)
      setPast(pastPage.data)
      setTotals({
        upcoming: upcomingPage.pagination.total,
        past: pastPage.pagination.total,
        cancelled: cancelledPage.pagination.total,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load your dashboard.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const next = upcoming[0]
  const careTeam = useMemo(() => {
    const doctors = new Map<string, Appointment['doctor']>()
    for (const appointment of upcoming) doctors.set(appointment.doctor.id, appointment.doctor)
    return [...doctors.values()]
  }, [upcoming])

  return <main className="mx-auto max-w-[1440px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10 lg:pt-10">
    <header className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <p className="text-xs font-medium text-primary">Your personal health space</p>
        <h1 className="mt-1 text-[30px] font-semibold leading-tight tracking-[-0.05em] text-foreground sm:text-[38px]">
          {greetingFor(CLINIC_TIMEZONE)}, {user?.first_name ?? 'there'}<span className="text-muted-foreground">.</span>
        </h1>
        <p className="mt-2 max-w-xl text-sm leading-relaxed text-muted-foreground">Your appointments and care team, pulled live from the clinic.</p>
      </div>
      <Button onClick={() => onNavigate('/book-appointment')} className="h-10 gap-2 rounded-xl px-4 text-xs font-semibold"><Plus size={15} aria-hidden="true" /> Book appointment</Button>
    </header>

    {error && <div className="mb-5"><ErrorState message={error} onRetry={load} /></div>}
    {loading && !error ? <LoadingState label="Loading your appointments…" /> : null}

    {!loading && !error && <div className="grid gap-5 xl:grid-cols-[minmax(0,1.65fr)_minmax(300px,0.95fr)]">
      <div className="flex min-w-0 flex-col gap-5">
        <section aria-label="Next appointment" className="overflow-hidden rounded-3xl border border-border bg-card shadow-sm">
          <div className="flex items-center justify-between border-b border-border/70 px-5 py-3 sm:px-6">
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground"><CalendarDays size={14} aria-hidden="true" /> Next appointment</div>
            {next && <StatusBadge status={next.status} />}
          </div>
          {next ? (() => {
            const parts = dateParts(next.start_time)
            return <div className="flex flex-col gap-5 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-6">
              <div className="flex items-start gap-4">
                <div className="flex w-[64px] shrink-0 flex-col items-center overflow-hidden rounded-2xl border border-border bg-muted/40">
                  <span className="w-full bg-neutral-950 py-1.5 text-center text-[9px] font-semibold tracking-[0.15em] text-white">{parts.month}</span>
                  <span className="py-1 text-[26px] font-semibold leading-none tracking-[-0.06em]">{parts.day}</span>
                  <span className="pb-2 pt-1 text-[9px] font-medium tracking-wider text-muted-foreground">{parts.weekday}</span>
                </div>
                <div className="min-w-0">
                  <h2 className="text-lg font-semibold tracking-[-0.03em]">Dr. {next.doctor.first_name} {next.doctor.last_name}</h2>
                  <p className="mt-0.5 text-xs text-muted-foreground">{next.doctor.specialty}</p>
                  <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
                    <span className="inline-flex items-center gap-1"><Clock3 size={13} aria-hidden="true" /> {formatTimeRange(next.start_time, next.end_time, CLINIC_TIMEZONE)}</span>
                    <span className="hidden size-1 rounded-full bg-border sm:inline-block" aria-hidden="true" />
                    <span>{formatShortDate(next.start_time, CLINIC_TIMEZONE)}</span>
                  </div>
                  <p className="mt-3 line-clamp-2 rounded-lg bg-muted/60 px-3 py-2 text-[11px] leading-relaxed text-foreground">{next.reason}</p>
                </div>
              </div>
              <div className="flex shrink-0 flex-col gap-2 sm:items-end">
                <Button variant="outline" className="gap-2" onClick={() => onNavigate('/my-appointments')}>Manage <ArrowRight size={14} aria-hidden="true" /></Button>
                <p className="text-[10px] text-muted-foreground">Times shown in {CLINIC_TIMEZONE.replace('_', ' ')}</p>
              </div>
            </div>
          })() : <div className="p-5 sm:p-6"><EmptyState title="No upcoming appointments" description="Book a visit with one of our doctors and it will appear here." action={<Button className="gap-2" onClick={() => onNavigate('/book-appointment')}><CalendarPlus size={15} aria-hidden="true" /> Book appointment</Button>} /></div>}
        </section>

        <section aria-label="Recent visits" className="rounded-3xl border border-border bg-card p-5 shadow-sm sm:p-6">
          <div className="mb-4 flex items-center justify-between">
            <div><h2 className="text-base font-semibold tracking-tight">Recent visits</h2><p className="mt-0.5 text-xs text-muted-foreground">Your latest completed appointments.</p></div>
            <button type="button" className="text-[11px] font-semibold text-primary hover:underline" onClick={() => onNavigate('/my-appointments')}>View all</button>
          </div>
          {past.length === 0 ? <p className="rounded-xl bg-muted/50 px-4 py-6 text-center text-xs text-muted-foreground">No past visits yet.</p> : <div className="divide-y divide-border/70">
            {past.map(appointment => <div key={appointment.id} className="flex items-center gap-3 py-3 first:pt-0 last:pb-0">
              <Avatar name={`${appointment.doctor.first_name} ${appointment.doctor.last_name}`} size="sm" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-semibold">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}</p>
                <p className="truncate text-[10px] text-muted-foreground">{appointment.reason}</p>
              </div>
              <div className="shrink-0 text-right"><p className="text-[10px] font-medium">{formatShortDate(appointment.start_time, CLINIC_TIMEZONE)}</p><p className="text-[9px] text-muted-foreground">{formatTime(appointment.start_time, CLINIC_TIMEZONE)}</p></div>
            </div>)}
          </div>}
        </section>
      </div>

      <div className="flex flex-col gap-5">
        <section aria-label="Appointment summary" className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {[
            { label: 'Upcoming', value: totals.upcoming, context: 'Scheduled visits' },
            { label: 'Past', value: totals.past, context: 'Completed visits' },
            { label: 'Cancelled', value: totals.cancelled, context: 'Cancelled visits' },
          ].map(item => <MetricCard key={item.label} label={item.label} value={item.value} context={item.context} />)}
        </section>

        <section aria-label="Care team" className="rounded-3xl border border-border bg-card p-5 shadow-sm">
          <div className="mb-3 flex items-center gap-2"><Stethoscope size={16} className="text-primary" aria-hidden="true" /><h2 className="text-base font-semibold tracking-tight">Your care team</h2></div>
          {careTeam.length === 0 ? <p className="text-xs leading-relaxed text-muted-foreground">Your care team will appear here after you book your first appointment.</p> : <div className="space-y-2">
            {careTeam.map(doctor => <div key={doctor.id} className="flex items-center gap-3 rounded-xl bg-muted/40 px-3 py-2.5">
              <Avatar name={`${doctor.first_name} ${doctor.last_name}`} size="sm" />
              <div className="min-w-0"><p className="truncate text-xs font-semibold">Dr. {doctor.first_name} {doctor.last_name}</p><p className="truncate text-[10px] text-muted-foreground">{doctor.specialty}</p></div>
            </div>)}
          </div>}
          <Button variant="outline" className="mt-3 w-full gap-2" onClick={() => onNavigate('/book-appointment')}><CalendarPlus size={14} aria-hidden="true" /> Book another visit</Button>
        </section>

        <section className="rounded-3xl bg-neutral-950 p-5 text-white shadow-sm">
          <div className="mb-3 flex size-9 items-center justify-center rounded-xl bg-white/10"><HeartPulse size={17} aria-hidden="true" /></div>
          <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-white/45">Need to change a visit?</p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-[-0.03em]">Manage it in one place.</h2>
          <p className="mt-1 text-[11px] leading-relaxed text-white/55">Reschedule or cancel from My appointments. Cancelling releases the slot back to the clinic immediately.</p>
          <Button className="mt-4 bg-white text-neutral-950 hover:bg-neutral-100" onClick={() => onNavigate('/my-appointments')}>Open my appointments</Button>
        </section>
      </div>
    </div>}
  </main>
}
