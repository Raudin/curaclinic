import type { Appointment, Slot } from './api-types'
import { formatDate, formatTimeRange } from './time'

export function canChangeAppointment(appointment: Appointment, now = Date.now()) {
  return appointment.status === 'confirmed' && Date.parse(appointment.start_time) > now
}

export function describeSlot(slot: Slot, zone: string) {
  return [formatDate(slot.start_time, zone), formatTimeRange(slot.start_time, slot.end_time, zone), zone].join(' · ')
}

export function describeAppointment(appointment: Appointment, zone: string) {
  const { patient, doctor } = appointment
  return [patient.first_name + ' ' + patient.last_name, 'Dr. ' + doctor.first_name + ' ' + doctor.last_name, describeSlot(appointment, zone)].join(' · ')
}
