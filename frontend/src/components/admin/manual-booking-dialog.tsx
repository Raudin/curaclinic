import { describeSlot } from '@/lib/appointment-display'
import { DEFAULT_TIMEZONE } from '@/lib/time'
import { useEffect, useState } from 'react'
import { AlertCircle, LoaderCircle, Search } from 'lucide-react'
import { MorphSelect } from '@/components/arc/morph-select/morph-select'
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
  const [doctorState, setDoctorState] = useState<'loading' | 'ready' | 'error'>('loading')
  const [patientState, setPatientState] = useState<'idle' | 'loading' | 'ready' | 'error'>('idle')
  const [doctorAttempt, setDoctorAttempt] = useState(0)
  const [patientAttempt, setPatientAttempt] = useState(0)
  const [saving, setSaving] = useState(false)
  const selectedDoctor = doctors.find(doctor => doctor.id === doctorId)

  useEffect(() => {
    if (!open) return
    setError(null); setPatient(null); setSlot(null); setReason(''); setSearch(''); setPatients([])
    setDoctorId('')
  }, [open])

  useEffect(() => {
    if (!open) return
    let active = true
    setDoctors([])
    setDoctorState('loading')
    doctorsApi.list({ accepting: true }).then(list => {
      if (active) { setDoctors(list); setDoctorState('ready') }
    }).catch(() => { if (active) setDoctorState('error') })
    return () => { active = false }
  }, [open, doctorAttempt])

  useEffect(() => {
    setPatients([])
    if (!open || patient || search.trim().length < 2) { setPatientState('idle'); return }
    let active = true
    setPatientState('loading')
    const timer = window.setTimeout(() => {
      usersApi.list({ role: 'patient', search: search.trim(), limit: 6 }).then(page => {
        if (active) { setPatients(page.data); setPatientState('ready') }
      }).catch(() => { if (active) setPatientState('error') })
    }, 250)
    return () => { active = false; window.clearTimeout(timer) }
  }, [open, search, patient, patientAttempt])

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
        <DialogDescription>Choose a patient, then select a doctor and an available time.</DialogDescription>
      </DialogHeader>
      {error && <div role="alert" className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div>}
      <div className="grid gap-2 sm:grid-cols-2">
        <div className="min-w-0"><label htmlFor="booking-patient" className="text-xs font-medium">Patient</label>
          <div className="relative mt-1"><Search size={13} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" aria-hidden="true" /><input id="booking-patient" value={patient ? `${patient.first_name} ${patient.last_name}` : search} onChange={event => { setPatient(null); setSearch(event.target.value) }} placeholder="Search patient by name or email" className="h-10 w-full rounded-md border border-border bg-background pl-7 pr-2 text-xs" /></div>
          {patients.length > 0 && !patient && <div className="mt-1 overflow-hidden rounded-lg border border-border bg-card">{patients.map(item => <button key={item.id} type="button" className="block w-full px-2 py-1.5 text-left text-[11px] hover:bg-muted" onClick={() => { setPatient(item); setPatients([]) }}><span className="font-semibold">{item.first_name} {item.last_name}</span><span className="ml-1 text-[10px] text-muted-foreground">{item.email}</span></button>)}</div>}
          {patientState === 'loading' && <p role="status" className="mt-2 text-xs">Searching patients…</p>}
          {patientState === 'ready' && !patients.length && !patient && <p role="status" className="mt-2 text-xs">No patients match this search.</p>}
          {patientState === 'error' && <p role="alert" className="mt-2 text-xs">Patient search unavailable. <Button size="sm" variant="outline" onClick={() => setPatientAttempt(value => value + 1)}>Retry search</Button></p>}
        </div>
        <div className="@container min-w-0">
          <p className="text-[10px] font-semibold uppercase tracking-[0.08em] text-muted-foreground" aria-hidden="true">Doctor</p>
          <MorphSelect
            label="Doctor"
            hideLabel
            className="mt-1 ![--ms-panel-w:min(17rem,100cqi)]"
            value={doctorId || null}
            items={doctors.map(doctor => ({
              value: doctor.id,
              label: `Dr. ${doctor.first_name} ${doctor.last_name}`,
              meta: doctor.specialty,
              keywords: doctor.email,
            }))}
            onValueChange={value => { setDoctorId(value); setSlot(null) }}
            searchable
            searchPlaceholder="Search doctors"
            placeholder={doctorState === 'loading' ? 'Loading doctors…' : doctorState === 'error' ? 'Directory unavailable' : doctors.length ? 'Choose a doctor' : 'No doctors accepting patients'}
            disabled={doctors.length === 0}
            maxListHeight={240}
          />
          {doctorState === 'error' && <Button size="sm" variant="outline" onClick={() => setDoctorAttempt(value => value + 1)}>Retry directory</Button>}
        </div>
      </div>
      {doctorId ? <SlotPicker key={doctorId} doctorId={doctorId} selected={slot} onSelect={setSlot} /> : <p className="rounded-xl border border-dashed border-border px-4 py-6 text-center text-[11px] text-muted-foreground">Select a doctor to see open times.</p>}
      <label className="grid gap-1 text-xs font-medium">Reason for visit<input value={reason} onChange={event => setReason(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 text-sm" /></label>
      {slot && patient && selectedDoctor && <p role="status" className="rounded-xl bg-muted p-3 text-sm"><strong>{patient.first_name} {patient.last_name}</strong> · Dr. {selectedDoctor.first_name} {selectedDoctor.last_name}<br />{describeSlot(slot, selectedDoctor.timezone || DEFAULT_TIMEZONE)}</p>}
      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>Cancel</Button>
        <Button onClick={() => void submit()} disabled={saving || !slot || !patient || !doctorId}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Book appointment</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
}
