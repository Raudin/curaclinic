import { useCallback, useEffect, useState } from 'react'
import { ErrorState, LoadingState } from '@/components/shared/page'
import { WeeklyHoursCard, type ScheduleInterval } from '@/components/shared/weekly-hours-card'
import { TimeBlocksCard } from '@/components/shared/time-blocks-card'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { scheduleApi } from '@/lib/api-admin'
import type { ScheduleBlock } from '@/lib/api-types'
import { DEFAULT_TIMEZONE } from '@/lib/time'

export function ScheduleEditor({ doctorId, timezone = DEFAULT_TIMEZONE, slotMinutes = 30, mode }: {
  doctorId: string
  timezone?: string
  slotMinutes?: number
  mode: 'self' | 'admin'
}) {
  const { can } = useAuth()
  const { notify } = useToast()
  const isAdminMode = mode === 'admin'
  const canEditHours = isAdminMode ? can(PERMISSIONS.adminWorkingHours) : can(PERMISSIONS.doctorWorkingHours)
  const canAddBlocks = isAdminMode ? can(PERMISSIONS.adminWorkingHours) : can(PERMISSIONS.doctorBlockTime)
  const canRemoveBlocks = isAdminMode ? can(PERMISSIONS.adminWorkingHours) : can(PERMISSIONS.doctorRemoveBlock)
  const [intervals, setIntervals] = useState<ScheduleInterval[]>([])
  const [blocks, setBlocks] = useState<ScheduleBlock[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [blockStart, setBlockStart] = useState('')
  const [blockEnd, setBlockEnd] = useState('')
  const [blockReason, setBlockReason] = useState('')
  const [addingBlock, setAddingBlock] = useState(false)
  const [blockError, setBlockError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      if (mode === 'self') {
        const schedule = await scheduleApi.mine()
        setIntervals(schedule.working_hours.map(hour => ({ weekday: hour.weekday, start_time: hour.start_time, end_time: hour.end_time })))
        setBlocks(schedule.blocks)
      } else {
        const [hours, blockList] = await Promise.all([
          scheduleApi.workingHours(doctorId),
          scheduleApi.blocks(doctorId),
        ])
        setIntervals(hours.intervals.map(hour => ({ weekday: hour.weekday, start_time: hour.start_time, end_time: hour.end_time })))
        setBlocks(blockList)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load the schedule.')
    } finally {
      setLoading(false)
    }
  }, [doctorId, mode])

  useEffect(() => { void load() }, [load])

  function addInterval(weekday: number) {
    setIntervals(current => [...current, { weekday, start_time: '09:00', end_time: '12:00' }])
  }

  function updateInterval(index: number, patch: Partial<ScheduleInterval>) {
    setIntervals(current => current.map((interval, position) => position === index ? { ...interval, ...patch } : interval))
  }

  function removeInterval(index: number) {
    setIntervals(current => current.filter((_, position) => position !== index))
  }

  async function saveIntervals() {
    setSaving(true)
    setError(null)
    try {
      if (intervals.some(interval => interval.start_time >= interval.end_time)) {
        setError('Each interval must end after it starts.')
        return
      }
      const result = mode === 'self'
        ? await scheduleApi.replaceMine(intervals)
        : await scheduleApi.replaceWorkingHours(doctorId, intervals)
      setIntervals(result.intervals.map(hour => ({ weekday: hour.weekday, start_time: hour.start_time, end_time: hour.end_time })))
      notify('Working hours saved.')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to save working hours.')
    } finally {
      setSaving(false)
    }
  }

  async function addBlock() {
    if (!blockStart || !blockEnd || !blockReason.trim()) {
      setBlockError('Choose a start, end and reason for the block.')
      return
    }
    if (new Date(blockEnd) <= new Date(blockStart)) {
      setBlockError('The block must end after it starts.')
      return
    }
    setAddingBlock(true)
    setBlockError(null)
    try {
      const payload = {
        start_time: new Date(blockStart).toISOString(),
        end_time: new Date(blockEnd).toISOString(),
        reason: blockReason.trim(),
      }
      const block = mode === 'self' ? await scheduleApi.createMyBlock(payload) : await scheduleApi.createBlock(doctorId, payload)
      setBlocks(current => [...current, block].sort((a, b) => a.start_time.localeCompare(b.start_time)))
      setBlockStart(''); setBlockEnd(''); setBlockReason('')
      notify('Time block added.')
    } catch (err) {
      setBlockError(err instanceof ApiError ? err.message : 'Unable to add this block.')
    } finally {
      setAddingBlock(false)
    }
  }

  async function removeBlock(id: string) {
    try {
      if (mode === 'self') await scheduleApi.deleteMyBlock(id)
      else await scheduleApi.deleteBlock(id)
      setBlocks(current => current.filter(block => block.id !== id))
      notify('Time block removed.')
    } catch (err) {
      notify(err instanceof ApiError ? err.message : 'Unable to remove this block.', 'Could not remove block')
    }
  }

  if (loading) return <LoadingState label="Loading schedule…" />
  if (error && intervals.length === 0 && blocks.length === 0) return <ErrorState message={error} onRetry={load} />

  return <div className="grid gap-4 xl:grid-cols-[minmax(0,1.5fr)_minmax(300px,0.8fr)]">
    <WeeklyHoursCard
      intervals={intervals}
      slotMinutes={slotMinutes}
      timezone={timezone}
      saving={saving}
      error={error}
      canEdit={canEditHours}
      onAdd={addInterval}
      onUpdate={updateInterval}
      onRemove={removeInterval}
      onSave={() => void saveIntervals()}
    />
    <TimeBlocksCard
      blocks={blocks}
      timezone={timezone}
      start={blockStart}
      end={blockEnd}
      reason={blockReason}
      adding={addingBlock}
      error={blockError}
      canAdd={canAddBlocks}
      canRemove={canRemoveBlocks}
      onStart={setBlockStart}
      onEnd={setBlockEnd}
      onReason={setBlockReason}
      onAdd={() => void addBlock()}
      onRemove={id => void removeBlock(id)}
    />
  </div>
}
