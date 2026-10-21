// Package domain holds the core types, errors and permission codes shared by
// every layer of the API.
package domain

import "github.com/google/uuid"

// Actor is the authenticated caller of a request, with resolved roles and
// permissions.
type Actor struct {
	UserID      uuid.UUID
	SessionID   uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Roles       []string
	Permissions map[string]struct{}
}

// PatientPermissionCodes are granted by the patient identity itself rather
// than through RBAC roles.
var PatientPermissionCodes = []string{
	PermAppointmentsViewOwn,
	PermAppointmentsCreateSelf,
	PermAppointmentsCancelOwn,
	PermAppointmentsRescheduleOwn,
	PermDoctorsView,
	PermProfilesUpdateOwn,
}

// IsPatientPermission reports whether code is a built-in patient permission.
func IsPatientPermission(code string) bool {
	for _, permission := range PatientPermissionCodes {
		if permission == code {
			return true
		}
	}
	return false
}

// Can reports whether the actor holds the permission, directly or through
// their patient identity.
func (a *Actor) Can(code string) bool {
	if a == nil || a.Permissions == nil {
		return false
	}
	if _, ok := a.Permissions[code]; ok {
		return true
	}
	return a.hasPatientIdentity() && IsPatientPermission(code)
}

func (a *Actor) hasPatientIdentity() bool {
	for _, accountType := range a.Roles {
		if accountType == "patient" {
			return true
		}
	}
	return false
}

// CanAny reports whether the actor holds at least one of the permissions.
func (a *Actor) CanAny(codes ...string) bool {
	for _, code := range codes {
		if a.Can(code) {
			return true
		}
	}
	return false
}

// DisplayName returns the actor's full name.
func (a *Actor) DisplayName() string {
	name := a.FirstName
	if a.LastName != "" {
		if name != "" {
			name += " "
		}
		name += a.LastName
	}
	return name
}

// RequestMeta identifies the request that triggered a change, for audit logs.
type RequestMeta struct {
	RequestID string
	IP        string
	UserAgent string
}
