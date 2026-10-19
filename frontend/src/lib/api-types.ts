export type RoleName = 'admin' | 'doctor' | 'patient'

export type RoleScope = 'doctor' | 'admin'

export type User = {
  id: string
  email: string
  first_name: string
  last_name: string
  phone?: string
  status: 'active' | 'inactive'
  roles: string[]
  created_at: string
}

export type TokenResponse = {
  access_token: string
  refresh_token: string
  token_type: 'Bearer'
  expires_in: number
  user: User
  roles: string[]
  permissions: string[]
}

export type Account = {
  user: User
  roles: string[]
  permissions: string[]
}

export type Doctor = {
  id: string
  profile_id: string
  first_name: string
  last_name: string
  email: string
  specialty: string
  bio?: string
  experience_years: number
  languages: string[]
  avatar_url?: string
  timezone: string
  slot_minutes: number
  is_accepting_patients: boolean
}

export type WorkingHour = {
  id: string
  weekday: number
  start_time: string
  end_time: string
}

export type ScheduleBlock = {
  id: string
  doctor_id: string
  start_time: string
  end_time: string
  reason: string
  created_at: string
}

export type Schedule = {
  working_hours: WorkingHour[]
  blocks: ScheduleBlock[]
}

export type Slot = {
  start_time: string
  end_time: string
}

export type Availability = {
  doctor_id: string
  date: string
  timezone: string
  slot_minutes: number
  slots: Slot[]
}

export type Hold = {
  token: string
  doctor_id: string
  start_time: string
  expires_at: string
}

export type AppointmentStatus = 'confirmed' | 'cancelled'

export type Appointment = {
  id: string
  doctor: {
    id: string
    first_name: string
    last_name: string
    specialty: string
  }
  patient: {
    id: string
    first_name: string
    last_name: string
    email: string
  }
  start_time: string
  end_time: string
  status: AppointmentStatus
  reason: string
  cancelled_at?: string
  created_at: string
  updated_at: string
}

export type AppointmentEvent = {
  id: number
  event_type: 'created' | 'cancelled' | 'rescheduled'
  from_start_time?: string
  to_start_time?: string
  reason?: string
  actor_name: string
  created_at: string
}

export type Pagination = {
  limit: number
  offset: number
  total: number
}

export type Page<T> = {
  data: T[]
  pagination: Pagination
}

export type Role = {
  id: string
  scope: RoleScope
  name: string
  description?: string
  is_system: boolean
  permissions: string[]
  created_at: string
}

export type Permission = {
  id: string
  scope: RoleScope
  resource: string
  action: string
  code: string
  description?: string
}

export type PermissionOverride = {
  id: string
  permission_id: string
  permission: string
  effect: 'allow' | 'deny'
  note?: string
  created_at: string
  updated_at: string
}

export type UserAccess = {
  roles: string[]
  permissions: string[]
  overrides: PermissionOverride[]
}

export type AuditLog = {
  id: number
  actor_id?: string
  action: string
  resource_type: string
  resource_id?: string
  metadata: Record<string, unknown>
  request_id?: string
  created_at: string
}

export type ApiErrorBody = {
  error?: {
    code?: string
    message?: string
    details?: Record<string, string>
    request_id?: string
  }
}
