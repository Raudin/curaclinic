export const DEFAULT_TIMEZONE = 'Africa/Nairobi'

type ZonedParts = {
  year: number
  month: number
  day: number
  hour: number
  minute: number
  second: number
}

function zonedParts(date: Date, timeZone: string): ZonedParts {
  const formatter = new Intl.DateTimeFormat('en-US', {
    timeZone,
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
  const parts: Record<string, string> = {}
  for (const part of formatter.formatToParts(date)) {
    if (part.type !== 'literal') parts[part.type] = part.value
  }
  return {
    year: Number(parts.year),
    month: Number(parts.month),
    day: Number(parts.day),
    hour: Number(parts.hour === '24' ? '0' : parts.hour),
    minute: Number(parts.minute),
    second: Number(parts.second),
  }
}

function zonedOffsetMs(date: Date, timeZone: string): number {
  const parts = zonedParts(date, timeZone)
  const asUTC = Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, parts.second)
  return asUTC - Math.floor(date.getTime() / 1000) * 1000
}

export function dateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function dateFromKey(key: string): Date {
  const [year, month, day] = key.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function dateKeyInZone(date: Date, timeZone: string): string {
  const parts = zonedParts(date, timeZone)
  return `${parts.year}-${String(parts.month).padStart(2, '0')}-${String(parts.day).padStart(2, '0')}`
}

export function addDays(date: Date, days: number): Date {
  const next = new Date(date)
  next.setDate(next.getDate() + days)
  return next
}

export function zonedDayBounds(date: Date, timeZone: string): { from: string; to: string } {
  const parts = zonedParts(date, timeZone)
  const startGuess = Date.UTC(parts.year, parts.month - 1, parts.day, 0, 0, 0)
  const endGuess = Date.UTC(parts.year, parts.month - 1, parts.day + 1, 0, 0, 0)
  const from = new Date(startGuess - zonedOffsetMs(new Date(startGuess), timeZone))
  const to = new Date(endGuess - zonedOffsetMs(new Date(endGuess), timeZone))
  return { from: from.toISOString(), to: to.toISOString() }
}

export function zonedDayBoundsFromKey(key: string, timeZone: string): { from: string; to: string } {
  const [year, month, day] = key.split('-').map(Number)
  const startGuess = Date.UTC(year, month - 1, day, 0, 0, 0)
  const endGuess = Date.UTC(year, month - 1, day + 1, 0, 0, 0)
  const from = new Date(startGuess - zonedOffsetMs(new Date(startGuess), timeZone))
  const to = new Date(endGuess - zonedOffsetMs(new Date(endGuess), timeZone))
  return { from: from.toISOString(), to: to.toISOString() }
}

export function formatTime(iso: string, timeZone: string): string {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(iso))
}

export function formatDate(iso: string, timeZone: string, options: Intl.DateTimeFormatOptions = {}): string {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone,
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    ...options,
  }).format(new Date(iso))
}

export function formatShortDate(iso: string, timeZone: string): string {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone,
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  }).format(new Date(iso))
}

export function formatTimeRange(startIso: string, endIso: string, timeZone: string): string {
  return `${formatTime(startIso, timeZone)} – ${formatTime(endIso, timeZone)}`
}

export function formatWeekday(date: Date): string {
  return new Intl.DateTimeFormat('en-GB', { weekday: 'long' }).format(date)
}

export function minutesToClock(minutes: number): string {
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`
}

export function clockToMinutes(value: string): number {
  const [hours, minutes] = value.split(':').map(Number)
  return hours * 60 + (minutes || 0)
}

export function weekdayName(weekday: number): string {
  return ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'][weekday] ?? ''
}
