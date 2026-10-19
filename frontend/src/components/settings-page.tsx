import { useEffect, useState, type FormEvent } from 'react'
import { AlertCircle, ImagePlus, LoaderCircle, Save, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/arc/tabs/tabs'
import { Input } from '@/components/arc/input/input'
import { Textarea } from '@/components/arc/textarea/textarea'
import { PasswordField } from '@/components/arc/password-field/password-field'
import { PasswordStrength } from '@/components/arc/password-strength/password-strength'
import { PageHeader, LoadingState } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { ApiError, apiRequest } from '@/lib/api'
import { doctorsApi } from '@/lib/api-clinic'
import type { Doctor } from '@/lib/api-types'

type SettingsVariant = 'patient' | 'clinician' | 'admin'

function errorMessage(err: unknown, fallback: string) {
  return err instanceof ApiError ? err.message : fallback
}

export function SettingsPage({ variant, onNavigate }: { variant: SettingsVariant; onNavigate: (path: string) => void }) {
  const { user, can, refreshAccount, changePassword } = useAuth()
  const { notify } = useToast()
  const canEditPractice = can(PERMISSIONS.doctorProfileUpdateOwn) || can(PERMISSIONS.adminUsersUpdate)
  const [firstName, setFirstName] = useState(user?.first_name ?? '')
  const [lastName, setLastName] = useState(user?.last_name ?? '')
  const [phone, setPhone] = useState(user?.phone ?? '')
  const [savingProfile, setSavingProfile] = useState(false)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [photoPreview, setPhotoPreview] = useState<string | null>(null)

  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [savingPassword, setSavingPassword] = useState(false)

  const [profile, setProfile] = useState<Doctor | null>(null)
  const [practiceLoading, setPracticeLoading] = useState(variant === 'clinician')
  const [practiceError, setPracticeError] = useState<string | null>(null)
  const [specialty, setSpecialty] = useState('')
  const [bio, setBio] = useState('')
  const [experienceYears, setExperienceYears] = useState('0')
  const [languages, setLanguages] = useState('')
  const [timezone, setTimezone] = useState('Africa/Nairobi')
  const [slotMinutes, setSlotMinutes] = useState('30')
  const [accepting, setAccepting] = useState(true)
  const [savingPractice, setSavingPractice] = useState(false)

  useEffect(() => {
    setFirstName(user?.first_name ?? '')
    setLastName(user?.last_name ?? '')
    setPhone(user?.phone ?? '')
  }, [user])

  useEffect(() => () => {
    if (photoPreview) URL.revokeObjectURL(photoPreview)
  }, [photoPreview])

  useEffect(() => {
    if (variant !== 'clinician' || !user || !canEditPractice) return
    let cancelled = false
    setPracticeLoading(true)
    doctorsApi.get(user.id).then(doctor => {
      if (cancelled) return
      setProfile(doctor)
      setSpecialty(doctor.specialty)
      setBio(doctor.bio ?? '')
      setExperienceYears(String(doctor.experience_years))
      setLanguages(doctor.languages.join(', '))
      setTimezone(doctor.timezone)
      setSlotMinutes(String(doctor.slot_minutes))
      setAccepting(doctor.is_accepting_patients)
    }).catch(err => {
      if (!cancelled) setPracticeError(errorMessage(err, 'Unable to load your practice profile.'))
    }).finally(() => {
      if (!cancelled) setPracticeLoading(false)
    })
    return () => { cancelled = true }
  }, [canEditPractice, user, variant])

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSavingProfile(true)
    setProfileError(null)
    try {
      await apiRequest('me/profile', {
        method: 'PATCH',
        body: JSON.stringify({ first_name: firstName.trim(), last_name: lastName.trim(), phone: phone.trim() }),
      })
      await refreshAccount()
      notify('Your profile has been updated.')
    } catch (err) {
      setProfileError(errorMessage(err, 'Unable to save your profile.'))
    } finally {
      setSavingProfile(false)
    }
  }

  async function savePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPasswordError(null)
    if (newPassword !== confirmPassword) {
      setPasswordError('The new passwords do not match.')
      return
    }
    if (newPassword.length < 12) {
      setPasswordError('Your new password must be at least 12 characters.')
      return
    }
    setSavingPassword(true)
    try {
      await changePassword(currentPassword, newPassword)
      notify('Your password has been changed. Please sign in again.', 'Password updated')
      onNavigate('/login')
    } catch (err) {
      setPasswordError(errorMessage(err, 'Unable to update your password.'))
    } finally {
      setSavingPassword(false)
    }
  }

  async function savePractice(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!user) return
    setSavingPractice(true)
    setPracticeError(null)
    try {
      const updated = await doctorsApi.updateProfile(user.id, {
        specialty: specialty.trim(),
        bio: bio.trim(),
        experience_years: Number(experienceYears) || 0,
        languages: languages.split(',').map(item => item.trim()).filter(Boolean),
        timezone,
        slot_minutes: Number(slotMinutes) || 30,
        is_accepting_patients: accepting,
      })
      setProfile(updated)
      notify('Your practice profile has been updated.')
    } catch (err) {
      setPracticeError(errorMessage(err, 'Unable to save your practice profile.'))
    } finally {
      setSavingPractice(false)
    }
  }

  return <main className="mx-auto w-full max-w-5xl px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader eyebrow="Account" title="Settings" description="Manage your profile, password and practice details. Changes are saved to the clinic database." />

    <Tabs defaultValue="profile" className="w-full">
      <TabsList aria-label="Account settings">
        <TabsTrigger value="profile">Profile</TabsTrigger>
        <TabsTrigger value="password">Password</TabsTrigger>
        {variant === 'clinician' && canEditPractice && <TabsTrigger value="practice">Practice</TabsTrigger>}
      </TabsList>

      <TabsContent value="profile" className="mt-5">
        <section className="rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-7">
          <div className="mb-6"><h2 className="text-lg font-semibold">Profile information</h2><p className="mt-1 text-sm text-muted-foreground">These details appear across your appointments and account.</p></div>
          {profileError && <div className="mb-4 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{profileError}</div>}
          <form className="grid gap-5 sm:grid-cols-2" onSubmit={saveProfile}>
            <div className="flex items-center gap-4 border-b border-border pb-5 sm:col-span-2">
              <Avatar name={`${firstName} ${lastName}`.trim() || 'Your profile'} src={photoPreview ?? undefined} size="xl" className="shrink-0" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{`${firstName} ${lastName}`.trim() || 'Your name'}</p>
                <p className="truncate text-sm text-muted-foreground">{user?.email}</p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <label className="inline-flex h-8 cursor-pointer items-center gap-2 rounded-full border border-border px-3 text-xs font-medium transition hover:bg-muted">
                    <ImagePlus size={14} aria-hidden="true" /> Change photo
                    <input
                      type="file"
                      accept="image/*"
                      className="sr-only"
                      aria-label="Change profile photo"
                      onChange={event => {
                        const file = event.target.files?.[0]
                        if (file) setPhotoPreview(URL.createObjectURL(file))
                      }}
                    />
                  </label>
                  <Button type="button" variant="ghost" size="sm" className="h-8 px-2 text-xs text-muted-foreground" onClick={() => setPhotoPreview(null)} disabled={!photoPreview}>Remove</Button>
                </div>
              </div>
            </div>
            <Input label="First name" name="given-name" autoComplete="given-name" value={firstName} onChange={event => setFirstName(event.target.value)} required />
            <Input label="Last name" name="family-name" autoComplete="family-name" value={lastName} onChange={event => setLastName(event.target.value)} required />
            <Input label="Phone number" name="phone" type="tel" autoComplete="tel" value={phone} onChange={event => setPhone(event.target.value)} />
            <Input label="Email address" name="email" type="email" value={user?.email ?? ''} readOnly disabled />
            <div className="flex items-center gap-2 rounded-xl bg-muted/50 px-3 py-2 text-[11px] text-muted-foreground sm:col-span-2"><ShieldCheck size={14} aria-hidden="true" /> Email changes are handled by a clinic administrator.</div>
            <div className="flex justify-end border-t pt-4 sm:col-span-2"><Button type="submit" disabled={savingProfile} className="gap-2">{savingProfile ? <LoaderCircle size={14} className="animate-spin" aria-hidden="true" /> : <Save size={14} aria-hidden="true" />} Save profile</Button></div>
          </form>
        </section>
      </TabsContent>

      <TabsContent value="password" className="mt-5">
        <section className="max-w-2xl rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-7">
          <div className="mb-6"><h2 className="text-lg font-semibold">Change password</h2><p className="mt-1 text-sm text-muted-foreground">Changing your password signs you out of every device.</p></div>
          {passwordError && <div className="mb-4 flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{passwordError}</div>}
          <form className="grid gap-5" onSubmit={savePassword}>
            <PasswordField label="Current password" name="current-password" autoComplete="current-password" value={currentPassword} onChange={event => setCurrentPassword(event.target.value)} required />
            <PasswordStrength label="New password" name="new-password" autoComplete="new-password" minLength={12} value={newPassword} onValueChange={setNewPassword} required />
            <PasswordField label="Confirm new password" name="confirm-password" autoComplete="new-password" minLength={12} value={confirmPassword} onChange={event => setConfirmPassword(event.target.value)} required />
            <div className="flex justify-end border-t pt-4"><Button type="submit" disabled={savingPassword} className="gap-2">{savingPassword ? <LoaderCircle size={14} className="animate-spin" aria-hidden="true" /> : <Save size={14} aria-hidden="true" />} Update password</Button></div>
          </form>
        </section>
      </TabsContent>


      {variant === 'clinician' && canEditPractice && <TabsContent value="practice" className="mt-5">
        <section className="rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-7">
          <div className="mb-6"><h2 className="text-lg font-semibold">Practice profile</h2><p className="mt-1 text-sm text-muted-foreground">Patients see this information while choosing a doctor.</p></div>
          {practiceLoading ? <LoadingState label="Loading your practice profile…" /> : practiceError && !profile ? <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{practiceError}</div> : <form className="grid gap-5 sm:grid-cols-2" onSubmit={savePractice}>
            {practiceError && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive sm:col-span-2"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{practiceError}</div>}
            <Input label="Specialty" name="specialty" value={specialty} onChange={event => setSpecialty(event.target.value)} required />
            <Input label="Languages" name="languages" placeholder="English, Swahili" value={languages} onChange={event => setLanguages(event.target.value)} />
            <Input label="Years of experience" name="experience" type="number" min={0} max={80} value={experienceYears} onChange={event => setExperienceYears(event.target.value)} />
            <Input label="Slot size (minutes)" name="slot_minutes" type="number" min={5} max={180} step={5} value={slotMinutes} onChange={event => setSlotMinutes(event.target.value)} />
            <Input label="Timezone" name="timezone" value={timezone} onChange={event => setTimezone(event.target.value)} />
            <label className="flex items-center justify-between gap-3 self-end rounded-xl border border-border px-3 py-2.5 text-xs font-medium">
              Accepting new patients
              <input type="checkbox" checked={accepting} onChange={event => setAccepting(event.target.checked)} className="size-4 accent-neutral-950" />
            </label>
            <div className="sm:col-span-2"><Textarea label="Professional bio" name="bio" rows={4} value={bio} onChange={event => setBio(event.target.value)} placeholder="A short introduction patients will see" /></div>
            <div className="flex justify-end border-t pt-4 sm:col-span-2"><Button type="submit" disabled={savingPractice} className="gap-2">{savingPractice ? <LoaderCircle size={14} className="animate-spin" aria-hidden="true" /> : <Save size={14} aria-hidden="true" />} Save practice profile</Button></div>
          </form>}
        </section>
      </TabsContent>}
    </Tabs>
  </main>
}
