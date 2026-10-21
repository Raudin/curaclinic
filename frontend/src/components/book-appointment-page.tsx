import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { AlertCircle, ArrowLeft, ArrowRight, CalendarDays, Check, Clock3, Languages, LoaderCircle, Stethoscope, UserRound } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Badge } from '@/components/arc/badge/badge'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/arc/calendar/calendar'
import { MorphSelect } from '@/components/arc/morph-select/morph-select'
import { Stepper } from '@/components/arc/stepper/stepper'
import { Textarea } from '@/components/arc/textarea/textarea'
import { EmptyState, ErrorState, LoadingState, PageHeader } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { appointmentsApi, availabilityApi, doctorsApi, holdsApi } from '@/lib/api-clinic'
import { ApiError } from '@/lib/api'
import type { Appointment, Availability, Doctor, Hold, Slot } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, addDays, dateFromKey, dateKey, dateKeyInZone, formatDate, formatTime, formatTimeRange } from '@/lib/time'

const BOOKING_STEPS = [
  { id: 'provider', label: 'Provider' },
  { id: 'date-time', label: 'Date & time' },
  { id: 'confirm', label: 'Confirm' },
]
const MAX_HORIZON_DAYS = 90

function splitSlots(slots: Slot[], timeZone: string) {
  const morning: Array<{ slot: Slot; label: string }> = []
  const afternoon: Array<{ slot: Slot; label: string }> = []
  for (const slot of slots) {
    const label = formatTime(slot.start_time, timeZone)
    const target = Number(label.slice(0, 2)) < 12 ? morning : afternoon
    target.push({ slot, label })
  }
  return { morning, afternoon }
}

