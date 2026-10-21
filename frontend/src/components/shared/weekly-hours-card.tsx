import { useState } from 'react'
import { AlertCircle, Check, Copy, LoaderCircle, Plus, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { weekdayName } from '@/lib/time'

export type ScheduleInterval = { weekday: number; start_time: string; end_time: string }

const WEEK = [1, 2, 3, 4, 5, 6, 0]
const WEEKDAYS = [1, 2, 3, 4, 5]

export function WeeklyHoursCard({ intervals, slotMinutes, timezone, saving, error, canEdit = true, onAdd, onUpdate, onRemove, onCopy, onSave }: {
  intervals: ScheduleInterval[]
  slotMinutes: number
  timezone: string
  saving: boolean
  error: string | null
  canEdit?: boolean
  onAdd: (weekday: number) => void
  onUpdate: (index: number, patch: Partial<ScheduleInterval>) => void
  onRemove: (index: number) => void
  /** Replaces the hours on each target day with a copy of the source day's hours. */
  onCopy: (fromWeekday: number, toWeekdays: number[]) => void
  onSave: () => void
}) {
  const [copyFrom, setCopyFrom] = useState<number | null>(null)
  const [targets, setTargets] = useState<number[]>([])
  const [copied, setCopied] = useState<{ from: number; to: number[] } | null>(null)

  function startCopy(weekday: number) {
    setCopied(null)
    if (copyFrom === weekday) { setCopyFrom(null); return }
    setCopyFrom(weekday)
    setTargets([])
  }

  function toggleTarget(weekday: number) {
    setTargets(current => current.includes(weekday) ? current.filter(day => day !== weekday) : [...current, weekday])
  }

  function applyCopy() {
    if (copyFrom === null || targets.length === 0) return
    onCopy(copyFrom, targets)
    setCopied({ from: copyFrom, to: targets })
    setCopyFrom(null)
    setTargets([])
  }

  return <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
    <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
      <div><h2 className="text-base font-semibold">Weekly recurring schedule</h2><p className="text-[11px] text-muted-foreground">{slotMinutes}-minute increments · {timezone.replace('_', ' ')}{!canEdit ? ' · Read-only' : ''}</p></div>
      {canEdit && <Button size="sm" className="gap-1.5" onClick={onSave} disabled={saving}>{saving ? <LoaderCircle size={13} className="animate-spin" aria-hidden="true" /> : <Save size={13} aria-hidden="true" />} Save hours</Button>}
    </div>
    {error && <div className="mb-3 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
    <div className="space-y-2">
      {WEEK.map(weekday => {
        const dayIntervals = intervals.map((interval, index) => ({ interval, index })).filter(item => item.interval.weekday === weekday)
        return <div key={weekday} className="rounded-xl border border-border/80 p-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-xs font-semibold">{weekdayName(weekday)}</p>
            {canEdit && <div className="flex items-center gap-1">
              {dayIntervals.length > 0 && <Button variant="ghost" size="sm" className="h-7 gap-1 text-[10px]" aria-expanded={copyFrom === weekday} onClick={() => startCopy(weekday)}><Copy size={11} aria-hidden="true" /> Copy to…</Button>}
              <Button variant="ghost" size="sm" className="h-7 gap-1 text-[10px]" onClick={() => onAdd(weekday)}><Plus size={11} aria-hidden="true" /> Add hours</Button>
            </div>}
          </div>
          {dayIntervals.length === 0 ? <p className="mt-1 text-[10px] text-muted-foreground">Not working</p> : <div className="mt-2 space-y-1.5">
            {dayIntervals.map(({ interval, index }) => <div key={index} className="flex flex-wrap items-center gap-1.5">
              <input type="time" value={interval.start_time} step={slotMinutes * 60} disabled={!canEdit} onChange={event => onUpdate(index, { start_time: event.target.value })} className="h-8 rounded-md border border-border bg-background px-2 text-[11px] disabled:opacity-60" aria-label={`${weekdayName(weekday)} start time`} />
              <span className="text-[10px] text-muted-foreground">to</span>
              <input type="time" value={interval.end_time} step={slotMinutes * 60} disabled={!canEdit} onChange={event => onUpdate(index, { end_time: event.target.value })} className="h-8 rounded-md border border-border bg-background px-2 text-[11px] disabled:opacity-60" aria-label={`${weekdayName(weekday)} end time`} />
              {canEdit && <Button variant="ghost" size="icon" className="size-7 text-muted-foreground hover:text-destructive" onClick={() => onRemove(index)} aria-label="Remove interval"><Trash2 size={13} aria-hidden="true" /></Button>}
            </div>)}
          </div>}
          {copyFrom === weekday && <CopyPanel
            from={weekday}
            targets={targets}
            onToggle={toggleTarget}
            onSet={setTargets}
            onApply={applyCopy}
            onCancel={() => setCopyFrom(null)}
          />}
          {copied?.from === weekday && copyFrom === null && <p className="mt-2 flex items-center gap-1 text-[10px] text-primary" role="status"><Check size={11} aria-hidden="true" /> Copied to {copied.to.length === 1 ? weekdayName(copied.to[0]!) : `${copied.to.length} days`}. Save hours to publish the change.</p>}
        </div>
      })}
    </div>
  </section>
}

function CopyPanel({ from, targets, onToggle, onSet, onApply, onCancel }: {
  from: number
  targets: number[]
  onToggle: (weekday: number) => void
  onSet: (weekdays: number[]) => void
  onApply: () => void
  onCancel: () => void
}) {
  const others = WEEK.filter(day => day !== from)
  const presets = [
    { label: 'Weekdays', days: WEEKDAYS.filter(day => day !== from) },
    { label: 'Every day', days: others },
  ]
  return <div className="mt-3 rounded-lg bg-muted/50 p-3">
    <p className="text-[11px] font-medium">Copy {weekdayName(from)}’s hours to</p>
    <div className="mt-2 flex flex-wrap gap-1.5" role="group" aria-label="Days to copy to">
      {others.map(day => {
        const on = targets.includes(day)
        return <button
          key={day}
          type="button"
          aria-pressed={on}
          onClick={() => onToggle(day)}
          className={`h-7 min-w-11 rounded-md border px-2 text-[10px] font-semibold transition ${on ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background hover:border-foreground/30'}`}
        >{weekdayName(day).slice(0, 3)}</button>
      })}
    </div>
    <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10px]">
      {presets.map(preset => <button key={preset.label} type="button" className="font-semibold text-primary hover:underline" onClick={() => onSet(preset.days)}>{preset.label}</button>)}
      {targets.length > 0 && <button type="button" className="text-muted-foreground hover:underline" onClick={() => onSet([])}>Clear</button>}
    </div>
    <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-border/70 pt-2">
      <p className="text-[10px] text-muted-foreground">Replaces any hours already set on those days.</p>
      <div className="flex gap-1.5">
        <Button variant="ghost" size="sm" className="h-7 text-[10px]" onClick={onCancel}>Cancel</Button>
        <Button size="sm" className="h-7 text-[10px]" disabled={targets.length === 0} onClick={onApply}>Apply to {targets.length || ''} {targets.length === 1 ? 'day' : 'days'}</Button>
      </div>
    </div>
  </div>
}
