package domain

// Patient permissions are built into the patient identity and are not stored
// in the RBAC tables. Doctor and admin codes are seeded and assigned through
// scoped roles and direct grants.
const (
	PermAppointmentsViewOwn       = "appointments.view_own"
	PermAppointmentsCreateSelf    = "appointments.create_self"
	PermAppointmentsCancelOwn     = "appointments.cancel_own"
	PermAppointmentsRescheduleOwn = "appointments.reschedule_own"
	PermDoctorsView               = "doctors.view"
	PermProfilesUpdateOwn         = "profiles.update_own"

	PermDoctorProfileUpdateOwn    = "doctor.profile.update_own"
	PermDoctorAppointmentsViewOwn = "doctor.appointments.view_own"
	PermDoctorAppointmentsCreate  = "doctor.appointments.create_for_patient"
	PermDoctorAppointmentsCancel  = "doctor.appointments.cancel_own"
	PermDoctorAppointmentsMove    = "doctor.appointments.reschedule_own"
	PermDoctorPatientsView        = "doctor.patients.view"
	PermDoctorScheduleView        = "doctor.schedule.view_own"
	PermDoctorWorkingHours        = "doctor.schedule.define_working_hours_own"
	PermDoctorBlockTime           = "doctor.schedule.block_time_own"
	PermDoctorRemoveBlock         = "doctor.schedule.remove_block_own"

	PermAdminUsersCreate        = "admin.users.create"
	PermAdminUsersView          = "admin.users.view"
	PermAdminUsersUpdate        = "admin.users.update_profiles"
	PermAdminUsersAssignRoles   = "admin.users.assign_roles"
	PermAdminRolesManage        = "admin.roles.manage"
	PermAdminPermissionsManage  = "admin.permissions.manage"
	PermAdminWorkingHours       = "admin.schedule.define_working_hours_any"
	PermAdminAuditView          = "admin.audit.view"
	PermAdminAppointmentsView   = "admin.appointments.view_any"
	PermAdminAppointmentsCreate = "admin.appointments.create_any"
	PermAdminAppointmentsCancel = "admin.appointments.cancel_any"
	PermAdminAppointmentsMove   = "admin.appointments.reschedule_any"

	// Legacy names remain aliases for service code that distinguishes patient
	// and clinic-wide appointment actions.
	PermAppointmentsViewAny        = PermAdminAppointmentsView
	PermAppointmentsCreateManual   = PermDoctorAppointmentsCreate
	PermAppointmentsCancelAny      = PermAdminAppointmentsCancel
	PermAppointmentsRescheduleAny  = PermAdminAppointmentsMove
	PermProfilesUpdateAny          = PermAdminUsersUpdate
	PermUsersView                  = PermAdminUsersView
	PermUsersManage                = PermAdminUsersCreate
	PermUsersManageRoles           = PermAdminUsersAssignRoles
	PermPermissionsAssign          = PermAdminPermissionsManage
	PermReportsView                = PermAdminAuditView
	PermDoctorsManage              = PermAdminUsersCreate
	PermAuditView                  = PermAdminAuditView
	PermScheduleView               = PermDoctorScheduleView
	PermScheduleDefineWorkingHours = PermDoctorWorkingHours
	PermScheduleUpdateWorkingHours = PermDoctorWorkingHours
	PermScheduleBlockTime          = PermDoctorBlockTime
	PermScheduleDeleteBlock        = PermDoctorRemoveBlock
)

// AllPermissions is the canonical scoped RBAC seed catalog.
var AllPermissions = []struct {
	Code        string
	Resource    string
	Action      string
	Description string
}{
	{PermDoctorProfileUpdateOwn, "doctor_profile", "update_own", "Update own doctor profile"},
	{PermDoctorAppointmentsViewOwn, "appointments", "view_own", "View appointments assigned to this doctor"},
	{PermDoctorAppointmentsCreate, "appointments", "create_for_patient", "Book an appointment for a patient"},
	{PermDoctorAppointmentsCancel, "appointments", "cancel_own", "Cancel appointments assigned to this doctor"},
	{PermDoctorAppointmentsMove, "appointments", "reschedule_own", "Reschedule appointments assigned to this doctor"},
	{PermDoctorPatientsView, "patients", "view", "View patients when managing appointments"},
	{PermDoctorScheduleView, "schedule", "view_own", "View own schedule"},
	{PermDoctorWorkingHours, "schedule", "define_working_hours_own", "Define own recurring working hours"},
	{PermDoctorBlockTime, "schedule", "block_time_own", "Block own time"},
	{PermDoctorRemoveBlock, "schedule", "remove_block_own", "Remove own schedule blocks"},
	{PermAdminUsersCreate, "users", "create", "Create doctor, patient and admin records"},
	{PermAdminUsersView, "users", "view", "View doctor, patient and admin records"},
	{PermAdminUsersUpdate, "users", "update_profiles", "Update user profiles including own"},
	{PermAdminUsersAssignRoles, "users", "assign_roles", "Assign roles to doctor and admin records"},
	{PermAdminRolesManage, "roles", "manage", "Create and update roles"},
	{PermAdminPermissionsManage, "permissions", "manage", "Manage permissions and role permission sets"},
	{PermAdminWorkingHours, "schedule", "define_working_hours_any", "Define recurring working hours for doctors"},
	{PermAdminAuditView, "audit", "view", "View audit records"},
	{PermAdminAppointmentsView, "appointments", "view_any", "View patient and doctor appointments"},
	{PermAdminAppointmentsCreate, "appointments", "create_any", "Book appointments for patients with doctors"},
	{PermAdminAppointmentsCancel, "appointments", "cancel_any", "Cancel patient and doctor appointments"},
	{PermAdminAppointmentsMove, "appointments", "reschedule_any", "Reschedule patient and doctor appointments"},
}
