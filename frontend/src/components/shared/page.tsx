import type { ReactNode } from 'react'
import { LoaderCircle, TriangleAlert } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function PageHeader({ eyebrow, title, description, actions }: {
  eyebrow?: string
  title: string
  description?: string
  actions?: ReactNode
}) {
  return <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
    <div>
      {eyebrow && <p className="text-xs font-medium text-primary">{eyebrow}</p>}
      <h1 className="mt-1 text-2xl font-semibold tracking-[-0.035em] text-foreground sm:text-3xl">{title}</h1>
      {description && <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">{description}</p>}
    </div>
    {actions && <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>}
  </header>
}

export function LoadingState({ label = 'Loading…' }: { label?: string }) {
  return <div className="flex items-center justify-center gap-2 rounded-2xl border border-dashed border-border bg-card/60 px-6 py-12 text-sm text-muted-foreground">
    <LoaderCircle size={16} className="animate-spin" aria-hidden="true" />
    {label}
  </div>
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return <div className="flex flex-col items-center gap-3 rounded-2xl border border-destructive/25 bg-destructive/5 px-6 py-10 text-center">
    <span className="flex size-10 items-center justify-center rounded-full bg-destructive/10 text-destructive"><TriangleAlert size={18} aria-hidden="true" /></span>
    <div>
      <p className="text-sm font-semibold text-foreground">Something went wrong</p>
      <p className="mt-1 max-w-md text-xs leading-relaxed text-muted-foreground">{message}</p>
    </div>
    {onRetry && <Button variant="outline" size="sm" onClick={onRetry}>Try again</Button>}
  </div>
}

export function EmptyState({ title, description, action }: { title: string; description?: string; action?: ReactNode }) {
  return <div className="flex flex-col items-center gap-3 rounded-2xl border border-dashed border-border bg-card/50 px-6 py-12 text-center">
    <div>
      <p className="text-sm font-semibold text-foreground">{title}</p>
      {description && <p className="mt-1 max-w-md text-xs leading-relaxed text-muted-foreground">{description}</p>}
    </div>
    {action}
  </div>
}

export function Pagination({ limit, offset, total, onChange }: {
  limit: number
  offset: number
  total: number
  onChange: (offset: number) => void
}) {
  if (total <= limit) return null
  const start = total === 0 ? 0 : offset + 1
  const end = Math.min(offset + limit, total)
  return <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground">
    <span>Showing {start}–{end} of {total}</span>
    <div className="flex gap-2">
      <Button variant="outline" size="sm" disabled={offset === 0} onClick={() => onChange(Math.max(0, offset - limit))}>Previous</Button>
      <Button variant="outline" size="sm" disabled={end >= total} onClick={() => onChange(offset + limit)}>Next</Button>
    </div>
  </div>
}

export function StatusBadge({ status }: { status: string }) {
  const styles = status === 'cancelled'
    ? 'bg-destructive/10 text-destructive'
    : 'bg-emerald-50 text-emerald-700'
  return <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold capitalize ${styles}`}>{status}</span>
}
