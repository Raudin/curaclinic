import { useCallback, useEffect, useState } from 'react'
import { ArrowRight, CalendarDays, Clock3, Plus } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { MetricCard } from '@/components/arc/metric-card/metric-card'
import { Button } from '@/components/ui/button'
import { PageHeader, ErrorState, LoadingState, StatusBadge } from '@/components/shared/page'
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

  return <main className="@container/page mx-auto w-full max-w-[1200px] px-4 pb-12 pt-6 sm:px-6 sm:pt-9 lg:px-8 lg:pt-10 2xl:px-10">
    <PageHeader
      title={greetingFor(CLINIC_TIMEZONE) + ', ' + (user?.first_name ?? 'there')}
      description={new Intl.DateTimeFormat('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', timeZone: CLINIC_TIMEZONE }).format(new Date())}
      actions={<Button className="h-11 gap-2" onClick={() => onNavigate('/book-appointment')}><Plus size={16} aria-hidden="true" /> Book appointment</Button>}
    />

    {error && <div className="mb-5"><ErrorState message={error} onRetry={load} /></div>}
    {loading && !error ? <LoadingState label="Loading your appointments…" /> : null}

    {!loading && !error && <div className="grid min-w-0 grid-cols-1 gap-4 sm:gap-5 @4xl/page:grid-cols-[minmax(0,1.65fr)_minmax(320px,0.95fr)]">
      <section aria-label="Next appointment" className="min-w-0 overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
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
                  <span className="w-full bg-primary py-1.5 text-center text-[9px] font-semibold tracking-[0.15em] text-white">{parts.month}</span>
                  <span className="py-1 text-[26px] font-semibold leading-none tracking-[-0.035em]">{parts.day}</span>
                  <span className="pb-2 pt-1 text-[9px] font-medium tracking-wider text-muted-foreground">{parts.weekday}</span>
                </div>
                <div className="min-w-0 flex-1">
                  <h2 className="break-words text-base font-semibold tracking-[-0.03em] sm:text-lg">Dr. {next.doctor.first_name} {next.doctor.last_name}</h2>
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">{next.doctor.specialty}</p>
                  <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
                    <span className="inline-flex items-center gap-1"><Clock3 size={13} aria-hidden="true" /> {formatTimeRange(next.start_time, next.end_time, CLINIC_TIMEZONE)}</span>
                    <span className="hidden size-1 rounded-full bg-border sm:inline-block" aria-hidden="true" />
                    <span>{formatShortDate(next.start_time, CLINIC_TIMEZONE)}</span>
                  </div>
                  <p className="mt-3 line-clamp-2 break-words rounded-lg bg-muted/60 px-3 py-2 text-sm leading-relaxed text-foreground">{next.reason}</p>
                </div>
              </div>
              <div className="flex shrink-0 flex-col gap-2 sm:items-end">
                <Button variant="outline" className="h-10 w-full justify-between gap-2 px-4 sm:w-auto sm:justify-center" onClick={() => onNavigate('/my-appointments')}>Manage visit <ArrowRight size={14} aria-hidden="true" /></Button>
                <p className="text-center text-xs text-muted-foreground sm:text-right">Times shown in {CLINIC_TIMEZONE.replace('_', ' ')}</p>
              </div>
            </div>
          </div>
        })() : <div className="px-4 py-8 sm:px-6"><h2 className="text-lg font-semibold">No upcoming appointments</h2><p className="mt-1 text-sm text-muted-foreground">Your next confirmed visit will appear here.</p></div>}
      </section>

      <section aria-label="Recent visits" className="min-w-0 rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-6 @4xl/page:col-span-2 @4xl/page:row-start-2">
        <div className="mb-3 flex items-start justify-between gap-3 sm:mb-4 sm:items-center">
          <div className="min-w-0"><h2 className="text-base font-semibold tracking-tight">Recent visits</h2><p className="mt-0.5 text-xs text-muted-foreground">Your latest completed appointments.</p></div>
          <button type="button" className="min-h-10 shrink-0 rounded px-1 text-sm font-semibold text-primary hover:underline" onClick={() => onNavigate('/my-appointments?tab=past')}>View past visits</button>
        </div>
        {past.length === 0 ? <p className="rounded-xl bg-muted/50 px-4 py-6 text-center text-xs text-muted-foreground">No past visits yet.</p> : <div className="divide-y divide-border/70">
          {past.map(appointment => <div key={appointment.id} className="flex items-center gap-2.5 py-3 first:pt-0 last:pb-0 sm:gap-3">
            <Avatar name={`${appointment.doctor.first_name} ${appointment.doctor.last_name}`} size="sm" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}</p>
              <p className="truncate text-xs text-muted-foreground">{appointment.reason}</p>
            </div>
            <div className="shrink-0 text-right"><p className="text-xs font-medium">{formatShortDate(appointment.start_time, CLINIC_TIMEZONE)}</p><p className="text-xs text-muted-foreground">{formatTime(appointment.start_time, CLINIC_TIMEZONE)}</p></div>
          </div>)}
        </div>}
      </section>

      <section aria-label="Appointment summary" className="@container grid min-w-0 grid-cols-2 gap-2 sm:gap-3 @4xl/page:col-start-2 @4xl/page:row-start-1">
        {[
          { label: 'Upcoming', value: totals.upcoming, context: 'Scheduled visits' },
          { label: 'Past', value: totals.past, context: 'Completed visits' },
        ].map(item => <MetricCard key={item.label} label={item.label} value={item.value} context={item.context} className="h-full" />)}
      </section>

    </div>}
  </main>
}
