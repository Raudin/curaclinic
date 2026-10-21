import { useCallback, useEffect, useState } from 'react'
import { ArrowRight, CalendarDays, CalendarPlus, Clock3, HeartPulse, Plus } from 'lucide-react'
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
  const [totals, setTotals] = useState({ upcoming: 0, past: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [upcomingPage, pastPage] = await Promise.all([
        appointmentsApi.list({ scope: 'upcoming', limit: 5 }),
        appointmentsApi.list({ scope: 'past', limit: 3 }),
      ])
      setUpcoming(upcomingPage.data)
      setPast(pastPage.data)
      setTotals({
        upcoming: upcomingPage.pagination.total,
        past: pastPage.pagination.total,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load your dashboard.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const next = upcoming[0]

  return <main className="@container/page mx-auto w-full max-w-[1440px] px-4 pb-12 pt-6 sm:px-6 sm:pt-9 lg:px-8 lg:pt-10 2xl:px-10">
    <header className="mb-6 flex flex-col justify-between gap-4 sm:mb-7 sm:flex-row sm:items-end">
      <div className="min-w-0">
        <p className="text-xs font-medium text-primary">Your personal health space</p>
        <h1 className="mt-1 text-balance text-[26px] font-semibold leading-tight tracking-[-0.05em] text-foreground sm:text-[38px]">
          {greetingFor(CLINIC_TIMEZONE)}, {user?.first_name ?? 'there'}<span className="text-muted-foreground">.</span>
        </h1>
        <p className="mt-2 max-w-xl text-sm leading-relaxed text-muted-foreground">Your appointments and visits, pulled live from the clinic.</p>
      </div>
      <Button onClick={() => onNavigate('/book-appointment')} className="h-11 w-full gap-2 rounded-xl px-4 text-xs font-semibold sm:h-10 sm:w-auto"><Plus size={15} aria-hidden="true" /> Book appointment</Button>
    </header>

    {error && <div className="mb-5"><ErrorState message={error} onRetry={load} /></div>}
    {loading && !error ? <LoadingState label="Loading your appointments…" /> : null}

    {!loading && !error && <div className="grid min-w-0 grid-cols-1 gap-4 sm:gap-5 @4xl/page:grid-cols-[minmax(0,1.65fr)_minmax(320px,0.95fr)]">
      {/* One 2×2 grid so each row's cards stretch to a shared height; DOM order stays the mobile reading order. */}
      <section aria-label="Next appointment" className="min-w-0 overflow-hidden rounded-3xl border border-border bg-card shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/70 px-4 py-3 sm:px-6">
          <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground"><CalendarDays size={14} aria-hidden="true" /> Next appointment</div>
          {next && <StatusBadge status={next.status} />}
        </div>
        {next ? (() => {
          const parts = dateParts(next.start_time)
          return <div className="@container p-4 sm:p-6">
            <div className="flex flex-col gap-4 @md:flex-row @md:items-center @md:justify-between @md:gap-5">
              <div className="flex min-w-0 items-start gap-3 sm:gap-4">
                <div className="flex w-[60px] shrink-0 flex-col items-center overflow-hidden rounded-2xl border border-border bg-muted/40 sm:w-16">
                  <span className="w-full bg-neutral-950 py-1.5 text-center text-[9px] font-semibold tracking-[0.15em] text-white">{parts.month}</span>
                  <span className="py-1 text-[26px] font-semibold leading-none tracking-[-0.06em]">{parts.day}</span>
                  <span className="pb-2 pt-1 text-[9px] font-medium tracking-wider text-muted-foreground">{parts.weekday}</span>
                </div>
                <div className="min-w-0 flex-1">
                  <h2 className="break-words text-base font-semibold tracking-[-0.03em] sm:text-lg">Dr. {next.doctor.first_name} {next.doctor.last_name}</h2>
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">{next.doctor.specialty}</p>
                  <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
                    <span className="inline-flex items-center gap-1"><Clock3 size={13} aria-hidden="true" /> {formatTimeRange(next.start_time, next.end_time, CLINIC_TIMEZONE)}</span>
                    <span className="hidden size-1 rounded-full bg-border sm:inline-block" aria-hidden="true" />
                    <span>{formatShortDate(next.start_time, CLINIC_TIMEZONE)}</span>
                  </div>
                  <p className="mt-3 line-clamp-2 break-words rounded-lg bg-muted/60 px-3 py-2 text-[11px] leading-relaxed text-foreground">{next.reason}</p>
                </div>
              </div>
              <div className="flex shrink-0 flex-col gap-2 sm:items-end">
                <Button variant="outline" className="h-10 w-full justify-between gap-2 px-4 sm:w-auto sm:justify-center" onClick={() => onNavigate('/my-appointments')}>Manage <ArrowRight size={14} aria-hidden="true" /></Button>
                <p className="text-center text-[10px] text-muted-foreground sm:text-right">Times shown in {CLINIC_TIMEZONE.replace('_', ' ')}</p>
              </div>
            </div>
          </div>
        })() : <div className="p-4 sm:p-6"><EmptyState title="No upcoming appointments" description="Book a visit with one of our doctors and it will appear here." action={<Button className="h-10 gap-2" onClick={() => onNavigate('/book-appointment')}><CalendarPlus size={15} aria-hidden="true" /> Book appointment</Button>} /></div>}
      </section>

      <section aria-label="Recent visits" className="min-w-0 rounded-3xl border border-border bg-card p-4 shadow-sm sm:p-6 @4xl/page:col-start-1 @4xl/page:row-start-2">
        <div className="mb-3 flex items-start justify-between gap-3 sm:mb-4 sm:items-center">
          <div className="min-w-0"><h2 className="text-base font-semibold tracking-tight">Recent visits</h2><p className="mt-0.5 text-xs text-muted-foreground">Your latest completed appointments.</p></div>
          <button type="button" className="shrink-0 rounded py-0.5 text-[11px] font-semibold text-primary hover:underline" onClick={() => onNavigate('/my-appointments')}>View all</button>
        </div>
        {past.length === 0 ? <p className="rounded-xl bg-muted/50 px-4 py-6 text-center text-xs text-muted-foreground">No past visits yet.</p> : <div className="divide-y divide-border/70">
          {past.map(appointment => <div key={appointment.id} className="flex items-center gap-2.5 py-3 first:pt-0 last:pb-0 sm:gap-3">
            <Avatar name={`${appointment.doctor.first_name} ${appointment.doctor.last_name}`} size="sm" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}</p>
              <p className="truncate text-[10px] text-muted-foreground">{appointment.reason}</p>
            </div>
            <div className="shrink-0 text-right"><p className="text-[10px] font-medium">{formatShortDate(appointment.start_time, CLINIC_TIMEZONE)}</p><p className="text-[9px] text-muted-foreground">{formatTime(appointment.start_time, CLINIC_TIMEZONE)}</p></div>
          </div>)}
        </div>}
      </section>

      <section aria-label="Appointment summary" className="@container grid min-w-0 grid-cols-2 gap-2 sm:gap-3 @4xl/page:col-start-2 @4xl/page:row-start-1">
        {[
          { label: 'Upcoming', value: totals.upcoming, context: 'Scheduled visits' },
          { label: 'Past', value: totals.past, context: 'Completed visits' },
        ].map(item => <MetricCard key={item.label} label={item.label} value={item.value} context={item.context} className="h-full" />)}
      </section>

      <section className="flex min-w-0 flex-col rounded-3xl bg-primary p-4 text-primary-foreground shadow-sm sm:p-5 @4xl/page:col-start-2 @4xl/page:row-start-2">
        <div className="mb-3 flex size-9 items-center justify-center rounded-xl bg-primary-foreground/15"><HeartPulse size={17} aria-hidden="true" /></div>
        <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-primary-foreground/70">Need to change a visit?</p>
        <h2 className="mt-1.5 text-base font-semibold tracking-[-0.03em] sm:text-lg">Manage it in one place.</h2>
        <p className="mt-1 text-[11px] leading-relaxed text-primary-foreground/80">Reschedule or cancel from My appointments. Cancelling releases the slot back to the clinic immediately.</p>
        <div className="mt-auto pt-4"><Button className="h-10 w-full bg-card text-card-foreground hover:bg-card/90 sm:w-auto" onClick={() => onNavigate('/my-appointments')}>Open my appointments</Button></div>
      </section>
    </div>}
  </main>
}
