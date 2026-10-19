import type { ReactNode } from 'react'
import { HeartPulse, ShieldCheck, Sparkles } from 'lucide-react'
import hero from '@/assets/hero.png'

export function AuthLayout({ title, subtitle, children, footer }: {
  title: string
  subtitle: string
  children: ReactNode
  footer?: ReactNode
}) {
  return <div className="min-h-svh bg-[#f7f8f7] lg:grid lg:grid-cols-[1.05fr_0.95fr]">
    <aside className="relative hidden overflow-hidden bg-neutral-950 p-10 text-white lg:flex lg:flex-col lg:justify-between xl:p-14">
      <img src={hero} alt="" className="absolute inset-0 h-full w-full object-cover opacity-20" aria-hidden="true" />
      <div className="absolute inset-0 bg-gradient-to-br from-neutral-950/95 via-neutral-950/80 to-neutral-900/70" aria-hidden="true" />
      <div className="relative flex items-center gap-3">
        <span className="flex size-9 items-center justify-center rounded-xl bg-white text-neutral-950"><HeartPulse size={18} aria-hidden="true" /></span>
        <span className="text-base font-semibold tracking-tight">CuraClinic</span>
      </div>
      <div className="relative max-w-lg">
        <span className="inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-white/10 px-3 py-1 text-[10px] font-medium text-white/70">
          <Sparkles size={12} aria-hidden="true" /> Care, connected
        </span>
        <h2 className="mt-5 text-4xl font-semibold leading-[1.08] tracking-[-0.055em] xl:text-5xl">Your clinic day, in one clear view.</h2>
        <p className="mt-4 max-w-md text-sm leading-relaxed text-white/60">Book visits, manage your schedule and keep every appointment in sync with your care team.</p>
        <ul className="mt-8 space-y-3 text-sm text-white/70">
          <li className="flex items-center gap-2"><ShieldCheck size={15} className="text-white" aria-hidden="true" /> Role-based access for patients, doctors and clinic staff</li>
          <li className="flex items-center gap-2"><ShieldCheck size={15} className="text-white" aria-hidden="true" /> Live availability with protected booking holds</li>
          <li className="flex items-center gap-2"><ShieldCheck size={15} className="text-white" aria-hidden="true" /> Every change recorded in the clinic audit trail</li>
        </ul>
      </div>
      <p className="relative text-[10px] text-white/35">© 2026 CuraClinic · For urgent medical help, contact your local emergency services.</p>
    </aside>

    <main className="flex min-h-svh items-center justify-center px-4 py-10 sm:px-8">
      <div className="w-full max-w-md">
        <div className="mb-7 flex items-center gap-2 lg:hidden">
          <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"><HeartPulse size={16} aria-hidden="true" /></span>
          <span className="text-sm font-semibold">CuraClinic</span>
        </div>
        <div className="rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
          <h1 className="text-2xl font-semibold tracking-[-0.04em] text-foreground sm:text-[28px]">{title}</h1>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{subtitle}</p>
          <div className="mt-6">{children}</div>
        </div>
        {footer && <div className="mt-5 text-center text-xs text-muted-foreground">{footer}</div>}
      </div>
    </main>
  </div>
}
