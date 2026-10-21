package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// Cancel marks an appointment cancelled after authorization.
func (s *AppointmentService) Cancel(ctx context.Context, actor *domain.Actor, appointmentID uuid.UUID, reason string, meta domain.RequestMeta) (*domain.Appointment, error) {
	reason, err := normalizeReason(reason, true)
	if err != nil {
		return nil, err
	}
	appointment, err := s.store.GetAppointment(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appointment.Status == domain.AppointmentStatusCancelled {
		return nil, domain.NewError(domain.CodeConflict, "this appointment is already cancelled")
	}
	if !canCancel(actor, appointment) {
		return nil, domain.ErrForbidden
	}
	if !appointment.StartTime.After(s.clock.Now()) {
		return nil, domain.NewError(domain.CodeConflict, "an appointment that has already started cannot be cancelled")
	}
	return s.store.CancelAppointment(ctx, *appointment, reason, actor.UserID, meta)
}

// Reschedule moves an appointment after validating the new slot and acquiring a
// short-lived hold so the slot cannot be taken mid-transaction.
func (s *AppointmentService) Reschedule(ctx context.Context, actor *domain.Actor, appointmentID uuid.UUID, newStart time.Time, reason string, meta domain.RequestMeta) (*domain.Appointment, error) {
	reason, err := normalizeReason(reason, false)
	if err != nil {
		return nil, err
	}
	appointment, err := s.store.GetAppointment(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appointment.Status == domain.AppointmentStatusCancelled {
		return nil, domain.NewError(domain.CodeConflict, "a cancelled appointment cannot be rescheduled")
	}
	if !canReschedule(actor, appointment) {
		return nil, domain.ErrForbidden
	}
	if !appointment.StartTime.After(s.clock.Now()) {
		return nil, domain.NewError(domain.CodeConflict, "an appointment that has already started cannot be rescheduled")
	}

	slot, err := s.validator.ValidateSlot(ctx, appointment.DoctorID, newStart)
	if err != nil {
		return nil, err
	}
	if slot.StartTime.Equal(appointment.StartTime) {
		return nil, domain.ValidationField("start_time", "must differ from the current appointment time")
	}

	hold, err := s.acquireHold(ctx, appointment.PatientID, slot)
	if err != nil {
		return nil, err
	}
	updated, err := s.store.RescheduleAppointment(ctx, *appointment, slot.StartTime, slot.EndTime, reason, actor.UserID, meta)
	if err != nil {
		_ = s.holds.Release(ctx, hold.Token, appointment.PatientID)
		return nil, err
	}
	_ = s.holds.Release(ctx, hold.Token, appointment.PatientID)
	return updated, nil
}

func canCancel(actor *domain.Actor, appointment *domain.Appointment) bool {
	if actor.Can(domain.PermAdminAppointmentsCancel) {
		return true
	}
	if appointment.PatientID == actor.UserID && actor.Can(domain.PermAppointmentsCancelOwn) {
		return true
	}
	return appointment.DoctorID == actor.UserID && actor.Can(domain.PermDoctorAppointmentsCancel)
}

func canReschedule(actor *domain.Actor, appointment *domain.Appointment) bool {
	if actor.Can(domain.PermAdminAppointmentsMove) {
		return true
	}
	if appointment.PatientID == actor.UserID && actor.Can(domain.PermAppointmentsRescheduleOwn) {
		return true
	}
	return appointment.DoctorID == actor.UserID && actor.Can(domain.PermDoctorAppointmentsMove)
}

func canView(actor *domain.Actor, appointment *domain.Appointment) bool {
	if actor.Can(domain.PermAdminAppointmentsView) {
		return true
	}
	return appointment.PatientID == actor.UserID && actor.Can(domain.PermAppointmentsViewOwn) ||
		appointment.DoctorID == actor.UserID && actor.Can(domain.PermDoctorAppointmentsViewOwn)
}
