package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// CancelAppointment marks a confirmed appointment cancelled and records the
// event and audit entry in one transaction.
func (r *Repo) CancelAppointment(ctx context.Context, appointment domain.Appointment, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error) {
	var updated domain.Appointment
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.UpdateAppointmentStatus(ctx, dbgen.UpdateAppointmentStatusParams{
			ID:     appointment.ID,
			Status: dbgen.AppointmentStatusCancelled,
		})
		if err != nil {
			return mapAppointmentWriteError(err)
		}
		if _, eventErr := q.CreateAppointmentEvent(ctx, dbgen.CreateAppointmentEventParams{
			AppointmentID: row.ID,
			EventType:     dbgen.AppointmentEventTypeCancelled,
			FromStartTime: &appointment.StartTime,
			EventReason:   optionalString(reason),
			CreatedBy:     &actor,
		}); eventErr != nil {
			return eventErr
		}
		full, err := q.GetAppointmentByID(ctx, row.ID)
		if err != nil {
			return err
		}
		updated = mapGetAppointmentRow(full)

		metadata := map[string]any{"reason": reason, "start_time": appointment.StartTime}
		if appointment.PatientID != actor {
			metadata["patient_id"] = appointment.PatientID
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditAppointmentCancelled,
			ResourceType: domain.ResourceAppointment,
			ResourceID:   row.ID.String(),
			Metadata:     metadata,
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// RescheduleAppointment moves an appointment and records the history.
func (r *Repo) RescheduleAppointment(ctx context.Context, appointment domain.Appointment, start, end time.Time, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error) {
	var updated domain.Appointment
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.UpdateAppointmentTimes(ctx, dbgen.UpdateAppointmentTimesParams{
			ID:        appointment.ID,
			StartTime: start,
			EndTime:   end,
		})
		if err != nil {
			return mapAppointmentWriteError(err)
		}
		if _, eventErr := q.CreateAppointmentEvent(ctx, dbgen.CreateAppointmentEventParams{
			AppointmentID: row.ID,
			EventType:     dbgen.AppointmentEventTypeRescheduled,
			FromStartTime: &appointment.StartTime,
			ToStartTime:   &start,
			EventReason:   optionalString(reason),
			CreatedBy:     &actor,
		}); eventErr != nil {
			return eventErr
		}
		full, err := q.GetAppointmentByID(ctx, row.ID)
		if err != nil {
			return err
		}
		updated = mapGetAppointmentRow(full)

		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditAppointmentRescheduled,
			ResourceType: domain.ResourceAppointment,
			ResourceID:   row.ID.String(),
			Metadata:     map[string]any{"from_start_time": appointment.StartTime, "to_start_time": start, "reason": reason},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// AppointmentEvents returns the append-only history for an appointment.
func (r *Repo) AppointmentEvents(ctx context.Context, appointmentID uuid.UUID) ([]domain.AppointmentEvent, error) {
	rows, err := r.q.ListAppointmentEvents(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.AppointmentEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.AppointmentEvent{
			ID:            row.ID,
			AppointmentID: row.AppointmentID,
			EventType:     string(row.EventType),
			FromStartTime: row.FromStartTime,
			ToStartTime:   row.ToStartTime,
			EventReason:   stringValue(row.EventReason),
			CreatedBy:     row.CreatedBy,
			ActorName:     row.ActorName,
			CreatedAt:     row.CreatedAt,
		})
	}
	return items, nil
}
