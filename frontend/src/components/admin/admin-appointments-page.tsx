import { DIARY_PAGE_SIZE, useDiary } from '@/hooks/use-diary'
import { canChangeAppointment, describeAppointment } from '@/lib/appointment-display'
import { useEffect, useRef, useState } from 'react'
import { AlertCircle, LoaderCircle, Plus, RotateCcw, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/arc/textarea/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState, ErrorState, LoadingState, Pagination, PageHeader, StatusBadge } from '@/components/shared/page'
import { ManualBookingDialog } from '@/components/admin/manual-booking-dialog'
import { RescheduleDialog } from '@/components/admin/reschedule-dialog'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import type { Appointment, AppointmentStatus, Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, dateKeyInZone, formatTimeRange } from '@/lib/time'

const CLINIC_TIMEZONE = import.meta.env.VITE_CLINIC_TIMEZONE || DEFAULT_TIMEZONE

export function AdminAppointmentsPage() {
  const { can, canAny } = useAuth()
  const { notify } = useToast()
  const canViewUsers = can(PERMISSIONS.adminUsersView)
  const canViewDoctors = canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)
  const canCreate = can(PERMISSIONS.adminAppointmentsCreate) && canViewUsers
  const canReschedule = can(PERMISSIONS.adminAppointmentsReschedule)
  const canCancel = can(PERMISSIONS.adminAppointmentsCancel)
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [date, setDate] = useState(() => dateKeyInZone(new Date(), CLINIC_TIMEZONE))
  const [doctorId, setDoctorId] = useState('')
  const [status, setStatus] = useState<AppointmentStatus | ''>('')
  const [offset, setOffset] = useState(0)
  const cancelTrigger = useRef<HTMLButtonElement>(null)
  const [cancelling, setCancelling] = useState<Appointment | null>(null)
  const [rescheduling, setRescheduling] = useState<Appointment | null>(null)
  const [reason, setReason] = useState('')
  const [cancelError, setCancelError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [bookingOpen, setBookingOpen] = useState(false)

  useEffect(() => {
    if (!canViewDoctors) return
    doctorsApi.list().then(setDoctors).catch(() => undefined)
  }, [canViewDoctors])

  const { items: visible, total, loading, error, now, refresh: load } = useDiary({
    date, timeZone: CLINIC_TIMEZONE, doctorId: doctorId || undefined, status: status || undefined, offset,
  })
  useEffect(() => {
    if (!loading && !error && offset > 0 && offset >= total) setOffset(Math.max(0, Math.ceil(total / DIARY_PAGE_SIZE) - 1) * DIARY_PAGE_SIZE)
  }, [loading, error, offset, total])

  async function confirmCancel() {
    if (!cancelling) return
    if (!canCancel || !canChangeAppointment(cancelling)) {
      notify('This visit can no longer be cancelled.')
      setCancelling(null)
      load()
      return
    }
    if (!reason.trim()) { setCancelError('Add a cancellation reason.'); return }
    setSaving(true)
    setCancelError(null)
    try {
      await appointmentsApi.cancel(cancelling.id, reason.trim())
      notify('Appointment cancelled.')
      setCancelling(null)
      load()
    } catch (err) {
      load()
      setCancelError(err instanceof ApiError ? err.message : 'Unable to cancel the appointment.')
    } finally {
      setSaving(false)
    }
  }


  return <main className="mx-auto w-full max-w-[1400px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader title="Appointments" description="Find and manage visits by date, doctor and status." actions={canCreate ? <Button className="gap-2" onClick={() => setBookingOpen(true)}><Plus size={15} aria-hidden="true" /> Book for patient</Button> : undefined} />
    <div className="mb-4 grid gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:grid-cols-[160px_minmax(0,1fr)_160px]">
      <input type="date" value={date} onChange={event => { if (event.target.value) { setDate(event.target.value); setOffset(0) } }} className="h-11 min-w-0 rounded-md border border-border bg-background px-3 text-sm" aria-label="Appointment date" />
      <select aria-label="Doctor" value={doctorId} onChange={event => { setDoctorId(event.target.value); setOffset(0) }} className="h-11 min-w-0 rounded-md border border-border bg-background px-3 text-sm">
        <option value="">All doctors</option>
        {doctors.map(doctor => <option key={doctor.id} value={doctor.id}>Dr. {doctor.first_name} {doctor.last_name}</option>)}
      </select>
      <select aria-label="Appointment status" value={status} onChange={event => { setStatus(event.target.value as AppointmentStatus | ''); setOffset(0) }} className="h-11 min-w-0 rounded-md border border-border bg-background px-3 text-sm">
        <option value="">All statuses</option>
        <option value="confirmed">Confirmed</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading appointments…" /> : null}
    {!loading && !error && visible.length === 0 ? <EmptyState title="No appointments" description="Nothing matches this date, doctor and status." /> : null}
    {!loading && !error && visible.length > 0 && <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="@container/diary">
        <table className="block w-full text-left text-sm @min-[860px]/diary:table @min-[860px]/diary:text-xs">
          <thead className="hidden @min-[860px]/diary:table-header-group border-b border-border bg-muted/40 text-xs uppercase tracking-[0.08em] text-muted-foreground"><tr><th className="px-4 py-3 font-semibold">Time</th><th className="px-4 py-3 font-semibold">Patient</th><th className="px-4 py-3 font-semibold">Doctor</th><th className="px-4 py-3 font-semibold">Reason</th><th className="px-4 py-3 font-semibold">Status</th><th className="px-4 py-3 text-right font-semibold">Actions</th></tr></thead>
          <tbody className="block @min-[860px]/diary:table-row-group divide-y divide-border/70">
            {visible.map(appointment => <tr key={appointment.id} className="grid gap-1 px-4 py-4 @min-[860px]/diary:table-row [&>td]:min-w-0 [&>td]:break-words @max-[860px]/diary:[&>td]:p-0">
              <td className="px-4 py-3 text-xs font-semibold">{formatTimeRange(appointment.start_time, appointment.end_time, CLINIC_TIMEZONE)}</td>
              <td className="px-4 py-3"><p className="font-semibold">{appointment.patient.first_name} {appointment.patient.last_name}</p><p className="text-xs text-muted-foreground">{appointment.patient.email}</p></td>
              <td className="px-4 py-3">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}<p className="text-xs text-muted-foreground">{appointment.doctor.specialty}</p></td>
              <td className="max-w-[220px] px-4 py-3"><p className="truncate text-xs text-muted-foreground" title={appointment.reason}>{appointment.reason}</p></td>
              <td className="px-4 py-3"><StatusBadge status={appointment.status} /></td>
              <td className="px-4 py-3"><div className="mt-2 flex flex-wrap gap-2 @min-[860px]/diary:mt-0 @min-[860px]/diary:justify-end">
                {canChangeAppointment(appointment, now) && canReschedule && <Button variant="outline" size="sm" className="min-h-10 gap-1" onClick={() => setRescheduling(appointment)}><RotateCcw size={11} aria-hidden="true" /> Reschedule</Button>}
                {canChangeAppointment(appointment, now) && canCancel && <Button variant="destructive" size="sm" className="min-h-10 gap-1" onClick={event => { cancelTrigger.current = event.currentTarget; setCancelling(appointment); setReason(''); setCancelError(null) }}><XCircle size={11} aria-hidden="true" /> Cancel</Button>}
              </div></td>
            </tr>)}
          </tbody>
        </table>
      </div>
    </div>}


    {!loading && !error && <Pagination limit={DIARY_PAGE_SIZE} offset={offset} total={total} onChange={setOffset} />}

    <Dialog open={Boolean(cancelling)} onOpenChange={open => { if (!open && !saving) setCancelling(null) }}>
      <DialogContent onCloseAutoFocus={event => {
        if (cancelTrigger.current?.isConnected) {
          event.preventDefault()
          cancelTrigger.current.focus()
        }
      }}>
        <DialogHeader><DialogTitle>Cancel appointment</DialogTitle><DialogDescription>{cancelling ? describeAppointment(cancelling, CLINIC_TIMEZONE) : ''}</DialogDescription></DialogHeader>
        {cancelError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{cancelError}</div>}
        <Textarea label="Reason" value={reason} onChange={event => setReason(event.target.value)} rows={3} required />
        <DialogFooter><Button variant="outline" onClick={() => setCancelling(null)} disabled={saving}>Keep appointment</Button><Button variant="destructive" onClick={() => void confirmCancel()} disabled={saving}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Cancel</Button></DialogFooter>
      </DialogContent>
    </Dialog>

    <RescheduleDialog key={rescheduling?.id ?? 'closed'} appointment={rescheduling} canReschedule={canReschedule} timeZone={CLINIC_TIMEZONE} open={Boolean(rescheduling)} onOpenChange={open => { if (!open) setRescheduling(null) }} onDone={load} />
    <ManualBookingDialog open={bookingOpen} onOpenChange={setBookingOpen} onBooked={load} />
  </main>
}

