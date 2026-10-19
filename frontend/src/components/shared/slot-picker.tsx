import { useCallback, useEffect, useState } from 'react'
import { CalendarDays, Clock3 } from 'lucide-react'
import { Calendar } from '@/components/arc/calendar/calendar'
import { ErrorState, LoadingState } from '@/components/shared/page'
import { availabilityApi } from '@/lib/api-clinic'
import { ApiError } from '@/lib/api'
import type { Availability, Slot } from '@/lib/api-types'
import { DEFAULT_TIMEZONE, addDays, dateFromKey, dateKey, dateKeyInZone, formatDate, formatTime } from '@/lib/time'

export function SlotPicker({ doctorId, selected, onSelect }: {
  doctorId: string
  selected: Slot | null
  onSelect: (slot: Slot | null) => void
}) {
  const [date, setDate] = useState(() => dateKeyInZone(new Date(), DEFAULT_TIMEZONE))
  const [availability, setAvailability] = useState<Availability | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const timeZone = availability?.timezone || DEFAULT_TIMEZONE

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await availabilityApi.get(doctorId, date)
      setAvailability(result)
    } catch (err) {
      setAvailability(null)
      setError(err instanceof ApiError ? err.message : 'Unable to load availability.')
    } finally {
      setLoading(false)
    }
  }, [date, doctorId])

  useEffect(() => { void load() }, [load])

  const minDate = dateFromKey(dateKeyInZone(new Date(), timeZone))
  const maxDate = addDays(minDate, 90)
  const morning = (availability?.slots ?? []).filter(slot => Number(formatTime(slot.start_time, timeZone).slice(0, 2)) < 12)
  const afternoon = (availability?.slots ?? []).filter(slot => Number(formatTime(slot.start_time, timeZone).slice(0, 2)) >= 12)

  return <div className="grid gap-3 sm:grid-cols-[260px_minmax(0,1fr)]">
    <div className="rounded-xl border border-border p-2">
      <Calendar value={dateFromKey(date)} onChange={value => { setDate(dateKey(value)); onSelect(null) }} minDate={minDate} maxDate={maxDate} locale="en-KE" />
      <p className="mt-1 flex items-center gap-1 px-1 text-[9px] text-muted-foreground"><CalendarDays size={10} aria-hidden="true" /> {formatDate(`${date}T12:00:00Z`, timeZone)}</p>
    </div>
    <div className="min-h-[220px] rounded-xl border border-border p-3">
      {loading ? <LoadingState label="Checking availability…" /> : error ? <ErrorState message={error} onRetry={load} /> : (availability?.slots.length ?? 0) === 0 ? <p className="py-8 text-center text-xs text-muted-foreground">No open times on this date.</p> : <div className="space-y-3">
        {[{ label: 'Morning', slots: morning }, { label: 'Afternoon', slots: afternoon }].map(group => group.slots.length > 0 && <div key={group.label}>
          <p className="mb-1.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">{group.label}</p>
          <div className="grid grid-cols-3 gap-1.5">
            {group.slots.map(slot => {
              const active = selected?.start_time === slot.start_time
              return <button key={slot.start_time} type="button" onClick={() => onSelect(active ? null : slot)} className={`h-8 rounded-md border text-[10px] font-semibold ${active ? 'border-neutral-950 bg-neutral-950 text-white' : 'border-border hover:border-neutral-400'}`}>{formatTime(slot.start_time, timeZone)}</button>
            })}
          </div>
        </div>)}
        <p className="flex items-center gap-1 text-[9px] text-muted-foreground"><Clock3 size={10} aria-hidden="true" /> {timeZone.replace('_', ' ')}</p>
      </div>}
    </div>
  </div>
}