export function BookAppointmentPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { notify } = useToast()
  const { roles } = useAuth()
  const reduceMotion = useReducedMotion()
  const rescheduleId = useMemo(() => new URLSearchParams(window.location.search).get('reschedule'), [])
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [doctorId, setDoctorId] = useState<string | null>(null)
  const [reschedule, setReschedule] = useState<Appointment | null>(null)
  const [step, setStep] = useState(0)
  const [selectedDate, setSelectedDate] = useState(() => dateKeyInZone(new Date(), DEFAULT_TIMEZONE))
  const [availability, setAvailability] = useState<Availability | null>(null)
  const [availabilityLoading, setAvailabilityLoading] = useState(false)
  const [availabilityError, setAvailabilityError] = useState<string | null>(null)
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null)
  const [hold, setHold] = useState<Hold | null>(null)
  const [holdLoading, setHoldLoading] = useState(false)
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [now, setNow] = useState(Date.now())
  const requestId = useRef(0)

  const doctor = doctors.find(item => item.id === doctorId) ?? null
  const acceptingDoctors = doctors.filter(item => item.is_accepting_patients)
  const timeZone = doctor?.timezone || DEFAULT_TIMEZONE
  const backPath = roles.includes('doctor') ? '/daily-agenda' : '/my-appointments'

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const list = await doctorsApi.list()
        if (cancelled) return
        setDoctors(list)
        if (rescheduleId) {
          const appointment = await appointmentsApi.get(rescheduleId)
          if (cancelled) return
          setReschedule(appointment)
          setDoctorId(appointment.doctor.id)
          setReason(appointment.reason)
          setStep(1)
          setSelectedDate(dateKeyInZone(new Date(appointment.start_time), DEFAULT_TIMEZONE))
        }
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Unable to load the booking form.')
      }
    }
    void load()
    return () => { cancelled = true }
  }, [rescheduleId])

  const loadAvailability = useCallback(async () => {
    if (!doctorId) return
    const current = ++requestId.current
    setAvailabilityLoading(true)
    setAvailabilityError(null)
    try {
      const result = await availabilityApi.get(doctorId, selectedDate)
      if (current === requestId.current) setAvailability(result)
    } catch (err) {
      if (current === requestId.current) {
        setAvailability(null)
        setAvailabilityError(err instanceof ApiError ? err.message : 'Unable to load available times.')
      }
    } finally {
      if (current === requestId.current) setAvailabilityLoading(false)
    }
  }, [doctorId, selectedDate])

  useEffect(() => {
    setSelectedSlot(null)
    void loadAvailability()
  }, [loadAvailability])

  useEffect(() => {
    if (!hold) return
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [hold])

  const holdSeconds = hold ? Math.max(0, Math.ceil((new Date(hold.expires_at).getTime() - now) / 1000)) : 0

  useEffect(() => {
    if (!hold || holdSeconds > 0) return
    const expiredToken = hold.token
    setHold(null)
    setSelectedSlot(null)
    notify('Your slot hold expired. Please choose a time again.', 'Hold expired')
    void holdsApi.release(expiredToken).catch(() => undefined)
  }, [hold, holdSeconds, notify])

  async function releaseHold() {
    const current = hold
    setHold(null)
    setSelectedSlot(null)
    if (current) await holdsApi.release(current.token).catch(() => undefined)
  }

  async function chooseDoctor(id: string) {
    if (id === doctorId) return
    await releaseHold()
    setDoctorId(id)
    setSelectedDate(dateKeyInZone(new Date(), DEFAULT_TIMEZONE))
  }

  async function chooseSlot(slot: Slot) {
    if (!doctorId) return
    if (hold && selectedSlot?.start_time === slot.start_time) return
    setHoldLoading(true)
    setError(null)
    const previous = hold
    setHold(null)
    setSelectedSlot(null)
    if (previous) void holdsApi.release(previous.token).catch(() => undefined)
    try {
      const created = await holdsApi.create(doctorId, slot.start_time)
      setHold(created)
      setSelectedSlot(slot)
    } catch (err) {
      notify(err instanceof ApiError ? err.message : 'That time is no longer available.', 'Slot unavailable')
      await loadAvailability()
    } finally {
      setHoldLoading(false)
    }
  }

  async function confirm() {
    if (!doctorId || !selectedSlot) return
    if (!reason.trim()) {
      setError('Please add a reason for the visit.')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      if (reschedule) {
        await appointmentsApi.reschedule(reschedule.id, selectedSlot.start_time, reason.trim())
        notify('Your appointment has been rescheduled.')
      } else {
        if (!hold) {
          setError('Choose a time again so we can protect your slot.')
          setStep(1)
          return
        }
        await appointmentsApi.book({
          doctor_id: doctorId,
          start_time: selectedSlot.start_time,
          reason: reason.trim(),
          hold_token: hold.token,
        })
        notify('Your appointment is booked.')
      }
      onNavigate(backPath)
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Unable to complete the booking.'
      setError(message)
      if (err instanceof ApiError && (err.code === 'hold_expired' || err.code === 'slot_unavailable')) {
        setHold(null)
        setSelectedSlot(null)
        setStep(1)
        void loadAvailability()
      }
    } finally {
      setSubmitting(false)
    }
  }

  const groups = splitSlots(availability?.slots ?? [], timeZone)
  const minDate = dateFromKey(dateKeyInZone(new Date(), timeZone))
  const maxDate = addDays(minDate, MAX_HORIZON_DAYS)


  return <main className="mx-auto w-full max-w-[1200px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow={reschedule ? 'Reschedule' : 'Book a visit'}
      title={reschedule ? 'Choose a new time' : 'Book an appointment'}
      description={reschedule ? 'Pick a new available slot for this appointment.' : 'Choose a doctor, then select a time that works for you.'}
    />

    <div className="rounded-3xl border border-border bg-card p-4 shadow-sm sm:p-6">
      <Stepper
        steps={BOOKING_STEPS}
        current={step}
        onStepSelect={index => { if ((index === 0 && reschedule) || (index >= 1 && !doctorId) || (index === 2 && !selectedSlot)) return; setStep(index) }}
        details="current"
        label="Booking progress"
        className="mb-5"
      />

      {error && <div className="mb-4 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}

      {step === 0 && !reschedule && (doctors.length === 0 ? <LoadingState label="Loading doctors…" /> : acceptingDoctors.length === 0 ? <EmptyState title="No doctors are accepting appointments" description="Please contact the clinic to arrange a visit." /> : <section aria-label="Choose a provider" className="grid items-start gap-4 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        <div className="rounded-2xl border border-border bg-background p-4 sm:p-5">
          <h2 className="text-sm font-semibold">Who would you like to see?</h2>
          <p className="mt-1 text-[11px] leading-relaxed text-muted-foreground">Search by name or specialty. {acceptingDoctors.length} {acceptingDoctors.length === 1 ? 'doctor is' : 'doctors are'} accepting new appointments.</p>
          <MorphSelect
            label="Doctor"
            hideLabel
            className="mt-4"
            value={doctorId}
            items={acceptingDoctors.map(item => ({
              value: item.id,
              label: `Dr. ${item.first_name} ${item.last_name}`,
              meta: item.specialty,
              keywords: `${item.specialty} ${item.languages.join(' ')}`,
              icon: <Avatar name={`${item.first_name} ${item.last_name}`} src={item.avatar_url} size="sm" />,
            }))}
            onValueChange={value => void chooseDoctor(value)}
            searchable="auto"
            searchPlaceholder="Search doctors or specialties"
            placeholder="Select a doctor"
            panelWidth={320}
          />
        </div>

        {/* The profile card only appears once a doctor is picked; the dashed slot holds its place so the layout doesn't jump. */}
        <AnimatePresence mode="wait" initial={false}>
          {doctor ? <motion.div
            key={doctor.id}
            initial={reduceMotion ? { opacity: 0 } : { opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={reduceMotion ? { opacity: 0 } : { opacity: 0, y: -4 }}
            transition={{ duration: 0.18, ease: 'easeOut' }}
            className="rounded-2xl border border-border bg-background p-4 sm:p-5"
            aria-live="polite"
          >
            <div className="flex items-start gap-3 sm:gap-4">
              <Avatar name={`${doctor.first_name} ${doctor.last_name}`} src={doctor.avatar_url} size="xl" className="shrink-0" />
              <div className="min-w-0 flex-1">
                <h3 className="text-base font-semibold tracking-tight">Dr. {doctor.first_name} {doctor.last_name}</h3>
                <p className="mt-0.5 text-xs text-muted-foreground">{doctor.specialty}</p>
                <div className="mt-2 flex flex-wrap gap-1.5">
                  <Badge tone="neutral" size="sm">{doctor.experience_years} years experience</Badge>
                  <Badge tone="neutral" size="sm">{doctor.slot_minutes}-minute visits</Badge>
                </div>
              </div>
            </div>
            <p className="mt-4 text-xs leading-relaxed text-muted-foreground">{doctor.bio || 'Available for in-clinic consultations.'}</p>
            {doctor.languages.length > 0 && <p className="mt-3 flex items-center gap-1.5 text-[11px] text-muted-foreground"><Languages size={13} aria-hidden="true" /> Speaks {doctor.languages.join(', ')}</p>}
            <div className="mt-4 flex justify-end border-t border-border pt-3">
              <Button size="sm" className="w-full gap-1.5 sm:w-auto" onClick={() => setStep(1)}>Choose a date & time <ArrowRight size={13} aria-hidden="true" /></Button>
            </div>
          </motion.div> : <motion.div
            key="placeholder"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.12 }}
            className="hidden min-h-[220px] flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-border px-6 text-center lg:flex"
          >
            <span className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground"><UserRound size={18} aria-hidden="true" /></span>
            <p className="text-xs font-semibold">No doctor selected</p>
            <p className="max-w-xs text-[11px] text-muted-foreground">Pick a doctor on the left to see their profile and continue to available times.</p>
          </motion.div>}
        </AnimatePresence>
      </section>)}

      {step === 1 && <section aria-label="Choose a date and time">
        {doctor && <div className="mb-4 flex items-center gap-3 rounded-2xl bg-muted/50 px-3 py-2.5">
          <Avatar name={`${doctor.first_name} ${doctor.last_name}`} src={doctor.avatar_url} size="sm" />
          <div className="min-w-0 flex-1"><p className="text-xs font-semibold">Dr. {doctor.first_name} {doctor.last_name}</p><p className="text-[10px] text-muted-foreground">{doctor.specialty} · {doctor.slot_minutes}-minute visits</p></div>
          {!reschedule && <button type="button" className="text-[10px] font-semibold underline underline-offset-4" onClick={() => setStep(0)}>Change</button>}
        </div>}
        <div className="grid gap-4 lg:grid-cols-[320px_minmax(0,1fr)]">
          <div className="rounded-2xl border border-border bg-background p-3">
            <Calendar
              value={dateFromKey(selectedDate)}
              onChange={date => setSelectedDate(dateKey(date))}
              minDate={minDate}
              maxDate={maxDate}
              locale="en-KE"
            />
            <div className="mt-2 flex items-center justify-between gap-2 rounded-xl bg-muted/60 px-3 py-2 text-[10px] text-muted-foreground">
              <span className="inline-flex items-center gap-1"><CalendarDays size={12} aria-hidden="true" /> {formatDate(`${selectedDate}T12:00:00Z`, timeZone, { weekday: 'long', year: undefined })}</span>
              <span>{availability?.slots.length ?? 0} open</span>
            </div>
            <p className="mt-2 flex items-center gap-1 px-1 text-[9px] text-muted-foreground"><Clock3 size={11} aria-hidden="true" /> Times shown in {timeZone.replace('_', ' ')}</p>
          </div>


          <div className="rounded-2xl border border-border bg-background p-3 sm:p-4">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <div><h2 className="text-sm font-semibold">Available times</h2><p className="text-[10px] text-muted-foreground">{hold ? `Held for ${Math.floor(holdSeconds / 60)}:${String(holdSeconds % 60).padStart(2, '0')}` : 'Select a slot to hold it while you finish booking.'}</p></div>
              {holdLoading && <span className="inline-flex items-center gap-1 text-[10px] text-muted-foreground"><LoaderCircle size={12} className="animate-spin" aria-hidden="true" /> Holding…</span>}
            </div>
            {availabilityLoading ? <LoadingState label="Checking availability…" /> : availabilityError ? <ErrorState message={availabilityError} onRetry={loadAvailability} /> : (availability?.slots.length ?? 0) === 0 ? <EmptyState title="No open times" description="This doctor has no free slots on this date. Try another day." /> : <div className="space-y-4">
              {[{ label: 'Morning', slots: groups.morning }, { label: 'Afternoon', slots: groups.afternoon }].map(group => group.slots.length > 0 && <div key={group.label}>
                <p className="mb-2 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">{group.label}</p>
                <div className="grid grid-cols-3 gap-1.5 sm:grid-cols-4">
                  {group.slots.map(entry => {
                    const active = selectedSlot?.start_time === entry.slot.start_time
                    return <button
                      key={entry.slot.start_time}
                      type="button"
                      disabled={holdLoading}
                      onClick={() => void chooseSlot(entry.slot)}
                      className={`flex h-9 items-center justify-center gap-1 rounded-lg border text-[11px] font-semibold transition disabled:cursor-wait ${active ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-border bg-background hover:border-neutral-400'}`}
                    >
                      {active && <Check size={11} aria-hidden="true" />}
                      {entry.label}
                    </button>
                  })}
                </div>
              </div>)}
            </div>}
            <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-border pt-3">
              <Button variant="outline" size="sm" className="gap-1.5" onClick={() => { if (reschedule) onNavigate(backPath); else setStep(0) }}><ArrowLeft size={13} aria-hidden="true" /> {reschedule ? 'Cancel' : 'Back'}</Button>
              <Button size="sm" className="ml-auto gap-1.5" disabled={!selectedSlot || holdLoading} onClick={() => setStep(2)}>Review booking <ArrowRight size={13} aria-hidden="true" /></Button>
            </div>
          </div>
        </div>
      </section>}


      {step === 2 && selectedSlot && doctor && <section aria-label="Confirm booking" className="grid gap-4 lg:grid-cols-2">
        <div className="rounded-2xl border border-border bg-background p-4 sm:p-5">
          <h2 className="text-base font-semibold tracking-tight">Appointment summary</h2>
          <div className="mt-4 flex items-center gap-3 rounded-xl bg-muted/50 p-3">
            <Avatar name={`${doctor.first_name} ${doctor.last_name}`} src={doctor.avatar_url} size="md" />
            <div><p className="text-sm font-semibold">Dr. {doctor.first_name} {doctor.last_name}</p><p className="text-[11px] text-muted-foreground">{doctor.specialty}</p></div>
          </div>
          <dl className="mt-4 space-y-3 text-xs">
            <div className="flex items-start justify-between gap-3 border-b border-border pb-3"><dt className="text-muted-foreground">Date</dt><dd className="text-right font-medium">{formatDate(selectedSlot.start_time, timeZone)}</dd></div>
            <div className="flex items-start justify-between gap-3 border-b border-border pb-3"><dt className="text-muted-foreground">{reschedule ? 'New time' : 'Time'}</dt><dd className="text-right font-medium">{formatTimeRange(selectedSlot.start_time, selectedSlot.end_time, timeZone)}</dd></div>
            {reschedule && <div className="flex items-start justify-between gap-3 border-b border-border pb-3"><dt className="text-muted-foreground">Current time</dt><dd className="text-right text-muted-foreground">{formatTimeRange(reschedule.start_time, reschedule.end_time, timeZone)}</dd></div>}
            <div className="flex items-start justify-between gap-3"><dt className="text-muted-foreground">Duration</dt><dd className="text-right font-medium">{doctor.slot_minutes} minutes</dd></div>
          </dl>
          {hold && <p className="mt-4 flex items-center gap-1.5 rounded-lg bg-amber-50 px-3 py-2 text-[10px] text-amber-700"><Clock3 size={12} aria-hidden="true" /> Slot held for {Math.floor(holdSeconds / 60)}:{String(holdSeconds % 60).padStart(2, '0')} · times in {timeZone.replace('_', ' ')}</p>}
        </div>

        <div className="rounded-2xl border border-border bg-background p-4 sm:p-5">
          <h2 className="text-base font-semibold tracking-tight">Visit details</h2>
          <p className="mt-1 text-[11px] leading-relaxed text-muted-foreground">Share a short reason so the care team can prepare. The clinic can also see this on the appointment.</p>
          <div className="mt-4">
            <Textarea
              label="Reason for visit"
              placeholder="e.g. Annual wellness checkup and bloodwork review"
              value={reason}
              onChange={event => setReason(event.target.value)}
              rows={5}
              required
            />
          </div>
          {error && <div className="mt-3 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
          <div className="mt-4 flex flex-wrap gap-2 border-t border-border pt-3">
            <Button variant="outline" size="sm" className="gap-1.5" onClick={() => setStep(1)} disabled={submitting}><ArrowLeft size={13} aria-hidden="true" /> Back</Button>
            <Button size="sm" className="ml-auto gap-1.5" onClick={() => void confirm()} disabled={submitting || (!hold && !reschedule)}>
              {submitting ? <LoaderCircle size={13} className="animate-spin" aria-hidden="true" /> : <Stethoscope size={13} aria-hidden="true" />}
              {submitting ? 'Saving…' : reschedule ? 'Confirm reschedule' : 'Confirm appointment'}
            </Button>
          </div>
        </div>
      </section>}
    </div>
  </main>
}
