import { useEffect, useState, type FormEvent } from 'react'
import { AlertCircle, LoaderCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useToast } from '@/components/feedback/toast-provider'
import { ApiError } from '@/lib/api'
import { doctorsApi } from '@/lib/api-clinic'
import { rolesApi, usersApi } from '@/lib/api-admin'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import type { Doctor, Role } from '@/lib/api-types'
import { DEFAULT_TIMEZONE } from '@/lib/time'

export function DoctorFormDialog({ mode, doctor, open, onOpenChange, onSaved }: {
  mode: 'create' | 'edit'
  doctor?: Doctor | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}) {
  const { can } = useAuth()
  const { notify } = useToast()
  const canViewAccess = can(PERMISSIONS.adminUsersView)
  const canListRoles = can(PERMISSIONS.adminRolesManage)
  const canAssignRoles = can(PERMISSIONS.adminUsersAssignRoles)
  const showRoleEditor = canListRoles && canAssignRoles && (mode === 'create' || canViewAccess)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [phone, setPhone] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [bio, setBio] = useState('')
  const [experience, setExperience] = useState('0')
  const [languages, setLanguages] = useState('')
  const [timezone, setTimezone] = useState(DEFAULT_TIMEZONE)
  const [slotMinutes, setSlotMinutes] = useState('30')
  const [accepting, setAccepting] = useState(true)
  const [roles, setRoles] = useState<Role[]>([])
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([])
  const [rolesLoading, setRolesLoading] = useState(false)
  const [rolesError, setRolesError] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setError(null)
    if (mode === 'edit' && doctor) {
      setFirstName(doctor.first_name); setLastName(doctor.last_name); setEmail(doctor.email)
      setSpecialty(doctor.specialty); setBio(doctor.bio ?? ''); setExperience(String(doctor.experience_years))
      setLanguages(doctor.languages.join(', ')); setTimezone(doctor.timezone); setSlotMinutes(String(doctor.slot_minutes)); setAccepting(doctor.is_accepting_patients)
    } else {
      setFirstName(''); setLastName(''); setEmail(''); setPassword(''); setPhone(''); setSpecialty(''); setBio('')
      setExperience('0'); setLanguages(''); setTimezone(DEFAULT_TIMEZONE); setSlotMinutes('30'); setAccepting(true)
    }
  }, [doctor, mode, open])

  useEffect(() => {
    if (!open || !showRoleEditor) return
    let cancelled = false
    setRolesLoading(true)
    setRolesError(null)
    Promise.all([
      rolesApi.list(),
      mode === 'edit' && doctor ? usersApi.access(doctor.id) : Promise.resolve(null),
    ]).then(([roleList, access]) => {
      if (cancelled) return
      setRoles(roleList)
      setSelectedRoleIds(access
        ? roleList.filter(role => access.roles.includes(role.name)).map(role => role.id)
        : roleList.filter(role => role.name === 'doctor').map(role => role.id))
    }).catch(err => {
      if (!cancelled) setRolesError(err instanceof Error ? err.message : 'Unable to load roles.')
    }).finally(() => {
      if (!cancelled) setRolesLoading(false)
    })
    return () => { cancelled = true }
  }, [doctor, mode, open, showRoleEditor])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const profile = {
        specialty: specialty.trim(),
        bio: bio.trim(),
        experience_years: Number(experience) || 0,
        languages: languages.split(',').map(item => item.trim()).filter(Boolean),
        timezone,
        slot_minutes: Number(slotMinutes) || 30,
        is_accepting_patients: accepting,
      }
      if (mode === 'create') {
        const created = await doctorsApi.create({
          email: email.trim(), password, first_name: firstName.trim(), last_name: lastName.trim(), phone: phone.trim(),
          avatar_url: '', ...profile,
        })
        if (showRoleEditor) await usersApi.assignRoles(created.id, selectedRoleIds)
        notify('Doctor account created.')
      } else if (doctor) {
        await doctorsApi.updateProfile(doctor.id, profile)
        if (showRoleEditor) await usersApi.assignRoles(doctor.id, selectedRoleIds)
        notify('Doctor profile updated.')
      }
      onSaved()
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to save the doctor profile.')
    } finally {
      setSaving(false)
    }
  }

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="max-w-xl">
      <DialogHeader>
        <DialogTitle>{mode === 'create' ? 'Add a doctor' : 'Edit doctor profile'}</DialogTitle>
        <DialogDescription>{mode === 'create' ? 'Create the clinician account and public profile.' : `${doctor?.first_name ?? ''} ${doctor?.last_name ?? ''}`}</DialogDescription>
      </DialogHeader>
      {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div>}
      <form className="grid gap-3 sm:grid-cols-2" onSubmit={submit}>
        {mode === 'create' && <>
          <input value={firstName} onChange={event => setFirstName(event.target.value)} placeholder="First name" required className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
          <input value={lastName} onChange={event => setLastName(event.target.value)} placeholder="Last name" required className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
          <input value={email} onChange={event => setEmail(event.target.value)} type="email" placeholder="Email" required className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
          <input value={password} onChange={event => setPassword(event.target.value)} type="password" placeholder="Password (min 12)" minLength={12} required className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
          <input value={phone} onChange={event => setPhone(event.target.value)} placeholder="Phone (optional)" className="h-9 rounded-md border border-border bg-background px-2 text-xs sm:col-span-2" />
        </>}
        <input value={specialty} onChange={event => setSpecialty(event.target.value)} placeholder="Specialty" required className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
        <input value={languages} onChange={event => setLanguages(event.target.value)} placeholder="Languages (comma separated)" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
        <input value={experience} onChange={event => setExperience(event.target.value)} type="number" min={0} max={80} placeholder="Years of experience" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
        <input value={slotMinutes} onChange={event => setSlotMinutes(event.target.value)} type="number" min={5} max={180} step={5} placeholder="Slot minutes" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
        <input value={timezone} onChange={event => setTimezone(event.target.value)} placeholder="Timezone" className="h-9 rounded-md border border-border bg-background px-2 text-xs" />
        <label className="flex items-center justify-between gap-2 rounded-md border border-border px-2 text-xs">Accepting patients<input type="checkbox" checked={accepting} onChange={event => setAccepting(event.target.checked)} className="size-3.5 accent-neutral-950" /></label>
        <textarea value={bio} onChange={event => setBio(event.target.value)} placeholder="Professional bio" rows={3} className="rounded-md border border-border bg-background px-2 py-1.5 text-xs sm:col-span-2" />
        {showRoleEditor && <fieldset className="space-y-2 rounded-xl border border-border p-3 sm:col-span-2" disabled={rolesLoading || saving}>
          <legend className="px-1 text-xs font-semibold">Account roles</legend>
          <p className="text-[11px] text-muted-foreground">Choose the roles assigned to this account.</p>
          {rolesLoading ? <p className="text-xs text-muted-foreground">Loading roles…</p> : rolesError ? <p className="text-xs text-destructive">{rolesError}</p> : roles.length === 0 ? <p className="text-xs text-muted-foreground">No roles available.</p> : <div className="grid gap-2 sm:grid-cols-2">
            {roles.map(role => <label key={role.id} className="flex items-start gap-2 rounded-lg border border-border px-3 py-2 text-xs">
              <input type="checkbox" className="mt-0.5 size-3.5 accent-neutral-950" checked={selectedRoleIds.includes(role.id)} onChange={event => setSelectedRoleIds(current => event.target.checked ? [...current, role.id] : current.filter(id => id !== role.id))} />
              <span><span className="block font-medium capitalize">{role.name}</span><span className="block text-[10px] text-muted-foreground">{role.scope === 'doctor' ? 'Doctor scope' : 'Admin scope'}{role.description ? ` · ${role.description}` : ''}</span></span>
            </label>)}
          </div>}
        </fieldset>}
        <DialogFooter className="sm:col-span-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>Cancel</Button>
          <Button type="submit" disabled={saving || (showRoleEditor && (rolesLoading || Boolean(rolesError) || selectedRoleIds.length === 0))}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} {mode === 'create' ? 'Create doctor' : 'Save changes'}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
}

