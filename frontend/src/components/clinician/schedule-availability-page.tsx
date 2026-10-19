import { useEffect, useState } from 'react'
import { PageHeader } from '@/components/shared/page'
import { ScheduleEditor } from '@/components/shared/schedule-editor'
import { useAuth } from '@/lib/auth-context'
import { doctorsApi } from '@/lib/api-clinic'
import type { Doctor } from '@/lib/api-types'
import { DEFAULT_TIMEZONE } from '@/lib/time'

export function ScheduleAvailabilityPage() {
  const { user } = useAuth()
  const [doctor, setDoctor] = useState<Doctor | null>(null)

  useEffect(() => {
    if (!user) return
    let cancelled = false
    doctorsApi.get(user.id).then(profile => { if (!cancelled) setDoctor(profile) }).catch(() => undefined)
    return () => { cancelled = true }
  }, [user])

  return <main className="mx-auto w-full max-w-[1300px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader
      eyebrow="Availability"
      title="Schedule & availability"
      description="Set the hours patients can book and block time for lunch, leave or other commitments."
    />
    {user && <ScheduleEditor
      doctorId={user.id}
      timezone={doctor?.timezone || DEFAULT_TIMEZONE}
      slotMinutes={doctor?.slot_minutes || 30}
      mode="self"
    />}
  </main>
}
