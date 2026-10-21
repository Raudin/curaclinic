package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
)

// DoctorStore is the persistence contract for doctors and schedules.
type DoctorStore interface {
	GetDoctor(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error)
	ListDoctors(ctx context.Context, filter domain.DoctorFilter) ([]domain.Doctor, error)
	CreateDoctorAccount(ctx context.Context, user domain.NewUser, roleName string, profile domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error)
	CreateDoctorProfileForUser(ctx context.Context, profile domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error)
	UpdateDoctorProfile(ctx context.Context, input domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error)
	WorkingHours(ctx context.Context, doctorID uuid.UUID) ([]domain.WorkingHour, error)
	WorkingHoursForWeekday(ctx context.Context, doctorID uuid.UUID, weekday int) ([]domain.WorkingHour, error)
	ReplaceWorkingHours(ctx context.Context, doctorID uuid.UUID, hours []domain.WorkingHour, actor uuid.UUID, meta domain.RequestMeta) error
	CreateScheduleBlock(ctx context.Context, block domain.ScheduleBlock, actor uuid.UUID, meta domain.RequestMeta) (*domain.ScheduleBlock, error)
	GetScheduleBlock(ctx context.Context, id uuid.UUID) (*domain.ScheduleBlock, error)
	DeleteScheduleBlock(ctx context.Context, id uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error
	ScheduleBlocks(ctx context.Context, doctorID uuid.UUID, from, to *time.Time) ([]domain.ScheduleBlock, error)
	ScheduleBlocksInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.ScheduleBlock, error)
	ConfirmedSlotsInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.Slot, error)
}

// AvailabilityStore is the persistence contract for availability calculation.
type AvailabilityStore interface {
	GetDoctor(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error)
	WorkingHoursForWeekday(ctx context.Context, doctorID uuid.UUID, weekday int) ([]domain.WorkingHour, error)
	ConfirmedSlotsInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.Slot, error)
	ScheduleBlocksInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.ScheduleBlock, error)
}

// HoldLister lists active holds for availability.
type HoldLister interface {
	ListHeld(ctx context.Context, doctorID uuid.UUID, date string) (map[string]struct{}, error)
}

// SlotHoldStore manages Redis holds.
type SlotHoldStore interface {
	Create(ctx context.Context, input cache.HoldInput) (*cache.Hold, error)
	Get(ctx context.Context, doctorID uuid.UUID, start time.Time, location *time.Location) (*cache.Hold, error)
	Release(ctx context.Context, token string, userID uuid.UUID) error
	ListHeld(ctx context.Context, doctorID uuid.UUID, date string) (map[string]struct{}, error)
}

// SlotValidator validates a requested slot against every booking rule.
type SlotValidator interface {
	ValidateSlot(ctx context.Context, doctorID uuid.UUID, start time.Time) (*domain.BookableSlot, error)
}

// AppointmentStore is the persistence contract for appointments.
type AppointmentStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetAppointment(ctx context.Context, id uuid.UUID) (*domain.Appointment, error)
	ListAppointments(ctx context.Context, filter domain.AppointmentFilter) ([]domain.Appointment, int64, error)
	CreateAppointment(ctx context.Context, doctorID, patientID uuid.UUID, start, end time.Time, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error)
	CancelAppointment(ctx context.Context, appointment domain.Appointment, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error)
	RescheduleAppointment(ctx context.Context, appointment domain.Appointment, start, end time.Time, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error)
	AppointmentEvents(ctx context.Context, appointmentID uuid.UUID) ([]domain.AppointmentEvent, error)
}

// AuditStore lists audit records.
type AuditStore interface {
	ListAuditLogs(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditLog, int64, error)
}
