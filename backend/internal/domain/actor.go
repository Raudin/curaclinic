package domain

import "github.com/google/uuid"

type Actor struct {
	UserID      uuid.UUID
	SessionID   uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Roles       []string
	Permissions map[string]struct{}
}

var PatientPermissionCodes = []string{
	PermAppointmentsViewOwn,
	PermAppointmentsCreateSelf,
	PermAppointmentsCancelOwn,
	PermAppointmentsRescheduleOwn,
	PermDoctorsView,
	PermProfilesUpdateOwn,
}


func IsPatientPermission(code string) bool {
	for _, permission := range PatientPermissionCodes {
		if permission == code {
			return true
		}
	}
	return false
}

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

func (a *Actor) CanAny(codes ...string) bool {
	for _, code := range codes {
		if a.Can(code) {
			return true
		}
	}
	return false
}

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
type RequestMeta struct {
	RequestID string
	IP        string
	UserAgent string
}
