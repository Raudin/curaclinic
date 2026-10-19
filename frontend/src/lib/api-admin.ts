import { apiList, apiRequest, requestRaw, toQuery } from './api'
import type { AuditLog, Permission, Role, RoleScope, Schedule, ScheduleBlock, User, UserAccess } from './api-types'

export const scheduleApi = {
  mine() {
    return apiRequest<Schedule>('me/schedule')
  },
  replaceMine(intervals: Array<{ weekday: number; start_time: string; end_time: string }>) {
    return apiRequest<{ intervals: Schedule['working_hours'] }>('me/working-hours', {
      method: 'PUT',
      body: JSON.stringify({ intervals }),
    })
  },
  createMyBlock(input: { start_time: string; end_time: string; reason: string }) {
    return apiRequest<ScheduleBlock>('me/schedule-blocks', { method: 'POST', body: JSON.stringify(input) })
  },
  deleteMyBlock(id: string) {
    return apiRequest<void>(`me/schedule-blocks/${id}`, { method: 'DELETE' })
  },
  workingHours(doctorId: string) {
    return apiRequest<{ intervals: Schedule['working_hours'] }>(`doctors/${doctorId}/working-hours`)
  },
  replaceWorkingHours(doctorId: string, intervals: Array<{ weekday: number; start_time: string; end_time: string }>) {
    return apiRequest<{ intervals: Schedule['working_hours'] }>(`doctors/${doctorId}/working-hours`, {
      method: 'PUT',
      body: JSON.stringify({ intervals }),
    })
  },
  blocks(doctorId: string, params?: { from?: string; to?: string }) {
    return apiRequest<ScheduleBlock[]>(`doctors/${doctorId}/schedule-blocks${toQuery(params)}`)
  },
  createBlock(doctorId: string, input: { start_time: string; end_time: string; reason: string }) {
    return apiRequest<ScheduleBlock>(`doctors/${doctorId}/schedule-blocks`, {
      method: 'POST',
      body: JSON.stringify(input),
    })
  },
  deleteBlock(id: string) {
    return apiRequest<void>(`schedule-blocks/${id}`, { method: 'DELETE' })
  },
}

export const usersApi = {
  list(params?: { search?: string; role?: string; limit?: number; offset?: number }) {
    return apiList<User>(`users${toQuery(params)}`)
  },
  get(id: string) {
    return apiRequest<User>(`users/${id}`)
  },
  create(input: {
    email: string
    password: string
    first_name: string
    last_name: string
    phone?: string
    status?: 'active' | 'inactive'
    role_ids: string[]
  }) {
    return apiRequest<User>('users', { method: 'POST', body: JSON.stringify(input) })
  },
  setStatus(id: string, status: 'active' | 'inactive') {
    return apiRequest<User>(`users/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    })
  },
  assignRoles(id: string, roleIds: string[]) {
    return apiRequest<User>(`users/${id}/roles`, {
      method: 'PUT',
      body: JSON.stringify({ role_ids: roleIds }),
    })
  },
  access(id: string) {
    return apiRequest<UserAccess>(`users/${id}/permissions`)
  },
  setOverride(id: string, input: { permission_id: string; effect: 'allow' | 'deny'; note?: string }) {
    return apiRequest<{ id: string }>(`users/${id}/permissions`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  },
  removeOverride(id: string, permissionId: string) {
    return apiRequest<void>(`users/${id}/permissions/${permissionId}`, { method: 'DELETE' })
  },
}

export const rolesApi = {
  list() {
    return apiRequest<Role[]>('roles')
  },
  create(input: { scope: RoleScope; name: string; description?: string; permission_ids: string[] }) {
    return apiRequest<Role>('roles', { method: 'POST', body: JSON.stringify(input) })
  },
  update(id: string, input: { name: string; description?: string; permission_ids: string[] }) {
    return apiRequest<Role>(`roles/${id}`, { method: 'PUT', body: JSON.stringify(input) })
  },
  remove(id: string) {
    return apiRequest<void>(`roles/${id}`, { method: 'DELETE' })
  },
}

export const permissionsApi = {
  list() {
    return apiRequest<Permission[]>('permissions')
  },
}

export const auditApi = {
  list(params?: { resource_type?: string; resource_id?: string; actor_id?: string; limit?: number; offset?: number }) {
    return apiList<AuditLog>(`audit-logs${toQuery(params)}`)
  },
}

// Re-export the raw request helper for cases that need full envelopes.
export { requestRaw }
