import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, LoaderCircle, Plus, RotateCcw, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/arc/textarea/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState, ErrorState, LoadingState, PageHeader, StatusBadge } from '@/components/shared/page'
import { ManualBookingDialog } from '@/components/admin/manual-booking-dialog'
import { RescheduleDialog } from '@/components/admin/reschedule-dialog'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import type { Appointment, Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, dateKeyInZone, formatDate, formatTimeRange, zonedDayBoundsFromKey } from '@/lib/time'

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
  const [status, setStatus] = useState('')
  const [items, setItems] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const bounds = zonedDayBoundsFromKey(date, CLINIC_TIMEZONE)
      const page = await appointmentsApi.list({ scope: 'all', from: bounds.from, to: bounds.to, doctorId: doctorId || undefined, order: 'asc', limit: 100 })
      setItems(page.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load appointments.')
    } finally {
      setLoading(false)
    }
  }, [date, doctorId])

  useEffect(() => { void load() }, [load])

  async function confirmCancel() {
    if (!cancelling) return
    if (!reason.trim()) { setCancelError('Add a cancellation reason.'); return }
    setSaving(true)
    setCancelError(null)
    try {
      await appointmentsApi.cancel(cancelling.id, reason.trim())
      notify('Appointment cancelled.')
      setCancelling(null)
      await load()
    } catch (err) {
      setCancelError(err instanceof ApiError ? err.message : 'Unable to cancel the appointment.')
    } finally {
      setSaving(false)
    }
  }

  const visible = status ? items.filter(item => item.status === status) : items

  return <main className="mx-auto w-full max-w-[1400px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader eyebrow="Clinic" title="Appointments" description="Review the diary, cancel with a reason, reschedule or book for a patient." actions={canCreate ? <Button className="gap-2" onClick={() => setBookingOpen(true)}><Plus size={15} aria-hidden="true" /> Book for patient</Button> : undefined} />
    <div className="mb-4 grid gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:grid-cols-[160px_minmax(0,1fr)_160px]">
      <input type="date" value={date} onChange={event => setDate(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-xs" aria-label="Appointment date" />
      <select value={doctorId} onChange={event => setDoctorId(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-xs">
        <option value="">All doctors</option>
        {doctors.map(doctor => <option key={doctor.id} value={doctor.id}>Dr. {doctor.first_name} {doctor.last_name}</option>)}
      </select>
      <select value={status} onChange={event => setStatus(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-xs">
        <option value="">All statuses</option>
        <option value="confirmed">Confirmed</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading appointments…" /> : null}
    {!loading && !error && visible.length === 0 ? <EmptyState title="No appointments" description="Nothing matches this date, doctor and status." /> : null}
    {!loading && !error && visible.length > 0 && <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[860px] text-left text-xs">
          <thead className="border-b border-border bg-muted/40 text-[10px] uppercase tracking-[0.08em] text-muted-foreground"><tr><th className="px-4 py-3 font-semibold">Time</th><th className="px-4 py-3 font-semibold">Patient</th><th className="px-4 py-3 font-semibold">Doctor</th><th className="px-4 py-3 font-semibold">Reason</th><th className="px-4 py-3 font-semibold">Status</th><th className="px-4 py-3 text-right font-semibold">Actions</th></tr></thead>
          <tbody className="divide-y divide-border/70">
            {visible.map(appointment => <tr key={appointment.id}>
              <td className="px-4 py-3 text-[11px] font-semibold">{formatTimeRange(appointment.start_time, appointment.end_time, CLINIC_TIMEZONE)}</td>
              <td className="px-4 py-3"><p className="font-semibold">{appointment.patient.first_name} {appointment.patient.last_name}</p><p className="text-[10px] text-muted-foreground">{appointment.patient.email}</p></td>
              <td className="px-4 py-3">Dr. {appointment.doctor.first_name} {appointment.doctor.last_name}<p className="text-[10px] text-muted-foreground">{appointment.doctor.specialty}</p></td>
              <td className="max-w-[220px] px-4 py-3"><p className="truncate text-[11px] text-muted-foreground" title={appointment.reason}>{appointment.reason}</p></td>
              <td className="px-4 py-3"><StatusBadge status={appointment.status} /></td>
              <td className="px-4 py-3"><div className="flex justify-end gap-1">
                {appointment.status === 'confirmed' && canReschedule && <Button variant="outline" size="sm" className="gap-1" onClick={() => setRescheduling(appointment)}><RotateCcw size={11} aria-hidden="true" /> Reschedule</Button>}
                {appointment.status === 'confirmed' && canCancel && <Button variant="destructive" size="sm" className="gap-1" onClick={() => { setCancelling(appointment); setReason(''); setCancelError(null) }}><XCircle size={11} aria-hidden="true" /> Cancel</Button>}
              </div></td>
            </tr>)}
          </tbody>
        </table>
      </div>
    </div>}


    <Dialog open={Boolean(cancelling)} onOpenChange={open => { if (!open && !saving) setCancelling(null) }}>
      <DialogContent>
        <DialogHeader><DialogTitle>Cancel appointment</DialogTitle><DialogDescription>{cancelling ? `${cancelling.patient.first_name} ${cancelling.patient.last_name} · ${formatDate(cancelling.start_time, CLINIC_TIMEZONE)}` : ''}</DialogDescription></DialogHeader>
        {cancelError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{cancelError}</div>}
        <Textarea label="Reason" value={reason} onChange={event => setReason(event.target.value)} rows={3} required />
        <DialogFooter><Button variant="outline" onClick={() => setCancelling(null)} disabled={saving}>Keep appointment</Button><Button variant="destructive" onClick={() => void confirmCancel()} disabled={saving}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Cancel</Button></DialogFooter>
      </DialogContent>
    </Dialog>

    <RescheduleDialog key={rescheduling?.id ?? 'closed'} appointment={rescheduling} open={Boolean(rescheduling)} onOpenChange={open => { if (!open) setRescheduling(null) }} onDone={load} />
    <ManualBookingDialog open={bookingOpen} onOpenChange={setBookingOpen} onBooked={load} />
  </main>
}

