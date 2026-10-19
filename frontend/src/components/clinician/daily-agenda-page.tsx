import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, CalendarDays, ChevronLeft, ChevronRight, Clock3, LoaderCircle, RotateCcw, XCircle } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/arc/textarea/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState, ErrorState, LoadingState, PageHeader, StatusBadge } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import type { Appointment, Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, addDays, formatDate, formatTime, formatTimeRange, zonedDayBounds } from '@/lib/time'

export function DailyAgendaPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { user, can } = useAuth()
  const { notify } = useToast()
  const canReschedule = can(PERMISSIONS.doctorAppointmentsReschedule)
  const canCancel = can(PERMISSIONS.doctorAppointmentsCancel)
  const [doctor, setDoctor] = useState<Doctor | null>(null)
  const [selectedDate, setSelectedDate] = useState(() => new Date())
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cancelling, setCancelling] = useState<Appointment | null>(null)
  const [reason, setReason] = useState('')
  const [cancelError, setCancelError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const timeZone = doctor?.timezone || DEFAULT_TIMEZONE

  useEffect(() => {
    if (!user) return
    let cancelled = false
    doctorsApi.get(user.id).then(profile => { if (!cancelled) setDoctor(profile) }).catch(() => undefined)
    return () => { cancelled = true }
  }, [user])

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const bounds = zonedDayBounds(selectedDate, timeZone)
      const page = await appointmentsApi.list({ scope: 'all', from: bounds.from, to: bounds.to, order: 'asc', limit: 100 })
      setAppointments(page.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load the agenda.')
    } finally {
      setLoading(false)
    }
  }, [selectedDate, timeZone])

  useEffect(() => { void load() }, [load])

  async function confirmCancel() {
    if (!cancelling) return
    if (!reason.trim()) {
      setCancelError('Add a reason for cancelling this visit.')
      return
    }
    setSubmitting(true)
    setCancelError(null)
    try {
      await appointmentsApi.cancel(cancelling.id, reason.trim())
      notify('The appointment has been cancelled.')
      setCancelling(null)
      await load()
    } catch (err) {
      setCancelError(err instanceof ApiError ? err.message : 'Unable to cancel the appointment.')
    } finally {
      setSubmitting(false)
    }
  }

  const confirmed = appointments.filter(item => item.status === 'confirmed').length
  const cancelled = appointments.filter(item => item.status === 'cancelled').length

  return <main className="mx-auto w-full max-w-[1200px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow="Clinic day"
      title="Daily agenda"
      description={doctor ? `Dr. ${doctor.first_name} ${doctor.last_name} · ${doctor.specialty}` : 'Your appointments for the selected day.'}
      actions={<Button variant="outline" className="gap-2" onClick={() => onNavigate('/schedule-availability')}><CalendarDays size={15} aria-hidden="true" /> Manage availability</Button>}
    />

    <section className="mb-5 flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-border bg-card p-3 shadow-sm">
      <div className="flex items-center gap-1.5">
        <Button variant="outline" size="icon" aria-label="Previous day" onClick={() => setSelectedDate(current => addDays(current, -1))}><ChevronLeft size={16} aria-hidden="true" /></Button>
        <div className="min-w-[200px] text-center"><p className="text-sm font-semibold">{formatDate(selectedDate.toISOString(), timeZone)}</p><p className="text-[10px] text-muted-foreground">Times in {timeZone.replace('_', ' ')}</p></div>
        <Button variant="outline" size="icon" aria-label="Next day" onClick={() => setSelectedDate(current => addDays(current, 1))}><ChevronRight size={16} aria-hidden="true" /></Button>
        <Button variant="secondary" size="sm" onClick={() => setSelectedDate(new Date())}>Today</Button>
      </div>
      <div className="flex gap-2 text-center text-[10px] text-muted-foreground">
        <span className="rounded-lg bg-muted px-2.5 py-1.5"><strong className="block text-sm text-foreground">{confirmed}</strong>Confirmed</span>
        <span className="rounded-lg bg-muted px-2.5 py-1.5"><strong className="block text-sm text-foreground">{cancelled}</strong>Cancelled</span>
      </div>
    </section>

    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading the clinic day…" /> : null}
    {!loading && !error && appointments.length === 0 ? <EmptyState title="No appointments" description="Nothing is booked for this day. Check another date or review your availability." /> : null}

    {!loading && !error && appointments.length > 0 && <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="divide-y divide-border/70">
        {appointments.map(appointment => <article key={appointment.id} className={`grid gap-3 p-4 sm:grid-cols-[92px_minmax(0,1fr)_auto] sm:items-start ${appointment.status === 'cancelled' ? 'opacity-65' : ''}`}>
          <div className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground sm:flex-col sm:items-start sm:gap-0.5">
            <span className="inline-flex items-center gap-1 text-foreground"><Clock3 size={12} aria-hidden="true" />{formatTime(appointment.start_time, timeZone)}</span>
            <span className="text-[10px]">{formatTime(appointment.end_time, timeZone)}</span>
          </div>
          <div className="flex min-w-0 items-start gap-3">
            <Avatar name={`${appointment.patient.first_name} ${appointment.patient.last_name}`} size="sm" />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <p className="text-sm font-semibold">{appointment.patient.first_name} {appointment.patient.last_name}</p>
                <StatusBadge status={appointment.status} />
              </div>
              <p className="mt-1 text-[11px] text-muted-foreground">{appointment.reason}</p>
              <p className="mt-1 text-[10px] text-muted-foreground">{appointment.patient.email}</p>
            </div>
          </div>
          {appointment.status === 'confirmed' && (canReschedule || canCancel) && <div className="flex gap-1.5 sm:flex-col">
            {canReschedule && <Button variant="outline" size="sm" className="gap-1" onClick={() => onNavigate(`/book-appointment?reschedule=${appointment.id}`)}><RotateCcw size={12} aria-hidden="true" /> Reschedule</Button>}
            {canCancel && <Button variant="destructive" size="sm" className="gap-1" onClick={() => { setCancelling(appointment); setReason(''); setCancelError(null) }}><XCircle size={12} aria-hidden="true" /> Cancel</Button>}
          </div>}
        </article>)}
      </div>
    </div>}


    <Dialog open={Boolean(cancelling)} onOpenChange={open => { if (!open && !submitting) setCancelling(null) }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel this visit?</DialogTitle>
          <DialogDescription>{cancelling ? `${cancelling.patient.first_name} ${cancelling.patient.last_name} · ${formatTimeRange(cancelling.start_time, cancelling.end_time, timeZone)}` : ''}</DialogDescription>
        </DialogHeader>
        {cancelError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{cancelError}</div>}
        <Textarea label="Reason" value={reason} onChange={event => setReason(event.target.value)} rows={3} required placeholder="Reason for cancelling" />
        <DialogFooter>
          <Button variant="outline" onClick={() => setCancelling(null)} disabled={submitting}>Keep visit</Button>
          <Button variant="destructive" onClick={() => void confirmCancel()} disabled={submitting}>{submitting && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Cancel visit</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </main>
}

