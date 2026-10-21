package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/handlers"
	"curaclinic/backend/internal/middleware"
)

func registerPublicAuthRoutes(api *gin.RouterGroup, h *handlers.Handlers, deps Deps) {
	authLimiter := middleware.RateLimit(deps.Redis, deps.Logger, middleware.RateLimitConfig{
		Scope:  "auth",
		Limit:  deps.Config.RateLimitAuthPer15Min,
		Window: 15 * time.Minute,
	})
	api.POST("/auth/login", authLimiter, h.Login)
	api.POST("/auth/register", authLimiter, h.Register)
	api.POST("/auth/refresh", authLimiter, h.Refresh)
	api.POST("/auth/logout", h.Logout)
}

func registerRoutes(api *gin.RouterGroup, h *handlers.Handlers, deps Deps) {
	secured := api.Group("")
	secured.Use(middleware.Auth(deps.Tokens, deps.UserLoader, deps.RBAC, deps.Logger))

	secured.GET("/auth/me", h.Me)
	secured.POST("/auth/change-password", h.ChangePassword)

	secured.GET("/me/profile", h.MyProfile)
	secured.PATCH("/me/profile", requireAny(deps, domain.PermProfilesUpdateOwn, domain.PermDoctorProfileUpdateOwn, domain.PermAdminUsersUpdate), h.UpdateMyProfile)
	secured.GET("/me/schedule", require(deps, domain.PermScheduleView), h.MySchedule)
	secured.PUT("/me/working-hours", requireAny(deps, domain.PermDoctorWorkingHours, domain.PermAdminWorkingHours), h.ReplaceMyWorkingHours)
	secured.GET("/me/schedule-blocks", requireAny(deps, domain.PermDoctorScheduleView, domain.PermAdminWorkingHours), h.MyBlocks)
	secured.POST("/me/schedule-blocks", requireAny(deps, domain.PermDoctorBlockTime, domain.PermAdminWorkingHours), h.CreateMyBlock)
	secured.DELETE("/me/schedule-blocks/:id", requireAny(deps, domain.PermDoctorRemoveBlock, domain.PermAdminWorkingHours), h.DeleteMyBlock)

	secured.GET("/doctors", requireAny(deps, domain.PermDoctorsView, domain.PermAdminUsersView, domain.PermDoctorPatientsView), h.ListDoctors)
	secured.GET("/doctors/:id", requireAny(deps, domain.PermDoctorsView, domain.PermAdminUsersView, domain.PermDoctorPatientsView), h.GetDoctor)
	secured.POST("/doctors", require(deps, domain.PermAdminUsersCreate), h.CreateDoctor)
	secured.PUT("/doctors/:id/profile", requireAny(deps, domain.PermDoctorProfileUpdateOwn, domain.PermAdminUsersUpdate), h.UpdateDoctorProfile)
	secured.GET("/doctors/:id/working-hours", requireAny(deps, domain.PermDoctorScheduleView, domain.PermAdminWorkingHours), h.GetDoctorWorkingHours)
	secured.PUT("/doctors/:id/working-hours", requireAny(deps, domain.PermDoctorWorkingHours, domain.PermAdminWorkingHours), h.ReplaceDoctorWorkingHours)
	secured.GET("/doctors/:id/schedule-blocks", requireAny(deps, domain.PermDoctorScheduleView, domain.PermAdminWorkingHours), h.GetDoctorBlocks)
	secured.POST("/doctors/:id/schedule-blocks", requireAny(deps, domain.PermDoctorBlockTime, domain.PermAdminWorkingHours), h.CreateDoctorBlock)
	secured.DELETE("/schedule-blocks/:id", requireAny(deps, domain.PermDoctorRemoveBlock, domain.PermAdminWorkingHours), h.DeleteBlock)
	secured.GET("/doctors/:id/availability", requireAny(deps, domain.PermAppointmentsCreateSelf, domain.PermDoctorAppointmentsCreate, domain.PermAdminAppointmentsCreate, domain.PermDoctorScheduleView), h.GetAvailability)

	secured.POST("/holds", require(deps, domain.PermAppointmentsCreateSelf), h.CreateHold)
	secured.DELETE("/holds/:token", h.ReleaseHold)

	secured.GET("/appointments", requireAny(deps, domain.PermAppointmentsViewOwn, domain.PermDoctorAppointmentsViewOwn, domain.PermAdminAppointmentsView), h.ListAppointments)
	secured.GET("/appointments/:id", requireAny(deps, domain.PermAppointmentsViewOwn, domain.PermDoctorAppointmentsViewOwn, domain.PermAdminAppointmentsView), h.GetAppointment)
	secured.GET("/appointments/:id/events", requireAny(deps, domain.PermAppointmentsViewOwn, domain.PermDoctorAppointmentsViewOwn, domain.PermAdminAppointmentsView), h.AppointmentEvents)
	secured.POST("/appointments", require(deps, domain.PermAppointmentsCreateSelf), h.BookAppointment)
	secured.POST("/appointments/manual", requireAny(deps, domain.PermDoctorAppointmentsCreate, domain.PermAdminAppointmentsCreate), h.BookManualAppointment)
	secured.PATCH("/appointments/:id/cancel", requireAny(deps, domain.PermAppointmentsCancelOwn, domain.PermDoctorAppointmentsCancel, domain.PermAdminAppointmentsCancel), h.CancelAppointment)
	secured.PATCH("/appointments/:id/reschedule", requireAny(deps, domain.PermAppointmentsRescheduleOwn, domain.PermDoctorAppointmentsMove, domain.PermAdminAppointmentsMove), h.RescheduleAppointment)

	secured.GET("/users", require(deps, domain.PermAdminUsersView), h.ListUsers)
	secured.GET("/users/:id", require(deps, domain.PermAdminUsersView), h.GetUser)
	secured.POST("/users", require(deps, domain.PermAdminUsersCreate), h.CreateUser)
	secured.PATCH("/users/:id/profile", require(deps, domain.PermAdminUsersUpdate), h.UpdateUserProfile)
	secured.PATCH("/users/:id/status", require(deps, domain.PermAdminUsersUpdate), h.UpdateUserStatus)
	secured.PUT("/users/:id/roles", require(deps, domain.PermAdminUsersAssignRoles), h.AssignUserRoles)
	secured.GET("/users/:id/permissions", require(deps, domain.PermAdminUsersView), h.GetUserAccess)
	secured.PUT("/users/:id/permissions", require(deps, domain.PermAdminPermissionsManage), h.SetUserOverride)
	secured.DELETE("/users/:id/permissions/:permissionId", require(deps, domain.PermAdminPermissionsManage), h.DeleteUserOverride)

	secured.GET("/roles", require(deps, domain.PermAdminRolesManage), h.ListRoles)
	secured.POST("/roles", require(deps, domain.PermAdminRolesManage), h.CreateRole)
	secured.PUT("/roles/:id", require(deps, domain.PermAdminRolesManage), h.UpdateRole)
	secured.DELETE("/roles/:id", require(deps, domain.PermAdminRolesManage), h.DeleteRole)
	secured.GET("/permissions", require(deps, domain.PermAdminPermissionsManage), h.ListPermissions)

	secured.GET("/audit-logs", require(deps, domain.PermAdminAuditView), h.ListAuditLogs)
}
