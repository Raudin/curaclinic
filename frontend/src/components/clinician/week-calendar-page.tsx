import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { AlertCircle, CalendarDays, ChevronLeft, ChevronRight, Clock3, LoaderCircle, Mail, RotateCcw, XCircle } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/arc/textarea/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ErrorState, PageHeader, StatusBadge } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import { scheduleApi } from '@/lib/api-admin'
import type { Appointment, Doctor, ScheduleBlock, WorkingHour } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, addDays, clockToMinutes, dateFromKey, dateKey, dateKeyInZone, formatDate, formatTime, formatTimeRange, zonedDayBoundsFromKey } from '@/lib/time'

const HOUR_HEIGHT = 56
const DAY_MINUTES = 24 * 60
const HOURS = Array.from({ length: 24 }, (_, hour) => hour)

type Span = { top: number; bottom: number }
type Placed<T> = { item: T; span: Span; column: number; columns: number }

/** Monday of the week containing the given date key. */
function weekStartKey(key: string): string {
  const date = dateFromKey(key)
  const offset = (date.getDay() + 6) % 7
  return dateKey(addDays(date, -offset))
}

/** The minutes of a day, in the doctor's zone, that an interval covers, or null when it misses the day. */
function spanOnDay(startIso: string, endIso: string, day: { from: string; to: string }, timeZone: string): Span | null {
  const start = Math.max(new Date(startIso).getTime(), new Date(day.from).getTime())
  const end = Math.min(new Date(endIso).getTime(), new Date(day.to).getTime())
  if (end <= start) return null
  const top = clockToMinutes(formatTime(new Date(start).toISOString(), timeZone))
  const bottom = end >= new Date(day.to).getTime() ? DAY_MINUTES : clockToMinutes(formatTime(new Date(end).toISOString(), timeZone))
  return { top, bottom: Math.max(bottom, top + 1) }
}

/** Overlapping events share the day column side by side, the way Google Calendar lays out a busy hour. */
function layout<T>(entries: Array<{ item: T; span: Span }>): Placed<T>[] {
  const sorted = [...entries].sort((a, b) => a.span.top - b.span.top || b.span.bottom - a.span.bottom)
  const placed: Placed<T>[] = []
  let cluster: Placed<T>[] = []
  let columnEnds: number[] = []
  let clusterEnd = -1
  const flush = () => { for (const entry of cluster) entry.columns = columnEnds.length; cluster = []; columnEnds = [] }
  for (const entry of sorted) {
    if (entry.span.top >= clusterEnd) flush()
    let column = columnEnds.findIndex(end => end <= entry.span.top)
    if (column === -1) { column = columnEnds.length; columnEnds.push(entry.span.bottom) } else columnEnds[column] = entry.span.bottom
    const next = { ...entry, column, columns: 1 }
    cluster.push(next)
    placed.push(next)
    clusterEnd = Math.max(clusterEnd, entry.span.bottom)
  }
  flush()
  return placed
}

async function listWeek(from: string, to: string): Promise<Appointment[]> {
  const items: Appointment[] = []
  for (let offset = 0; ; offset += 100) {
    const page = await appointmentsApi.list({ scope: 'all', from, to, order: 'asc', limit: 100, offset })
    items.push(...page.data)
    if (page.data.length < 100 || items.length >= page.pagination.total) return items
  }
}

