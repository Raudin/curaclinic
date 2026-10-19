import { useCallback, useEffect, useState } from 'react'
import { CalendarDays, Pencil, Plus, Power, Stethoscope } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState, ErrorState, LoadingState, PageHeader } from '@/components/shared/page'
import { TeamDirectory } from '@/components/arc/blocks/team-directory/team-directory'
import { DoctorFormDialog } from '@/components/admin/doctor-form-dialog'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError } from '@/lib/api'
import { doctorsApi } from '@/lib/api-clinic'
import type { Doctor } from '@/lib/api-types'

export function AdminDoctorsPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { can } = useAuth()
  const { notify } = useToast()
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [editing, setEditing] = useState<Doctor | null>(null)
  const [creating, setCreating] = useState(false)

  const canCreate = can(PERMISSIONS.adminUsersCreate)
  const canUpdate = can(PERMISSIONS.adminUsersUpdate)
  const canSchedule = can(PERMISSIONS.adminWorkingHours)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setDoctors(await doctorsApi.list())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load doctors.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  async function toggleAccepting(doctor: Doctor) {
    try {
      await doctorsApi.updateProfile(doctor.id, {
        specialty: doctor.specialty,
        bio: doctor.bio ?? '',
        experience_years: doctor.experience_years,
        languages: doctor.languages,
        avatar_url: doctor.avatar_url ?? '',
        timezone: doctor.timezone,
        slot_minutes: doctor.slot_minutes,
        is_accepting_patients: !doctor.is_accepting_patients,
      })
      notify(`${doctor.first_name} ${doctor.last_name} is ${doctor.is_accepting_patients ? 'no longer' : 'now'} accepting patients.`)
      await load()
    } catch (err) {
      notify(err instanceof ApiError ? err.message : 'Unable to update this doctor.', 'Update failed')
    }
  }

  return <main className="mx-auto w-full max-w-[1300px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader eyebrow="Clinic" title="Doctors" description="Manage clinician profiles, availability and the public directory." actions={canCreate ? <Button className="gap-2" onClick={() => setCreating(true)}><Plus size={15} aria-hidden="true" /> Add doctor</Button> : undefined} />
    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading doctors…" /> : null}
    {!loading && !error && doctors.length === 0 ? <EmptyState title="No doctors yet" description="Add a doctor to start building the clinic directory." action={canCreate ? <Button className="gap-2" onClick={() => setCreating(true)}><Plus size={15} aria-hidden="true" /> Add doctor</Button> : undefined} /> : null}
    {!loading && !error && doctors.length > 0 && <TeamDirectory
      people={doctors.map(doctor => ({
        id: doctor.id,
        name: `Dr. ${doctor.first_name} ${doctor.last_name}`,
        role: doctor.specialty,
        team: doctor.specialty,
        email: doctor.email,
        timezone: doctor.timezone,
        focus: `${doctor.experience_years} years in practice · ${doctor.slot_minutes}-minute visits${doctor.languages.length ? ` · ${doctor.languages.join(', ')}` : ''}`,
        bio: doctor.bio,
        avatarUrl: doctor.avatar_url,
        available: doctor.is_accepting_patients,
      }))}
      renderActions={person => {
        const doctor = doctors.find(item => item.id === person.id)
        if (!doctor) return null
        return <>
          {canUpdate && <>
            <Button variant="outline" size="sm" className="gap-1" onClick={() => setEditing(doctor)}><Pencil size={12} aria-hidden="true" /> Edit</Button>
            <Button variant="ghost" size="sm" className="gap-1" onClick={() => void toggleAccepting(doctor)}><Power size={12} aria-hidden="true" /> {doctor.is_accepting_patients ? 'Pause' : 'Resume'}</Button>
          </>}
          {canSchedule && <Button variant="ghost" size="sm" className="gap-1" onClick={() => onNavigate(`/admin/schedule?doctor=${doctor.id}`)}><CalendarDays size={12} aria-hidden="true" /> Schedule</Button>}
          {!canUpdate && !canSchedule && <span className="inline-flex items-center gap-1 text-xs text-muted-foreground"><Stethoscope size={13} aria-hidden="true" /> Read-only directory</span>}
        </>
      }}
    />}
    <DoctorFormDialog mode="create" open={creating} onOpenChange={setCreating} onSaved={load} />
    <DoctorFormDialog mode="edit" doctor={editing} open={Boolean(editing)} onOpenChange={open => { if (!open) setEditing(null) }} onSaved={load} />
  </main>
}
