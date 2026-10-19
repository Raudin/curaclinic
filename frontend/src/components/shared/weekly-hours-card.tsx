import { AlertCircle, LoaderCircle, Plus, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { weekdayName } from '@/lib/time'

export type ScheduleInterval = { weekday: number; start_time: string; end_time: string }

export function WeeklyHoursCard({ intervals, slotMinutes, timezone, saving, error, canEdit = true, onAdd, onUpdate, onRemove, onSave }: {
  intervals: ScheduleInterval[]
  slotMinutes: number
  timezone: string
  saving: boolean
  error: string | null
  canEdit?: boolean
  onAdd: (weekday: number) => void
  onUpdate: (index: number, patch: Partial<ScheduleInterval>) => void
  onRemove: (index: number) => void
  onSave: () => void
}) {
  return <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
    <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
      <div><h2 className="text-base font-semibold">Weekly recurring schedule</h2><p className="text-[11px] text-muted-foreground">{slotMinutes}-minute increments · {timezone.replace('_', ' ')}{!canEdit ? ' · Read-only' : ''}</p></div>
      {canEdit && <Button size="sm" className="gap-1.5" onClick={onSave} disabled={saving}>{saving ? <LoaderCircle size={13} className="animate-spin" aria-hidden="true" /> : <Save size={13} aria-hidden="true" />} Save hours</Button>}
    </div>
    {error && <div className="mb-3 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
    <div className="space-y-2">
      {[1, 2, 3, 4, 5, 6, 0].map(weekday => {
        const dayIntervals = intervals.map((interval, index) => ({ interval, index })).filter(item => item.interval.weekday === weekday)
        return <div key={weekday} className="rounded-xl border border-border/80 p-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-xs font-semibold">{weekdayName(weekday)}</p>
            {canEdit && <Button variant="ghost" size="sm" className="h-7 gap-1 text-[10px]" onClick={() => onAdd(weekday)}><Plus size={11} aria-hidden="true" /> Add hours</Button>}
          </div>
          {dayIntervals.length === 0 ? <p className="mt-1 text-[10px] text-muted-foreground">Not working</p> : <div className="mt-2 space-y-1.5">
            {dayIntervals.map(({ interval, index }) => <div key={index} className="flex flex-wrap items-center gap-1.5">
              <input type="time" value={interval.start_time} step={slotMinutes * 60} disabled={!canEdit} onChange={event => onUpdate(index, { start_time: event.target.value })} className="h-8 rounded-md border border-border bg-background px-2 text-[11px] disabled:opacity-60" aria-label={`${weekdayName(weekday)} start time`} />
              <span className="text-[10px] text-muted-foreground">to</span>
              <input type="time" value={interval.end_time} step={slotMinutes * 60} disabled={!canEdit} onChange={event => onUpdate(index, { end_time: event.target.value })} className="h-8 rounded-md border border-border bg-background px-2 text-[11px] disabled:opacity-60" aria-label={`${weekdayName(weekday)} end time`} />
              {canEdit && <Button variant="ghost" size="icon" className="size-7 text-muted-foreground hover:text-destructive" onClick={() => onRemove(index)} aria-label="Remove interval"><Trash2 size={13} aria-hidden="true" /></Button>}
            </div>)}
          </div>}
        </div>
      })}
    </div>
  </section>
}