export function WeekCalendarPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { user, can } = useAuth()
  const { notify } = useToast()
  const canReschedule = can(PERMISSIONS.doctorAppointmentsReschedule)
  const canCancel = can(PERMISSIONS.doctorAppointmentsCancel)
  const canViewSchedule = can(PERMISSIONS.doctorScheduleView)
  const [doctor, setDoctor] = useState<Doctor | null>(null)
  const timeZone = doctor?.timezone || DEFAULT_TIMEZONE
  const [now, setNow] = useState(() => Date.now())
  const todayKey = dateKeyInZone(new Date(now), timeZone)
  const [weekKey, setWeekKey] = useState(() => weekStartKey(dateKeyInZone(new Date(), DEFAULT_TIMEZONE)))
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [workingHours, setWorkingHours] = useState<WorkingHour[]>([])
  const [blocks, setBlocks] = useState<ScheduleBlock[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<Appointment | null>(null)
  const [cancelMode, setCancelMode] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const [cancelError, setCancelError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)
  const scrolledFor = useRef<string | null>(null)
  const requestId = useRef(0)

  useEffect(() => {
    if (!user) return
    let cancelled = false
    doctorsApi.get(user.id).then(profile => { if (!cancelled) setDoctor(profile) }).catch(() => undefined)
    return () => { cancelled = true }
  }, [user])

  useEffect(() => {
    if (!canViewSchedule) return
    scheduleApi.mine().then(schedule => { setWorkingHours(schedule.working_hours); setBlocks(schedule.blocks) }).catch(() => undefined)
  }, [canViewSchedule])

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 60_000)
    return () => window.clearInterval(timer)
  }, [])

  const days = useMemo(() => Array.from({ length: 7 }, (_, index) => {
    const key = dateKey(addDays(dateFromKey(weekKey), index))
    return { key, bounds: zonedDayBoundsFromKey(key, timeZone), weekday: dateFromKey(key).getDay() }
  }), [timeZone, weekKey])

  const load = useCallback(async () => {
    const current = ++requestId.current
    setLoading(true)
    setError(null)
    try {
      const items = await listWeek(days[0]!.bounds.from, days[6]!.bounds.to)
      if (current === requestId.current) setAppointments(items)
    } catch (err) {
      if (current === requestId.current) setError(err instanceof Error ? err.message : 'Unable to load your appointments.')
    } finally {
      if (current === requestId.current) setLoading(false)
    }
  }, [days])

  useEffect(() => { void load() }, [load])

  const columns = useMemo(() => days.map(day => {
    const events = layout(appointments.flatMap(item => {
      if (dateKeyInZone(new Date(item.start_time), timeZone) !== day.key) return []
      const span = spanOnDay(item.start_time, item.end_time, day.bounds, timeZone)
      return span ? [{ item, span }] : []
    }))
    const blocked = blocks.flatMap(block => {
      const span = spanOnDay(block.start_time, block.end_time, day.bounds, timeZone)
      return span ? [{ block, span }] : []
    })
    const hours = workingHours.filter(hour => hour.weekday === day.weekday).map(hour => ({ top: clockToMinutes(hour.start_time), bottom: clockToMinutes(hour.end_time) }))
    return { ...day, events, blocked, hours }
  }), [appointments, blocks, days, timeZone, workingHours])

  // Open the week scrolled to the start of the working day (or the first visit, or now), the way a calendar app lands.
  useLayoutEffect(() => {
    const scroller = scrollRef.current
    if (!scroller || loading || scrolledFor.current === weekKey) return
    scrolledFor.current = weekKey
    const candidates = columns.flatMap(column => [...column.hours.map(hour => hour.top), ...column.events.map(event => event.span.top)])
    const nowMinutes = clockToMinutes(formatTime(new Date(now).toISOString(), timeZone))
    const target = candidates.length ? Math.min(...candidates) : columns.some(column => column.key === todayKey) ? nowMinutes - 60 : 8 * 60
    scroller.scrollTop = Math.max(0, (target - 30) / 60 * HOUR_HEIGHT)
  }, [columns, loading, now, timeZone, todayKey, weekKey])

  const weekLabel = (() => {
    const first = days[0]!.key, last = days[6]!.key
    const label = (key: string, options: Intl.DateTimeFormatOptions) => formatDate(`${key}T12:00:00Z`, 'UTC', { weekday: undefined, ...options })
    if (first.slice(0, 7) === last.slice(0, 7)) return `${label(first, { month: undefined, year: undefined })} – ${label(last, { month: 'short' })}`
    return `${label(first, { month: 'short', year: first.slice(0, 4) === last.slice(0, 4) ? undefined : 'numeric' })} – ${label(last, { month: 'short' })}`
  })()
  const confirmed = appointments.filter(item => item.status === 'confirmed').length
  const nowMinutes = clockToMinutes(formatTime(new Date(now).toISOString(), timeZone))

  function openDetails(appointment: Appointment) {
    setSelected(appointment)
    setCancelMode(false)
    setCancelReason('')
    setCancelError(null)
  }

  async function confirmCancel() {
    if (!selected) return
    if (!cancelReason.trim()) { setCancelError('Add a reason for cancelling this visit.'); return }
    setSubmitting(true)
    setCancelError(null)
    try {
      await appointmentsApi.cancel(selected.id, cancelReason.trim())
      notify('The appointment has been cancelled.')
      setSelected(null)
      await load()
    } catch (err) {
      setCancelError(err instanceof ApiError ? err.message : 'Unable to cancel the appointment.')
    } finally {
      setSubmitting(false)
    }
  }

  return <main className="mx-auto w-full max-w-[1440px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow="Calendar"
      title="My appointments"
      description={doctor ? `Dr. ${doctor.first_name} ${doctor.last_name} · ${doctor.specialty}` : 'Your week at a glance.'}
      actions={<Button variant="outline" className="gap-2" onClick={() => onNavigate('/daily-agenda')}><CalendarDays size={15} aria-hidden="true" /> Daily agenda</Button>}
    />

    <section className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-3 py-2.5 sm:px-4">
        <div className="flex items-center gap-1.5">
          <Button variant="outline" size="sm" onClick={() => setWeekKey(weekStartKey(todayKey))}>Today</Button>
          <Button variant="ghost" size="icon" aria-label="Previous week" onClick={() => setWeekKey(key => dateKey(addDays(dateFromKey(key), -7)))}><ChevronLeft size={16} aria-hidden="true" /></Button>
          <Button variant="ghost" size="icon" aria-label="Next week" onClick={() => setWeekKey(key => dateKey(addDays(dateFromKey(key), 7)))}><ChevronRight size={16} aria-hidden="true" /></Button>
          <h2 className="ml-1 text-sm font-semibold tracking-tight sm:text-base" aria-live="polite">{weekLabel}</h2>
          {loading && <LoaderCircle size={14} className="ml-1 animate-spin text-muted-foreground" aria-label="Loading" />}
        </div>
        <p className="text-[11px] text-muted-foreground">{confirmed} confirmed this week · {timeZone.replace('_', ' ')}</p>
      </div>

      {error ? <div className="p-4"><ErrorState message={error} onRetry={load} /></div> : <div ref={scrollRef} className="max-h-[calc(100svh-260px)] min-h-[420px] overflow-auto overscroll-contain">
        <div className="min-w-[760px]">
          {/* Day headers stay pinned while the hours scroll beneath them. */}
          <div className="sticky top-0 z-20 grid grid-cols-[56px_repeat(7,minmax(0,1fr))] border-b border-border bg-card">
            <div className="sticky left-0 z-10 bg-card" />
            {columns.map(column => {
              const isToday = column.key === todayKey
              const date = dateFromKey(column.key)
              return <div key={column.key} className="flex flex-col items-center gap-0.5 border-l border-border/70 py-2" aria-current={isToday ? 'date' : undefined}>
                <span className={`text-[10px] font-semibold uppercase tracking-[0.1em] ${isToday ? 'text-primary' : 'text-muted-foreground'}`}>{new Intl.DateTimeFormat('en-GB', { weekday: 'short' }).format(date)}</span>
                <span className={`flex size-8 items-center justify-center rounded-full text-base font-semibold ${isToday ? 'bg-primary text-primary-foreground' : 'text-foreground'}`}>{date.getDate()}</span>
              </div>
            })}
          </div>

          <div className="relative grid grid-cols-[56px_repeat(7,minmax(0,1fr))]" style={{ height: HOUR_HEIGHT * 24 }}>
            <div className="sticky left-0 z-10 bg-card">
              {HOURS.map(hour => <div key={hour} className="relative" style={{ height: HOUR_HEIGHT }}>
                {hour > 0 && <span className="absolute -top-2 right-2 text-[10px] tabular-nums text-muted-foreground">{String(hour).padStart(2, '0')}:00</span>}
              </div>)}
            </div>

            {columns.map(column => {
              const isToday = column.key === todayKey
              return <div key={column.key} className="relative border-l border-border/70 bg-muted/35">
                {/* Working hours read as open paper; everything outside them stays tinted, like off-hours in a calendar app. */}
                {column.hours.map((hour, index) => <div key={index} className="absolute inset-x-0 bg-card" style={{ top: hour.top / 60 * HOUR_HEIGHT, height: (hour.bottom - hour.top) / 60 * HOUR_HEIGHT }} aria-hidden="true" />)}
                {HOURS.map(hour => <div key={hour} className="pointer-events-none absolute inset-x-0 border-t border-border/60" style={{ top: hour * HOUR_HEIGHT }} aria-hidden="true" />)}

                {column.blocked.map(({ block, span }) => <div
                  key={block.id}
                  className="absolute inset-x-0.5 overflow-hidden rounded-md border border-border bg-[repeating-linear-gradient(135deg,var(--muted)_0_6px,transparent_6px_12px)] px-1.5 py-1 text-[10px] text-muted-foreground"
                  style={{ top: span.top / 60 * HOUR_HEIGHT, height: (span.bottom - span.top) / 60 * HOUR_HEIGHT }}
                  title={`Blocked${block.reason ? ` · ${block.reason}` : ''}`}
                ><span className="font-semibold">Blocked</span>{block.reason ? ` · ${block.reason}` : ''}</div>)}

                {column.events.map(({ item, span, column: lane, columns: lanes }) => {
                  const height = (span.bottom - span.top) / 60 * HOUR_HEIGHT
                  const cancelled = item.status === 'cancelled'
                  const past = new Date(item.end_time).getTime() < now
                  const name = `${item.patient.first_name} ${item.patient.last_name}`
                  return <button
                    key={item.id}
                    type="button"
                    onClick={() => openDetails(item)}
                    className={`absolute z-[1] flex flex-col overflow-hidden rounded-md border-l-[3px] px-1.5 py-1 text-left text-[11px] leading-tight shadow-sm outline-none transition hover:z-[2] hover:shadow-md focus-visible:z-[2] focus-visible:ring-2 focus-visible:ring-ring ${cancelled ? 'border-muted-foreground/40 bg-muted text-muted-foreground' : 'border-primary bg-[color-mix(in_oklab,var(--primary)_16%,var(--card))] text-foreground'} ${past && !cancelled ? 'opacity-70' : ''}`}
                    style={{ top: span.top / 60 * HOUR_HEIGHT + 1, height: Math.max(height - 2, 18), left: `calc(${lane / lanes * 100}% + 2px)`, width: `calc(${100 / lanes}% - 4px)` }}
                    aria-label={`${name}, ${formatTimeRange(item.start_time, item.end_time, timeZone)}${cancelled ? ', cancelled' : ''}`}
                  >
                    {height < 36
                      ? <span className={`truncate ${cancelled ? 'line-through' : ''}`}><span className="font-semibold">{name}</span> · {formatTime(item.start_time, timeZone)}</span>
                      : <>
                        <span className={`truncate font-semibold ${cancelled ? 'line-through' : ''}`}>{name}</span>
                        <span className="truncate text-[10px] text-muted-foreground">{formatTimeRange(item.start_time, item.end_time, timeZone)}</span>
                        {height >= 64 && <span className="mt-0.5 line-clamp-2 text-[10px] text-muted-foreground">{item.reason}</span>}
                      </>}
                  </button>
                })}

                {isToday && <div className="pointer-events-none absolute inset-x-0 z-[3] flex items-center" style={{ top: nowMinutes / 60 * HOUR_HEIGHT }} aria-hidden="true">
                  <span className="-ml-1 size-2.5 rounded-full bg-destructive" />
                  <span className="h-0.5 flex-1 bg-destructive" />
                </div>}
              </div>
            })}
          </div>
        </div>
      </div>}
    </section>

    <Dialog open={Boolean(selected)} onOpenChange={open => { if (!open && !submitting) setSelected(null) }}>
      <DialogContent>
        {selected && <>
          <DialogHeader>
            <DialogTitle>{cancelMode ? 'Cancel this visit?' : `${selected.patient.first_name} ${selected.patient.last_name}`}</DialogTitle>
            <DialogDescription>{formatDate(selected.start_time, timeZone, { year: undefined })} · {formatTimeRange(selected.start_time, selected.end_time, timeZone)}</DialogDescription>
          </DialogHeader>
          {cancelMode ? <>
            {cancelError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{cancelError}</div>}
            <Textarea label="Reason" value={cancelReason} onChange={event => setCancelReason(event.target.value)} rows={3} required placeholder="Reason for cancelling" />
            <DialogFooter>
              <Button variant="outline" onClick={() => setCancelMode(false)} disabled={submitting}>Keep visit</Button>
              <Button variant="destructive" onClick={() => void confirmCancel()} disabled={submitting}>{submitting && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Cancel visit</Button>
            </DialogFooter>
          </> : <>
            <div className="flex items-start gap-3 rounded-xl bg-muted/50 p-3">
              <Avatar name={`${selected.patient.first_name} ${selected.patient.last_name}`} size="md" />
              <div className="min-w-0 flex-1 space-y-1 text-xs">
                <StatusBadge status={selected.status} />
                <p className="flex items-center gap-1.5 text-muted-foreground"><Clock3 size={12} aria-hidden="true" /> {formatTimeRange(selected.start_time, selected.end_time, timeZone)}</p>
                <p className="flex items-center gap-1.5 break-all text-muted-foreground"><Mail size={12} className="shrink-0" aria-hidden="true" /> {selected.patient.email}</p>
              </div>
            </div>
            <div>
              <p className="text-[10px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">Reason for visit</p>
              <p className="mt-1 text-sm leading-relaxed">{selected.reason}</p>
            </div>
            {selected.status === 'confirmed' && new Date(selected.start_time).getTime() > now && (canReschedule || canCancel) && <DialogFooter>
              {canCancel && <Button variant="outline" className="gap-1.5 text-destructive hover:text-destructive" onClick={() => setCancelMode(true)}><XCircle size={13} aria-hidden="true" /> Cancel visit</Button>}
              {canReschedule && <Button className="gap-1.5" onClick={() => onNavigate(`/book-appointment?reschedule=${selected.id}`)}><RotateCcw size={13} aria-hidden="true" /> Reschedule</Button>}
            </DialogFooter>}
          </>}
        </>}
      </DialogContent>
    </Dialog>
  </main>
}
