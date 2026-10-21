package domain

const (
	AuditUserCreated            = "user.created"
	AuditUserStatusUpdated      = "user.status_updated"
	AuditProfileUpdated         = "user.profile_updated"
	AuditRolesAssigned          = "rbac.roles_assigned"
	AuditOverrideSet            = "rbac.override_set"
	AuditOverrideRemoved        = "rbac.override_removed"
	AuditRoleCreated            = "rbac.role_created"
	AuditRoleUpdated            = "rbac.role_updated"
	AuditRoleDeleted            = "rbac.role_deleted"
	AuditDoctorProfileCreated   = "doctor.profile_created"
	AuditDoctorProfileUpdated   = "doctor.profile_updated"
	AuditWorkingHoursUpdated    = "schedule.working_hours_updated"
	AuditBlockCreated           = "schedule.block_created"
	AuditBlockDeleted           = "schedule.block_deleted"
	AuditAppointmentCreated     = "appointment.created"
	AuditAppointmentCancelled   = "appointment.cancelled"
	AuditAppointmentRescheduled = "appointment.rescheduled"
)

const (
	ResourceUser        = "user"
	ResourceRole        = "role"
	ResourceAppointment = "appointment"
	ResourceSchedule    = "schedule"
	ResourceDoctor      = "doctor"
)
