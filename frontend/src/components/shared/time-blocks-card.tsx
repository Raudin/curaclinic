import { AlertCircle, CalendarOff, LoaderCircle, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/shared/page'
import { formatDate, formatTime } from '@/lib/time'
import type { ScheduleBlock } from '@/lib/api-types'

export function TimeBlocksCard({ blocks, timezone, start, end, reason, adding, error, canAdd = true, canRemove = true, onStart, onEnd, onReason, onAdd, onRemove }: {
  blocks: ScheduleBlock[]
  timezone: string
  start: string
  end: string
  reason: string
  adding: boolean
  error: string | null
  canAdd?: boolean
  canRemove?: boolean
  onStart: (value: string) => void
  onEnd: (value: string) => void
  onReason: (value: string) => void
  onAdd: () => void
  onRemove: (id: string) => void
}) {
  return <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-5">
    <div className="mb-3 flex items-center gap-2"><CalendarOff size={16} className="text-primary" aria-hidden="true" /><div><h2 className="text-base font-semibold">Time blocks</h2><p className="text-[11px] text-muted-foreground">Lunch, leave or other unavailability.</p></div></div>
    {error && <div className="mb-3 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-[11px] text-destructive"><AlertCircle size={13} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
    {canAdd && <div className="grid gap-2">
      <input type="datetime-local" value={start} onChange={event => onStart(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-[11px]" aria-label="Block start" />
      <input type="datetime-local" value={end} onChange={event => onEnd(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-[11px]" aria-label="Block end" />
      <input value={reason} onChange={event => onReason(event.target.value)} placeholder="Reason (e.g. lunch, sick leave)" className="h-9 rounded-md border border-border bg-background px-2 text-[11px]" aria-label="Block reason" />
      <Button size="sm" className="gap-1.5" onClick={onAdd} disabled={adding}>{adding ? <LoaderCircle size={13} className="animate-spin" aria-hidden="true" /> : <Plus size={13} aria-hidden="true" />} Add block</Button>
    </div>}
    <div className="mt-4 space-y-2">
      {blocks.length === 0 ? <EmptyState title="No blocks" description="Blocked time will appear here." /> : blocks.map(block => <div key={block.id} className="flex items-start justify-between gap-2 rounded-xl bg-muted/50 px-3 py-2.5">
        <div className="min-w-0"><p className="truncate text-[11px] font-semibold">{block.reason}</p><p className="mt-0.5 text-[10px] text-muted-foreground">{formatDate(block.start_time, timezone)} · {formatTime(block.start_time, timezone)}–{formatTime(block.end_time, timezone)}</p></div>
        {canRemove && <Button variant="ghost" size="icon" className="size-7 shrink-0 text-muted-foreground hover:text-destructive" aria-label="Remove block" onClick={() => onRemove(block.id)}><Trash2 size={13} aria-hidden="true" /></Button>}
      </div>)}
    </div>
  </section>
}
