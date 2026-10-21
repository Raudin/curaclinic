package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type NewSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	FamilyID         uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	IP               string
	UserAgent        string
}

type AuditEntry struct {
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	Meta         RequestMeta
}

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

type AuditFilter struct {
	ResourceType *string
	ResourceID   *string
	ActorID      *uuid.UUID
	Limit        int
	Offset       int
}
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

type DoctorFilter struct {
	Specialty *string
	Accepting *bool
}
