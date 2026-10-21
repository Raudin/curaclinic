import { DailyAgendaPage } from '@/components/clinician/daily-agenda-page'
import { ScheduleAvailabilityPage } from '@/components/clinician/schedule-availability-page'
import { WeekCalendarPage } from '@/components/clinician/week-calendar-page'

export function ClinicianPages({ pathname, onNavigate }: { pathname: string; onNavigate: (path: string) => void }) {
  if (pathname === '/schedule-availability') return <ScheduleAvailabilityPage />
  if (pathname === '/doctor-appointments') return <WeekCalendarPage onNavigate={onNavigate} />
  return <DailyAgendaPage onNavigate={onNavigate} />
}
