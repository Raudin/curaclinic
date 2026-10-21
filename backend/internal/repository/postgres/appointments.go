package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

func mapGetAppointmentRow(row dbgen.GetAppointmentByIDRow) domain.Appointment {
	return domain.Appointment{
		ID:          row.ID,
		DoctorID:    row.DoctorID,
		PatientID:   row.PatientID,
		StartTime:   row.StartTime,
		EndTime:     row.EndTime,
		Reason:      row.Reason,
		Status:      string(row.Status),
		CreatedBy:   row.CreatedBy,
		CancelledAt: row.CancelledAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Doctor: domain.DoctorSummary{
			PersonSummary: domain.PersonSummary{
				ID:        row.DoctorID,
				FirstName: row.DoctorFirstName,
				LastName:  row.DoctorLastName,
				Email:     row.DoctorEmail,
			},
			Specialty: row.DoctorSpecialty,
		},
		Patient: domain.PersonSummary{
			ID:        row.PatientID,
			FirstName: row.PatientFirstName,
			LastName:  row.PatientLastName,
			Email:     row.PatientEmail,
		},
	}
}

func mapListAppointmentRow(row dbgen.ListAppointmentsRow) domain.Appointment {
	return domain.Appointment{
		ID:          row.ID,
		DoctorID:    row.DoctorID,
		PatientID:   row.PatientID,
		StartTime:   row.StartTime,
		EndTime:     row.EndTime,
		Reason:      row.Reason,
		Status:      string(row.Status),
		CreatedBy:   row.CreatedBy,
		CancelledAt: row.CancelledAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Doctor: domain.DoctorSummary{
			PersonSummary: domain.PersonSummary{
				ID:        row.DoctorID,
				FirstName: row.DoctorFirstName,
				LastName:  row.DoctorLastName,
				Email:     row.DoctorEmail,
			},
			Specialty: row.DoctorSpecialty,
		},
		Patient: domain.PersonSummary{
			ID:        row.PatientID,
			FirstName: row.PatientFirstName,
			LastName:  row.PatientLastName,
			Email:     row.PatientEmail,
		},
	}
}

// CreateAppointment inserts an appointment, its creation event and an audit
// record atomically.
func (r *Repo) CreateAppointment(ctx context.Context, doctorID, patientID uuid.UUID, start, end time.Time, reason string, actor uuid.UUID, meta domain.RequestMeta) (*domain.Appointment, error) {
	var appointment domain.Appointment
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.CreateAppointment(ctx, dbgen.CreateAppointmentParams{
			DoctorID:  doctorID,
			PatientID: patientID,
			StartTime: start,
			EndTime:   end,
			Reason:    reason,
			CreatedBy: &actor,
		})
		if err != nil {
			return mapAppointmentWriteError(err)
		}
		if _, eventErr := q.CreateAppointmentEvent(ctx, dbgen.CreateAppointmentEventParams{
			AppointmentID: row.ID,
			EventType:     dbgen.AppointmentEventTypeCreated,
			ToStartTime:   &row.StartTime,
			CreatedBy:     &actor,
		}); eventErr != nil {
			return eventErr
		}
		full, err := q.GetAppointmentByID(ctx, row.ID)
		if err != nil {
			return err
		}
		appointment = mapGetAppointmentRow(full)
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditAppointmentCreated,
			ResourceType: domain.ResourceAppointment,
			ResourceID:   row.ID.String(),
			Metadata:     map[string]any{"doctor_id": doctorID, "patient_id": patientID, "start_time": start, "end_time": end},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

// GetAppointment loads an appointment with doctor and patient summaries.
func (r *Repo) GetAppointment(ctx context.Context, id uuid.UUID) (*domain.Appointment, error) {
	row, err := r.q.GetAppointmentByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	appointment := mapGetAppointmentRow(row)
	return &appointment, nil
}

// ListAppointments returns a filtered page and the total count.
func (r *Repo) ListAppointments(ctx context.Context, filter domain.AppointmentFilter) ([]domain.Appointment, int64, error) {
	rows, err := r.q.ListAppointments(ctx, dbgen.ListAppointmentsParams{
		PatientID: filter.PatientID,
		DoctorID:  filter.DoctorID,
		Status:    filter.Status,
		FromTime:  filter.FromTime,
		ToTime:    filter.ToTime,
		EndsAfter: filter.EndsAfter,
		Ascending: filter.Ascending,
		Limit:     int32FromInt(filter.Limit),
		Offset:    int32FromInt(filter.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.Appointment, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapListAppointmentRow(row))
	}
	total, err := r.q.CountAppointments(ctx, dbgen.CountAppointmentsParams{
		PatientID: filter.PatientID,
		DoctorID:  filter.DoctorID,
		Status:    filter.Status,
		FromTime:  filter.FromTime,
		ToTime:    filter.ToTime,
		EndsAfter: filter.EndsAfter,
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ConfirmedSlotsInRange returns confirmed appointment ranges for a doctor.
func (r *Repo) ConfirmedSlotsInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.Slot, error) {
	rows, err := r.q.ListConfirmedAppointmentsInRange(ctx, dbgen.ListConfirmedAppointmentsInRangeParams{
		DoctorID:   doctorID,
		RangeStart: from,
		RangeEnd:   to,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Slot, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Slot{StartTime: row.StartTime, EndTime: row.EndTime})
	}
	return items, nil
}
