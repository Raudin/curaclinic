package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// ListAppointmentsInput narrows an appointment listing.
type ListAppointmentsInput struct {
	Scope     string
	As        string
	DoctorID  *uuid.UUID
	PatientID *uuid.UUID
	FromTime  *time.Time
	ToTime    *time.Time
	Order     string
	Limit     int
	Offset    int
}

// List returns appointments the actor is allowed to see.
func (s *AppointmentService) List(ctx context.Context, actor *domain.Actor, input ListAppointmentsInput) ([]domain.Appointment, int64, error) {
	limit, offset := normalizePagination(input.Limit, input.Offset)
	filter := domain.AppointmentFilter{
		Limit:    limit,
		Offset:   offset,
		FromTime: input.FromTime,
		ToTime:   input.ToTime,
	}

	viewAs := input.As
	if viewAs == "" {
		switch {
		case actor.Can(domain.PermAdminAppointmentsView):
			viewAs = "admin"
		case containsString(actor.Roles, "doctor") && actor.Can(domain.PermDoctorAppointmentsViewOwn):
			viewAs = "doctor"
		default:
			viewAs = "patient"
		}
	}
	switch viewAs {
	case "admin":
		if !actor.Can(domain.PermAdminAppointmentsView) {
			return nil, 0, domain.ErrForbidden
		}
		filter.DoctorID = input.DoctorID
		filter.PatientID = input.PatientID
	case "doctor":
		if !containsString(actor.Roles, "doctor") || !actor.Can(domain.PermDoctorAppointmentsViewOwn) {
			return nil, 0, domain.ErrForbidden
		}
		doctorID := actor.UserID
		filter.DoctorID = &doctorID
	case "patient":
		if !containsString(actor.Roles, "patient") || !actor.Can(domain.PermAppointmentsViewOwn) {
			return nil, 0, domain.ErrForbidden
		}
		patientID := actor.UserID
		filter.PatientID = &patientID
	default:
		return nil, 0, domain.ValidationField("as", "must be patient, doctor or admin")
	}

	now := s.clock.Now()
	scope := input.Scope
	if scope == "" {
		scope = "upcoming"
	}
	switch scope {
	case "upcoming":
		status := domain.AppointmentStatusConfirmed
		filter.Status = &status
		// Upcoming includes appointments still in progress, so they never
		// fall between the upcoming and past listings.
		if filter.FromTime == nil {
			filter.EndsAfter = &now
		}
	case "past":
		status := domain.AppointmentStatusConfirmed
		filter.Status = &status
		if filter.ToTime == nil {
			filter.ToTime = &now
		}
	case "cancelled":
		status := domain.AppointmentStatusCancelled
		filter.Status = &status
	case "all":
		// no status constraint
	default:
		return nil, 0, domain.ValidationField("scope", "must be upcoming, past, cancelled or all")
	}

	switch input.Order {
	case "asc":
		filter.Ascending = true
	case "desc":
		filter.Ascending = false
	case "":
		filter.Ascending = scope == "upcoming"
	default:
		return nil, 0, domain.ValidationField("order", "must be asc or desc")
	}

	return s.store.ListAppointments(ctx, filter)
}

// Get returns one appointment when the actor may view it.
func (s *AppointmentService) Get(ctx context.Context, actor *domain.Actor, appointmentID uuid.UUID) (*domain.Appointment, error) {
	appointment, err := s.store.GetAppointment(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if !canView(actor, appointment) {
		return nil, domain.ErrForbidden
	}
	return appointment, nil
}

// Events returns the audit history for an appointment the actor may view.
func (s *AppointmentService) Events(ctx context.Context, actor *domain.Actor, appointmentID uuid.UUID) ([]domain.AppointmentEvent, error) {
	if _, err := s.Get(ctx, actor, appointmentID); err != nil {
		return nil, err
	}
	return s.store.AppointmentEvents(ctx, appointmentID)
}

func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
