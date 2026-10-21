package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

func mapWorkingHours(rows []dbgen.ListWorkingHoursRow) []domain.WorkingHour {
	items := make([]domain.WorkingHour, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.WorkingHour{
			ID:           row.ID,
			DoctorID:     row.DoctorID,
			Weekday:      int(row.Weekday),
			StartMinutes: int(row.StartMinutes),
			EndMinutes:   int(row.EndMinutes),
		})
	}
	return items
}

func mapScheduleBlock(row dbgen.ScheduleBlock) domain.ScheduleBlock {
	return domain.ScheduleBlock{
		ID:        row.ID,
		DoctorID:  row.DoctorID,
		StartTime: row.StartTime,
		EndTime:   row.EndTime,
		Reason:    row.Reason,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
}

// WorkingHours returns every recurring interval for a doctor.
func (r *Repo) WorkingHours(ctx context.Context, doctorID uuid.UUID) ([]domain.WorkingHour, error) {
	rows, err := r.q.ListWorkingHours(ctx, doctorID)
	if err != nil {
		return nil, err
	}
	return mapWorkingHours(rows), nil
}

// WorkingHoursForWeekday returns the intervals for one weekday.
func (r *Repo) WorkingHoursForWeekday(ctx context.Context, doctorID uuid.UUID, weekday int) ([]domain.WorkingHour, error) {
	rows, err := r.q.ListWorkingHoursForWeekday(ctx, dbgen.ListWorkingHoursForWeekdayParams{
		DoctorID: doctorID,
		Weekday:  weekdayInt16(weekday),
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.WorkingHour, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.WorkingHour{
			ID:           row.ID,
			DoctorID:     row.DoctorID,
			Weekday:      int(row.Weekday),
			StartMinutes: int(row.StartMinutes),
			EndMinutes:   int(row.EndMinutes),
		})
	}
	return items, nil
}

// ReplaceWorkingHours replaces a doctor's weekly schedule atomically.
func (r *Repo) ReplaceWorkingHours(ctx context.Context, doctorID uuid.UUID, hours []domain.WorkingHour, actor uuid.UUID, meta domain.RequestMeta) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		if err := q.SoftDeleteWorkingHours(ctx, doctorID); err != nil {
			return err
		}
		for _, hour := range hours {
			if err := q.InsertWorkingHour(ctx, dbgen.InsertWorkingHourParams{
				DoctorID:     doctorID,
				Weekday:      weekdayInt16(hour.Weekday),
				StartSeconds: int32FromInt(hour.StartMinutes * 60),
				EndSeconds:   int32FromInt(hour.EndMinutes * 60),
				CreatedBy:    &actor,
			}); err != nil {
				return err
			}
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditWorkingHoursUpdated,
			ResourceType: domain.ResourceSchedule,
			ResourceID:   doctorID.String(),
			Metadata:     map[string]any{"intervals": hours},
			Meta:         meta,
		})
	})
}

// CreateScheduleBlock inserts a block of unavailability.
func (r *Repo) CreateScheduleBlock(ctx context.Context, block domain.ScheduleBlock, actor uuid.UUID, meta domain.RequestMeta) (*domain.ScheduleBlock, error) {
	var created domain.ScheduleBlock
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.CreateScheduleBlock(ctx, dbgen.CreateScheduleBlockParams{
			DoctorID:  block.DoctorID,
			StartTime: block.StartTime,
			EndTime:   block.EndTime,
			Reason:    block.Reason,
			CreatedBy: &actor,
		})
		if err != nil {
			return mapBlockError(err)
		}
		created = mapScheduleBlock(row)
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditBlockCreated,
			ResourceType: domain.ResourceSchedule,
			ResourceID:   row.ID.String(),
			Metadata:     map[string]any{"doctor_id": block.DoctorID, "start_time": block.StartTime, "end_time": block.EndTime, "reason": block.Reason},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// GetScheduleBlock loads one block.
func (r *Repo) GetScheduleBlock(ctx context.Context, id uuid.UUID) (*domain.ScheduleBlock, error) {
	row, err := r.q.GetScheduleBlockByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	block := mapScheduleBlock(row)
	return &block, nil
}

// DeleteScheduleBlock soft-deletes one block.
func (r *Repo) DeleteScheduleBlock(ctx context.Context, id uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		if err := q.SoftDeleteScheduleBlock(ctx, id); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditBlockDeleted,
			ResourceType: domain.ResourceSchedule,
			ResourceID:   id.String(),
			Metadata:     map[string]any{},
			Meta:         meta,
		})
	})
}

// ScheduleBlocksInRange returns blocks overlapping [from, to).
func (r *Repo) ScheduleBlocksInRange(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]domain.ScheduleBlock, error) {
	rows, err := r.q.ListScheduleBlocksInRange(ctx, dbgen.ListScheduleBlocksInRangeParams{
		DoctorID:   doctorID,
		RangeStart: from,
		RangeEnd:   to,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.ScheduleBlock, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapScheduleBlock(row))
	}
	return items, nil
}

// ScheduleBlocks lists a doctor's blocks with optional time filters.
func (r *Repo) ScheduleBlocks(ctx context.Context, doctorID uuid.UUID, from, to *time.Time) ([]domain.ScheduleBlock, error) {
	rows, err := r.q.ListScheduleBlocks(ctx, dbgen.ListScheduleBlocksParams{
		DoctorID: doctorID,
		FromTime: from,
		ToTime:   to,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.ScheduleBlock, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapScheduleBlock(row))
	}
	return items, nil
}
