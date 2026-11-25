import { useEffect, useMemo, useState } from 'react'
import { ErrorState, LoadingState, PageHeader } from '@/components/shared/page'
import { MorphSelect } from '@/components/arc/morph-select/morph-select'
import { ScheduleEditor } from '@/components/shared/schedule-editor'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { doctorsApi } from '@/lib/api-clinic'
import type { Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE } from '@/lib/time'

export function AdminSchedulePage() {
  const { canAny } = useAuth()
  const canViewDoctors = canAny(PERMISSIONS.adminUsersView, PERMISSIONS.doctorsView, PERMISSIONS.doctorPatientsView)
  const initialDoctor = useMemo(() => new URLSearchParams(window.location.search).get('doctor') ?? '', [])
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [doctorId, setDoctorId] = useState(initialDoctor)

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    if (!canViewDoctors) return
    let active = true
    setLoading(true)
    setError(false)
    doctorsApi.list().then(list => {
      if (!active) return
      setDoctors(list)
      setDoctorId(current => list.some(item => item.id === current) ? current : list[0]?.id || '')
    }).catch(() => { if (active) setError(true) }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [canViewDoctors, attempt])

  const doctor = doctors.find(item => item.id === doctorId) ?? null

  return <main className="mx-auto w-full max-w-[1300px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader title="Schedule & availability" description="Manage recurring clinic hours and blocked time for any doctor." />
    {error ? <ErrorState message="Unable to load the doctor directory." onRetry={() => setAttempt(value => value + 1)} /> : loading ? <LoadingState label="Loading doctors…" /> : <div className="@container mb-4 rounded-2xl border border-border bg-card p-3 shadow-sm">
      <MorphSelect
        className="![--ms-panel-w:min(17rem,100cqi)]"
        label="Doctor"
        value={doctorId || null}
        items={doctors.map(item => ({
          value: item.id,
          label: `Dr. ${item.first_name} ${item.last_name}`,
          meta: item.specialty,
          keywords: item.email,
        }))}
        onValueChange={value => setDoctorId(value)}
        searchable
        searchPlaceholder="Search doctors"
        placeholder={doctors.length ? 'Select a doctor' : 'No doctors available'}
        disabled={doctors.length === 0}
      />
    </div>}
    {doctor && <ScheduleEditor
      key={doctorId}
      doctorId={doctorId}
      timezone={doctor?.timezone || DEFAULT_TIMEZONE}
      slotMinutes={doctor?.slot_minutes || 30}
      mode="admin"
    />}
  </main>
}
