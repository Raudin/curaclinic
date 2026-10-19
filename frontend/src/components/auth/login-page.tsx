import { useState, type FormEvent } from 'react'
import { AlertCircle, ArrowRight, LoaderCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/arc/input/input'
import { PasswordField } from '@/components/arc/password-field/password-field'
import { useAuth, homeForRoles } from '@/lib/auth-context'
import { ApiError } from '@/lib/api'
import { AuthLayout } from '@/components/auth/auth-layout'

export function LoginPage({ onNavigate, notice }: { onNavigate: (path: string) => void; notice?: string | null }) {
  const { login } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const result = await login(email.trim(), password)
      const next = sessionStorage.getItem('curaclinic.redirect')
      sessionStorage.removeItem('curaclinic.redirect')
      onNavigate(next || homeForRoles(result.roles, result.permissions))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to sign in. Please try again.')
    } finally {
      setSubmitting(false)
    }
  }

  return <AuthLayout
    title="Welcome back"
    subtitle="Sign in to manage appointments, schedules and clinic access."
    footer={<>New patient? <button type="button" onClick={() => onNavigate('/signup')} className="font-semibold text-foreground underline underline-offset-4">Create an account</button></>}
  >
    <form onSubmit={onSubmit} className="grid gap-4" noValidate>
      {notice && <div className="rounded-xl border border-border bg-muted/70 px-3 py-2 text-xs text-muted-foreground">{notice}</div>}
      {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="mt-0.5 shrink-0" aria-hidden="true" />{error}</div>}
      <Input
        label="Email address"
        name="email"
        type="email"
        autoComplete="email"
        placeholder="you@example.com"
        value={email}
        onChange={event => setEmail(event.target.value)}
        required
      />
      <PasswordField
        label="Password"
        name="password"
        autoComplete="current-password"
        placeholder="Your password"
        value={password}
        onChange={event => setPassword(event.target.value)}
        required
      />
      <Button type="submit" disabled={submitting} className="mt-1 h-10 w-full gap-2 rounded-xl">
        {submitting ? <LoaderCircle size={15} className="animate-spin" aria-hidden="true" /> : null}
        {submitting ? 'Signing in…' : 'Sign in'}
        {!submitting && <ArrowRight size={15} aria-hidden="true" />}
      </Button>
    </form>
  </AuthLayout>
}
