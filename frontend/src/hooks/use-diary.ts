import { useCallback, useEffect, useState } from 'react'
import { appointmentsApi } from '@/lib/api-clinic'
import type { Appointment, AppointmentStatus } from '@/lib/api-types'
import { zonedDayBoundsFromKey } from '@/lib/time'

const EMPTY: Appointment[] = []
export const DIARY_PAGE_SIZE = 25

type DiaryOptions = {
  date: string
  timeZone: string
  doctorId?: string
  status?: AppointmentStatus
  offset?: number
}

// An omitted offset requests the clinician's entire day.
export function useDiary({ date, timeZone, doctorId, status, offset }: DiaryOptions) {
  const [revision, setRevision] = useState(0)
  const refresh = useCallback(() => setRevision(value => value + 1), [])
  const [now, setNow] = useState(Date.now)
  const key = JSON.stringify([date, timeZone, doctorId, status, offset, revision])
  const [result, setResult] = useState<{ key: string; items: Appointment[]; total: number; error: string | null } | null>(null)

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const bounds = zonedDayBoundsFromKey(date, timeZone)
        const items: Appointment[] = []
        const limit = offset === undefined ? 100 : DIARY_PAGE_SIZE
        let cursor = offset ?? 0
        let total: number
        do {
          const page = await appointmentsApi.list({ ...bounds, scope: 'all', doctorId, status, order: 'asc', limit, offset: cursor })
          if (!active) return
          items.push(...page.data)
          total = page.pagination.total
          cursor += page.data.length
          if (offset !== undefined || page.data.length === 0) break
        } while (cursor < total)
        setNow(Date.now())
        setResult({ key, items, total, error: null })
      } catch (err) {
        if (active) setResult({ key, items: EMPTY, total: 0, error: err instanceof Error ? err.message : 'Unable to load appointments.' })
      }
    }
    void load()
    return () => { active = false }
  }, [key, date, timeZone, doctorId, status, offset])

  const current = result?.key === key ? result : null
  const items = current?.items ?? EMPTY
  useEffect(() => {
    const time = Date.now()
    let next = Infinity
    for (const item of items) {
      const start = Date.parse(item.start_time)
      if (item.status === 'confirmed' && start > time) next = Math.min(next, start)
    }
    if (!Number.isFinite(next)) return
    const timer = window.setTimeout(() => setNow(Date.now()), Math.min(next - time + 1, 2147483647))
    return () => clearTimeout(timer)
  }, [items, now])

  return { items, now, total: current?.total ?? 0, error: current?.error, loading: !current, refresh }
}
