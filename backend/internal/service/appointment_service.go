package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
)

// AppointmentService implements booking, cancellation and rescheduling.
type AppointmentService struct {
	store     AppointmentStore
	validator SlotValidator
	holds     SlotHoldStore
	clock     clock.Clock
}

// NewAppointmentService creates the appointment service.
func NewAppointmentService(store AppointmentStore, validator SlotValidator, holds SlotHoldStore, clk clock.Clock) *AppointmentService {
	return &AppointmentService{store: store, validator: validator, holds: holds, clock: clk}
}

// BookForSelfInput books an appointment for the authenticated patient.
type BookForSelfInput struct {
	DoctorID  uuid.UUID
	StartTime time.Time
	Reason    string
	HoldToken string
}

// BookManualInput books an appointment on behalf of a patient.
type BookManualInput struct {
	DoctorID  uuid.UUID
	PatientID uuid.UUID
	StartTime time.Time
	Reason    string
}

// BookForSelf validates the slot and the caller's hold, then creates the
// appointment.
func (s *AppointmentService) BookForSelf(ctx context.Context, actor *domain.Actor, input BookForSelfInput, meta domain.RequestMeta) (*domain.Appointment, error) {
	reason, err := normalizeReason(input.Reason, true)
	if err != nil {
		return nil, err
	}
	slot, err := s.validator.ValidateSlot(ctx, input.DoctorID, input.StartTime)
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(slot.Timezone)
	if err != nil {
		return nil, domain.NewError(domain.CodeInternal, "the doctor's timezone is not configured correctly").WithCause(err)
	}

	hold, err := s.holds.Get(ctx, input.DoctorID, slot.StartTime, location)
	if err != nil {
		return nil, err
	}
	if hold == nil {
		return nil, domain.ErrHoldExpired
	}
	if hold.PatientID != actor.UserID {
		return nil, domain.ErrHoldNotOwned
	}
	if input.HoldToken != hold.Token {
		return nil, domain.ErrHoldNotOwned
	}

	appointment, err := s.store.CreateAppointment(ctx, input.DoctorID, actor.UserID, slot.StartTime, slot.EndTime, reason, actor.UserID, meta)
	if err != nil {
		return nil, err
	}
	_ = s.holds.Release(ctx, hold.Token, actor.UserID)
	return appointment, nil
}

// BookManual creates an appointment for a patient without requiring a hold.
func (s *AppointmentService) BookManual(ctx context.Context, actor *domain.Actor, input BookManualInput, meta domain.RequestMeta) (*domain.Appointment, error) {
	if actor.Can(domain.PermDoctorAppointmentsCreate) && !actor.Can(domain.PermAdminAppointmentsCreate) && input.DoctorID != actor.UserID {
		return nil, domain.ErrForbidden
	}
	reason, err := normalizeReason(input.Reason, true)
	if err != nil {
		return nil, err
	}
	patient, err := s.store.GetUserByID(ctx, input.PatientID)
	if err != nil {
		return nil, err
	}
	if patient.Status != domain.UserStatusActive {
		return nil, domain.ValidationField("patient_id", "the patient account is inactive")
	}
	slot, err := s.validator.ValidateSlot(ctx, input.DoctorID, input.StartTime)
	if err != nil {
		return nil, err
	}
	return s.store.CreateAppointment(ctx, input.DoctorID, input.PatientID, slot.StartTime, slot.EndTime, reason, actor.UserID, meta)
}

// acquireHold reserves a slot for a patient, reusing their active hold.
func (s *AppointmentService) acquireHold(ctx context.Context, patientID uuid.UUID, slot *domain.BookableSlot) (*cache.Hold, error) {
	return reserveSlot(ctx, s.holds, patientID, slot)
}

// reserveSlot holds a validated slot for a patient. An existing hold is reused
// when the patient owns it; a hold owned by anyone else makes the slot
// unavailable.
func reserveSlot(ctx context.Context, holds SlotHoldStore, patientID uuid.UUID, slot *domain.BookableSlot) (*cache.Hold, error) {
	location, err := time.LoadLocation(slot.Timezone)
	if err != nil {
		return nil, domain.NewError(domain.CodeInternal, "the doctor's timezone is not configured correctly").WithCause(err)
	}
	existing, err := holds.Get(ctx, slot.DoctorID, slot.StartTime, location)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.PatientID != patientID {
			return nil, domain.ErrSlotUnavailable
		}
		return existing, nil
	}
	return holds.Create(ctx, cache.HoldInput{
		DoctorID:  slot.DoctorID,
		PatientID: patientID,
		StartUTC:  slot.StartTime,
		Location:  location,
	})
}

func normalizeReason(reason string, required bool) (string, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		if required {
			return "", domain.ValidationField("reason", "is required")
		}
		return "", nil
	}
	if len(reason) > 1000 {
		return "", domain.ValidationField("reason", "must be at most 1000 characters")
	}
	return reason, nil
}
