import { useState } from 'react'
import { AlertCircle, LoaderCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { SlotPicker } from '@/components/shared/slot-picker'
import { useToast } from '@/components/feedback/toast-provider'
import { ApiError } from '@/lib/api'
import { appointmentsApi } from '@/lib/api-clinic'
import type { Appointment, Slot } from '@/lib/api-types'

export function RescheduleDialog({ appointment, open, onOpenChange, onDone }: {
  appointment: Appointment | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onDone: () => void
}) {
  const { notify } = useToast()
  const [slot, setSlot] = useState<Slot | null>(null)
  const [reason, setReason] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  async function submit() {
    if (!appointment || !slot) return
    setSaving(true)
    setError(null)
    try {
      await appointmentsApi.reschedule(appointment.id, slot.start_time, reason.trim() || undefined)
      notify('Appointment rescheduled.')
      onDone()
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to reschedule this appointment.')
    } finally {
      setSaving(false)
    }
  }

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Reschedule appointment</DialogTitle>
        <DialogDescription>{appointment ? `${appointment.patient.first_name} ${appointment.patient.last_name} with Dr. ${appointment.doctor.first_name} ${appointment.doctor.last_name}` : ''}</DialogDescription>
      </DialogHeader>
      {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div>}
      {appointment && <SlotPicker doctorId={appointment.doctor.id} selected={slot} onSelect={setSlot} />}
      <input value={reason} onChange={event => setReason(event.target.value)} placeholder="Reason for rescheduling (optional)" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>Cancel</Button>
        <Button onClick={() => void submit()} disabled={!slot || saving}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Confirm reschedule</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
}
