import { useEffect, useState } from 'react'
import { AlertCircle, LoaderCircle, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { SlotPicker } from '@/components/shared/slot-picker'
import { useToast } from '@/components/feedback/toast-provider'
import { ApiError } from '@/lib/api'
import { appointmentsApi, doctorsApi } from '@/lib/api-clinic'
import { usersApi } from '@/lib/api-admin'
import type { Doctor, Slot, User } from '@/lib/api-types'

export function ManualBookingDialog({ open, onOpenChange, onBooked }: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onBooked: () => void
}) {
  const { notify } = useToast()
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [doctorId, setDoctorId] = useState('')
  const [search, setSearch] = useState('')
  const [patients, setPatients] = useState<User[]>([])
  const [patient, setPatient] = useState<User | null>(null)
  const [slot, setSlot] = useState<Slot | null>(null)
  const [reason, setReason] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setError(null); setPatient(null); setSlot(null); setReason(''); setSearch(''); setPatients([])
    doctorsApi.list({ accepting: true }).then(setDoctors).catch(() => undefined)
  }, [open])

  useEffect(() => {
    if (!open || search.trim().length < 2) { setPatients([]); return }
    const timer = window.setTimeout(() => {
      usersApi.list({ role: 'patient', search: search.trim(), limit: 6 }).then(page => setPatients(page.data)).catch(() => undefined)
    }, 250)
    return () => window.clearTimeout(timer)
  }, [open, search])

  async function submit() {
    if (!doctorId || !patient || !slot || !reason.trim()) {
      setError('Choose a patient, doctor, time and reason.')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await appointmentsApi.manualBook({ doctor_id: doctorId, patient_id: patient.id, start_time: slot.start_time, reason: reason.trim() })
      notify('Appointment booked for the patient.')
      onBooked()
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to book this appointment.')
    } finally {
      setSaving(false)
    }
  }

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Book for a patient</DialogTitle>
        <DialogDescription>Front-desk booking does not require a patient hold; the database still guarantees no double booking.</DialogDescription>
      </DialogHeader>
      {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div>}
      <div className="grid gap-2 sm:grid-cols-2">
        <label className="text-[10px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">Patient
          <div className="relative mt-1"><Search size={13} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" aria-hidden="true" /><input value={patient ? `${patient.first_name} ${patient.last_name}` : search} onChange={event => { setPatient(null); setSearch(event.target.value) }} placeholder="Search patient by name or email" className="h-9 w-full rounded-md border border-border bg-background pl-7 pr-2 text-xs" /></div>
          {patients.length > 0 && !patient && <div className="mt-1 overflow-hidden rounded-lg border border-border bg-card">{patients.map(item => <button key={item.id} type="button" className="block w-full px-2 py-1.5 text-left text-[11px] hover:bg-muted" onClick={() => { setPatient(item); setPatients([]) }}><span className="font-semibold">{item.first_name} {item.last_name}</span><span className="ml-1 text-[10px] text-muted-foreground">{item.email}</span></button>)}</div>}
        </label>
        <label className="text-[10px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">Doctor
          <select value={doctorId} onChange={event => { setDoctorId(event.target.value); setSlot(null) }} className="mt-1 h-9 w-full rounded-md border border-border bg-background px-2 text-xs">
            <option value="">Choose a doctor…</option>
            {doctors.map(doctor => <option key={doctor.id} value={doctor.id}>Dr. {doctor.first_name} {doctor.last_name} · {doctor.specialty}</option>)}
          </select>
        </label>
      </div>
      {doctorId ? <SlotPicker key={doctorId} doctorId={doctorId} selected={slot} onSelect={setSlot} /> : <p className="rounded-xl border border-dashed border-border px-4 py-6 text-center text-[11px] text-muted-foreground">Select a doctor to see open times.</p>}
      <input value={reason} onChange={event => setReason(event.target.value)} placeholder="Reason for visit" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>Cancel</Button>
        <Button onClick={() => void submit()} disabled={saving || !slot || !patient || !doctorId}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Book appointment</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
}
