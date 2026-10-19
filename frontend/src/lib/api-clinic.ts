import { apiList, apiRequest, toQuery } from './api'
import type { Appointment, AppointmentEvent, Availability, Doctor, Hold } from './api-types'

export const doctorsApi = {
  list(params?: { specialty?: string; accepting?: boolean }) {
    return apiRequest<Doctor[]>(`doctors${toQuery(params)}`)
  },
  get(id: string) {
    return apiRequest<Doctor>(`doctors/${id}`)
  },
  create(input: Record<string, unknown>) {
    return apiRequest<Doctor>('doctors', { method: 'POST', body: JSON.stringify(input) })
  },
  updateProfile(id: string, input: {
    specialty: string
    bio?: string
    experience_years: number
    languages: string[]
    avatar_url?: string
    timezone: string
    slot_minutes: number
    is_accepting_patients: boolean
  }) {
    return apiRequest<Doctor>(`doctors/${id}/profile`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  },
}

export const availabilityApi = {
  get(doctorId: string, date: string) {
    return apiRequest<Availability>(`doctors/${doctorId}/availability${toQuery({ date })}`)
  },
}

export const holdsApi = {
  create(doctorId: string, startTime: string) {
    return apiRequest<Hold>('holds', {
      method: 'POST',
      body: JSON.stringify({ doctor_id: doctorId, start_time: startTime }),
    })
  },
  release(token: string) {
    return apiRequest<void>(`holds/${encodeURIComponent(token)}`, { method: 'DELETE' })
  },
}

export const appointmentsApi = {
  list(params?: {
    scope?: 'upcoming' | 'past' | 'cancelled' | 'all'
    doctorId?: string
    patientId?: string
    from?: string
    to?: string
    order?: 'asc' | 'desc'
    limit?: number
    offset?: number
  }) {
    return apiList<Appointment>(`appointments${toQuery({
      scope: params?.scope,
      doctor_id: params?.doctorId,
      patient_id: params?.patientId,
      from: params?.from,
      to: params?.to,
      order: params?.order,
      limit: params?.limit,
      offset: params?.offset,
    })}`)
  },
  get(id: string) {
    return apiRequest<Appointment>(`appointments/${id}`)
  },
  book(input: { doctor_id: string; start_time: string; reason: string; hold_token: string }) {
    return apiRequest<Appointment>('appointments', { method: 'POST', body: JSON.stringify(input) })
  },
  manualBook(input: { doctor_id: string; patient_id: string; start_time: string; reason: string }) {
    return apiRequest<Appointment>('appointments/manual', { method: 'POST', body: JSON.stringify(input) })
  },
  cancel(id: string, reason: string) {
    return apiRequest<Appointment>(`appointments/${id}/cancel`, {
      method: 'PATCH',
      body: JSON.stringify({ reason }),
    })
  },
  reschedule(id: string, startTime: string, reason?: string) {
    return apiRequest<Appointment>(`appointments/${id}/reschedule`, {
      method: 'PATCH',
      body: JSON.stringify({ start_time: startTime, reason }),
    })
  },
  events(id: string) {
    return apiRequest<AppointmentEvent[]>(`appointments/${id}/events`)
  },
}
