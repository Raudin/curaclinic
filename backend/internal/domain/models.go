package domain

import (
	"time"

	"github.com/google/uuid"
)

// User stores identity and login data shared by clinic account types.
type User struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Phone       string
	DateOfBirth string
	AvatarURL   string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserSummary includes the roles assigned to a user.
type UserSummary struct {
	User
	Roles []string
}

// FullName returns "First Last".
func (u User) FullName() string {
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// PersonSummary is the minimal identity embedded in appointment responses.
type PersonSummary struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

// FullName returns "First Last".
func (p PersonSummary) FullName() string {
	if p.LastName == "" {
		return p.FirstName
	}
	return p.FirstName + " " + p.LastName
}

// DoctorSummary is the doctor identity embedded in appointment responses.
type DoctorSummary struct {
	PersonSummary
	Specialty string
}

// Doctor is a doctor profile joined with its user record.
type Doctor struct {
	ProfileID           uuid.UUID
	UserID              uuid.UUID
	FirstName           string
	LastName            string
	Email               string
	Specialty           string
	Bio                 string
	ExperienceYears     int
	Languages           []string
	AvatarURL           string
	Timezone            string
	SlotMinutes         int
	IsAcceptingPatients bool
}

// FullName returns "First Last".
func (d Doctor) FullName() string {
	if d.LastName == "" {
		return d.FirstName
	}
	return d.FirstName + " " + d.LastName
}

// WorkingHour is one recurring availability interval for a weekday.
// Weekday follows Go's time.Weekday (0 = Sunday).
type WorkingHour struct {
	ID           uuid.UUID
	DoctorID     uuid.UUID
	Weekday      int
	StartMinutes int
	EndMinutes   int
}

// ScheduleBlock is intentional unavailability (lunch, leave, sick time).
type ScheduleBlock struct {
	ID        uuid.UUID
	DoctorID  uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	Reason    string
	CreatedBy *uuid.UUID
	CreatedAt time.Time
}

// Slot is a bookable time range in UTC.
type Slot struct {
	StartTime time.Time
	EndTime   time.Time
}

// BookableSlot is a validated slot plus the doctor's scheduling context.
type BookableSlot struct {
	DoctorID    uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	Timezone    string
	SlotMinutes int
}

// Appointment is the current state of a booking.
type Appointment struct {
	ID          uuid.UUID
	DoctorID    uuid.UUID
	PatientID   uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	Reason      string
	Status      string
	CreatedBy   *uuid.UUID
	CancelledAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Doctor  DoctorSummary
	Patient PersonSummary
}

// AppointmentEvent is one row of the append-only appointment history.
type AppointmentEvent struct {
	ID            int64
	AppointmentID uuid.UUID
	EventType     string
	FromStartTime *time.Time
	ToStartTime   *time.Time
	EventReason   string
	CreatedBy     *uuid.UUID
	ActorName     string
	CreatedAt     time.Time
}

// Role is an RBAC role.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	Scope       string
	IsSystem    bool
	Permissions []string
	CreatedAt   time.Time
}

// Permission is an RBAC permission definition.
type Permission struct {
	ID          uuid.UUID
	Scope       string
	Resource    string
	Action      string
	Code        string
	Description string
}

// PermissionOverride is a per-user allow or deny rule.
type PermissionOverride struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	PermissionID   uuid.UUID
	PermissionCode string
	Effect         string
	Note           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserAccess is the cached authorization snapshot for a user.
type UserAccess struct {
	Roles       []string
	Permissions []string
}

// Session is a refresh-token session.
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	FamilyID         uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time
}

// AuditLog is one administrative or domain audit entry.
type AuditLog struct {
	ID           int64
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	RequestID    string
	IP           string
	UserAgent    string
	CreatedAt    time.Time
}
