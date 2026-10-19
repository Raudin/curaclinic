import { DailyAgendaPage } from '@/components/clinician/daily-agenda-page'
import { ScheduleAvailabilityPage } from '@/components/clinician/schedule-availability-page'

export function ClinicianPages({ pathname, onNavigate }: { pathname: string; onNavigate: (path: string) => void }) {
  if (pathname === '/schedule-availability') return <ScheduleAvailabilityPage />
  return <DailyAgendaPage onNavigate={onNavigate} />
}
