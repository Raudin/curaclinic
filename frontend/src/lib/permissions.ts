import type { RoleScope } from './api-types'

// Keep these codes in sync with backend/internal/domain/permissions.go.
// Patient permissions are implicit on the server (domain.Actor.Can) and are
// therefore not returned by GET /permissions or stored in the RBAC tables.
export const PERMISSIONS = {
  appointmentsViewOwn: 'appointments.view_own',
  appointmentsCreateSelf: 'appointments.create_self',
  appointmentsCancelOwn: 'appointments.cancel_own',
  appointmentsRescheduleOwn: 'appointments.reschedule_own',
  doctorsView: 'doctors.view',
  profilesUpdateOwn: 'profiles.update_own',

  doctorProfileUpdateOwn: 'doctor.profile.update_own',
  doctorAppointmentsViewOwn: 'doctor.appointments.view_own',
  doctorAppointmentsCreate: 'doctor.appointments.create_for_patient',
  doctorAppointmentsCancel: 'doctor.appointments.cancel_own',
  doctorAppointmentsReschedule: 'doctor.appointments.reschedule_own',
  doctorPatientsView: 'doctor.patients.view',
  doctorScheduleView: 'doctor.schedule.view_own',
  doctorWorkingHours: 'doctor.schedule.define_working_hours_own',
  doctorBlockTime: 'doctor.schedule.block_time_own',
  doctorRemoveBlock: 'doctor.schedule.remove_block_own',

  adminUsersCreate: 'admin.users.create',
  adminUsersView: 'admin.users.view',
  adminUsersUpdate: 'admin.users.update_profiles',
  adminUsersAssignRoles: 'admin.users.assign_roles',
  adminRolesManage: 'admin.roles.manage',
  adminPermissionsManage: 'admin.permissions.manage',
  adminWorkingHours: 'admin.schedule.define_working_hours_any',
  adminAuditView: 'admin.audit.view',
  adminAppointmentsView: 'admin.appointments.view_any',
  adminAppointmentsCreate: 'admin.appointments.create_any',
  adminAppointmentsCancel: 'admin.appointments.cancel_any',
  adminAppointmentsReschedule: 'admin.appointments.reschedule_any',
} as const

export type PermissionCode = (typeof PERMISSIONS)[keyof typeof PERMISSIONS]

export const PATIENT_PERMISSIONS: readonly PermissionCode[] = [
  PERMISSIONS.appointmentsViewOwn,
  PERMISSIONS.appointmentsCreateSelf,
  PERMISSIONS.appointmentsCancelOwn,
  PERMISSIONS.appointmentsRescheduleOwn,
  PERMISSIONS.doctorsView,
  PERMISSIONS.profilesUpdateOwn,
]

// Mirrors the server-side patient identity grants so client-side checks agree
// with domain.Actor.Can.
export function resolvePermissions(roles: readonly string[], permissions: readonly string[]): Set<string> {
  const granted = new Set(permissions)
  if (roles.includes('patient')) {
    for (const code of PATIENT_PERMISSIONS) granted.add(code)
  }
  return granted
}

export function hasScopePermission(granted: ReadonlySet<string>, scope: RoleScope): boolean {
  for (const code of granted) {
    if (code.startsWith(`${scope}.`)) return true
  }
  return false
}

export function permissionScope(code: string): RoleScope | null {
  if (code.startsWith('doctor.')) return 'doctor'
  if (code.startsWith('admin.')) return 'admin'
  return null
}

export function friendlyRole(role: string): string {
  switch (role) {
    case 'admin': return 'Clinic administrator'
    case 'doctor': return 'Clinician'
    case 'patient': return 'Patient account'
    default: return role
  }
}
