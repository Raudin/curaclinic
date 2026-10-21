package domain

import (
	"time"

	"github.com/google/uuid"
)

// NewUser is the data needed to create a user account.
type NewUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	DateOfBirth  string
	AvatarURL    string
	Status       string
}

// NewSession is the data needed to store a refresh-token session.
type NewSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	FamilyID         uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	IP               string
	UserAgent        string
}

// AuditEntry is one audit log record to write.
type AuditEntry struct {
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	Meta         RequestMeta
}

// AppointmentFilter narrows an appointment listing query.
type AppointmentFilter struct {
	PatientID *uuid.UUID
	DoctorID  *uuid.UUID
	Status    *string
	FromTime  *time.Time
	ToTime    *time.Time
	EndsAfter *time.Time
	Ascending bool
	Limit     int
	Offset    int
}

// AuditFilter narrows an audit log listing query.
type AuditFilter struct {
	ResourceType *string
	ResourceID   *string
	ActorID      *uuid.UUID
	Limit        int
	Offset       int
}

// NewDoctorProfile is the data needed to create a doctor profile.
type NewDoctorProfile struct {
	UserID              uuid.UUID
	Specialty           string
	Bio                 string
	ExperienceYears     int
	Languages           []string
	Timezone            string
	SlotMinutes         int
	IsAcceptingPatients bool
}

// DoctorFilter narrows a doctor listing query.
type DoctorFilter struct {
	Specialty *string
	Accepting *bool
}
