import { useState, type FormEvent } from 'react'
import { AlertCircle, ArrowRight, LoaderCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/arc/input/input'
import { PasswordField } from '@/components/arc/password-field/password-field'
import { PasswordStrength } from '@/components/arc/password-strength/password-strength'
import { useAuth, homeForRoles } from '@/lib/auth-context'
import { ApiError } from '@/lib/api'
import { AuthLayout } from '@/components/auth/auth-layout'

export function SignupPage({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { register } = useAuth()
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [fieldError, setFieldError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setFieldError(null)

    if (!firstName.trim() || !lastName.trim() || !email.trim()) {
      setError('Please fill in your name and email address.')
      return
    }
    if (password.length < 12) {
      setFieldError('Your password must be at least 12 characters.')
      return
    }
    if (password !== confirmPassword) {
      setFieldError('The passwords do not match.')
      return
    }

    setSubmitting(true)
    try {
      const result = await register({
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        email: email.trim(),
        phone: phone.trim() || undefined,
        password,
      })
      onNavigate(homeForRoles(result.roles, result.permissions))
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
        if (err.details?.email) setFieldError(err.details.email)
      } else {
        setError('Unable to create your account. Please try again.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return <AuthLayout
    title="Create your patient account"
    subtitle="Book visits, manage appointments and keep your care details in one place."
    footer={<>Already registered? <button type="button" onClick={() => onNavigate('/login')} className="font-semibold text-foreground underline underline-offset-4">Sign in</button></>}
  >
    <form onSubmit={onSubmit} className="grid gap-4" noValidate>
      {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
      <div className="grid gap-4 sm:grid-cols-2">
        <Input label="First name" name="first_name" autoComplete="given-name" value={firstName} onChange={event => setFirstName(event.target.value)} required />
        <Input label="Last name" name="last_name" autoComplete="family-name" value={lastName} onChange={event => setLastName(event.target.value)} required />
      </div>
      <Input label="Email address" name="email" type="email" autoComplete="email" placeholder="you@example.com" value={email} onChange={event => setEmail(event.target.value)} required />
      <Input label="Phone number (optional)" name="phone" type="tel" autoComplete="tel" placeholder="+254 700 000 000" value={phone} onChange={event => setPhone(event.target.value)} />
      <PasswordStrength
        label="Password"
        name="password"
        autoComplete="new-password"
        value={password}
        onValueChange={value => { setPassword(value); setFieldError(null) }}
        error={fieldError ?? undefined}
        required
      />
      <PasswordField
        label="Confirm password"
        name="confirm_password"
        autoComplete="new-password"
        value={confirmPassword}
        onChange={event => setConfirmPassword(event.target.value)}
        required
      />
      <p className="text-[11px] leading-relaxed text-muted-foreground">Use at least 12 characters. Patient accounts can book and manage only their own appointments.</p>
      <Button type="submit" disabled={submitting} className="mt-1 h-10 w-full gap-2 rounded-xl">
        {submitting ? <LoaderCircle size={15} className="animate-spin" aria-hidden="true" /> : null}
        {submitting ? 'Creating account…' : 'Create account'}
        {!submitting && <ArrowRight size={15} aria-hidden="true" />}
      </Button>
    </form>
  </AuthLayout>
}
