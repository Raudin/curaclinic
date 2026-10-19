import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, CalendarDays, CalendarPlus, CalendarX2, Clock3, LoaderCircle, RotateCcw } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Tabs, TabsList, TabsTrigger } from '@/components/arc/tabs/tabs'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/arc/textarea/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState, ErrorState, LoadingState, PageHeader, StatusBadge } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { appointmentsApi } from '@/lib/api-clinic'
import { ApiError } from '@/lib/api'
import type { Appointment } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, formatDate, formatTimeRange } from '@/lib/time'

const CLINIC_TIMEZONE = import.meta.env.VITE_CLINIC_TIMEZONE || DEFAULT_TIMEZONE
type Scope = 'upcoming' | 'past' | 'cancelled'

const TABS: Array<{ id: Scope; label: string }> = [
  { id: 'upcoming', label: 'Upcoming' },
  { id: 'past', label: 'Past visits' },
  { id: 'cancelled', label: 'Cancelled' },
]

export function MyAppointmentsPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { notify } = useToast()
  const [tab, setTab] = useState<Scope>('upcoming')
  const [items, setItems] = useState<Appointment[]>([])
  const [counts, setCounts] = useState<Record<Scope, number>>({ upcoming: 0, past: 0, cancelled: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cancelling, setCancelling] = useState<Appointment | null>(null)
  const [reason, setReason] = useState('')
  const [cancelError, setCancelError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const loadCounts = useCallback(async () => {
    const [upcoming, past, cancelled] = await Promise.all([
      appointmentsApi.list({ scope: 'upcoming', limit: 1 }),
      appointmentsApi.list({ scope: 'past', limit: 1 }),
      appointmentsApi.list({ scope: 'cancelled', limit: 1 }),
    ])
    setCounts({ upcoming: upcoming.pagination.total, past: past.pagination.total, cancelled: cancelled.pagination.total })
  }, [])

  const loadTab = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const page = await appointmentsApi.list({ scope: tab, limit: 20 })
      setItems(page.data)
      setCounts(current => ({ ...current, [tab]: page.pagination.total }))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load appointments.')
    } finally {
      setLoading(false)
    }
  }, [tab])

  useEffect(() => { void loadCounts().catch(() => undefined) }, [loadCounts])
  useEffect(() => { void loadTab() }, [loadTab])

  function openCancel(appointment: Appointment) {
    setCancelling(appointment)
    setReason('')
    setCancelError(null)
  }

  async function confirmCancel() {
    if (!cancelling) return
    if (!reason.trim()) {
      setCancelError('Please tell the clinic why this appointment is being cancelled.')
      return
    }
    setSubmitting(true)
    setCancelError(null)
    try {
      await appointmentsApi.cancel(cancelling.id, reason.trim())
      notify('Your appointment has been cancelled and the slot released.')
      setCancelling(null)
      await Promise.all([loadCounts(), loadTab()])
    } catch (err) {
      setCancelError(err instanceof ApiError ? err.message : 'Unable to cancel this appointment.')
    } finally {
      setSubmitting(false)
    }
  }

  return <main className="mx-auto w-full max-w-[1200px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow="Appointments"
      title="My appointments"
      description="View upcoming visits, review past care, or release a slot back to the clinic."
      actions={<Button className="gap-2" onClick={() => onNavigate('/book-appointment')}><CalendarPlus size={15} aria-hidden="true" /> Book new visit</Button>}
    />

    <Tabs value={tab} onValueChange={value => setTab(value as Scope)} className="mb-5 w-full">
      <TabsList aria-label="Appointment status" className="!flex !w-full !justify-between">
        {TABS.map(entry => <TabsTrigger key={entry.id} value={entry.id} className="!min-w-0 !flex-1 text-xs">
          <span className="flex items-center gap-2">
            {entry.label}
            <span className="rounded-full bg-muted px-1.5 py-0.5 text-[9px] text-muted-foreground">{counts[entry.id]}</span>
          </span>
        </TabsTrigger>)}
      </TabsList>
    </Tabs>

    {error && <ErrorState message={error} onRetry={loadTab} />}
    {loading && !error ? <LoadingState label="Loading appointments…" /> : null}
    {!loading && !error && items.length === 0 ? <EmptyState
      title={tab === 'upcoming' ? 'No upcoming appointments' : tab === 'past' ? 'No past visits yet' : 'No cancelled appointments'}
      description={tab === 'upcoming' ? 'Book a visit with one of our doctors to get started.' : tab === 'past' ? 'Completed visits will appear here.' : 'Cancelled appointments will appear here.'}
      action={tab === 'upcoming' ? <Button className="gap-2" onClick={() => onNavigate('/book-appointment')}><CalendarPlus size={15} aria-hidden="true" /> Book appointment</Button> : undefined}
    /> : null}

    {!loading && !error && items.length > 0 && <div className="space-y-3">
      {items.map(appointment => <article key={appointment.id} className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 items-start gap-3">
            <Avatar name={`${appointment.doctor.first_name} ${appointment.doctor.last_name}`} size="lg" className="shrink-0" />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-sm font-semibold tracking-tight">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}</h2>
                <span className="rounded-full bg-muted px-2 py-0.5 text-[9px] font-medium text-muted-foreground">{appointment.doctor.specialty}</span>
                <StatusBadge status={appointment.status} />
              </div>
              <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-[11px] text-muted-foreground">
                <span className="inline-flex items-center gap-1"><CalendarDays size={12} aria-hidden="true" /> {formatDate(appointment.start_time, CLINIC_TIMEZONE)}</span>
                <span className="inline-flex items-center gap-1"><Clock3 size={12} aria-hidden="true" /> {formatTimeRange(appointment.start_time, appointment.end_time, CLINIC_TIMEZONE)}</span>
              </div>
              <p className="mt-2 rounded-lg bg-muted/60 px-3 py-2 text-[11px] leading-relaxed text-foreground">{appointment.reason}</p>
              {appointment.cancelled_at && <p className="mt-2 text-[10px] text-muted-foreground">Cancelled on {formatDate(appointment.cancelled_at, CLINIC_TIMEZONE, { weekday: undefined })}</p>}
            </div>
          </div>
          {appointment.status === 'confirmed' && <div className="flex shrink-0 flex-wrap gap-2 sm:flex-col">
            {tab !== 'past' && <Button variant="outline" size="sm" className="gap-1.5" onClick={() => onNavigate(`/book-appointment?reschedule=${appointment.id}`)}><RotateCcw size={13} aria-hidden="true" /> Reschedule</Button>}
            <Button variant="destructive" size="sm" className="gap-1.5" onClick={() => openCancel(appointment)}><CalendarX2 size={13} aria-hidden="true" /> Cancel</Button>
          </div>}
        </div>
      </article>)}
    </div>}


    <Dialog open={Boolean(cancelling)} onOpenChange={open => { if (!open && !submitting) setCancelling(null) }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel this appointment?</DialogTitle>
          <DialogDescription>
            {cancelling ? `Dr. ${cancelling.doctor.first_name} ${cancelling.doctor.last_name} · ${formatDate(cancelling.start_time, CLINIC_TIMEZONE)}` : ''}
          </DialogDescription>
        </DialogHeader>
        {cancelError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{cancelError}</div>}
        <Textarea
          label="Reason for cancellation"
          placeholder="Let the clinic know why you need to cancel"
          value={reason}
          onChange={event => setReason(event.target.value)}
          rows={3}
          required
        />
        <DialogFooter>
          <Button variant="outline" onClick={() => setCancelling(null)} disabled={submitting}>Keep appointment</Button>
          <Button variant="destructive" onClick={confirmCancel} disabled={submitting}>
            {submitting && <LoaderCircle size={14} className="animate-spin" aria-hidden="true" />} Cancel appointment
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </main>
}
