package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// IntervalInput is one recurring availability interval.
type IntervalInput struct {
	Weekday      int
	StartMinutes int
	EndMinutes   int
}

// WorkingHours returns a doctor's recurring schedule.
func (s *DoctorService) WorkingHours(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID) ([]domain.WorkingHour, error) {
	if !canManageDoctorSchedule(actor, doctorID, domain.PermDoctorScheduleView) {
		return nil, domain.ErrForbidden
	}
	return s.store.WorkingHours(ctx, doctorID)
}

// ReplaceWorkingHours validates and replaces the weekly schedule.
func (s *DoctorService) ReplaceWorkingHours(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID, intervals []IntervalInput, meta domain.RequestMeta) ([]domain.WorkingHour, error) {
	if !canManageDoctorSchedule(actor, doctorID, domain.PermDoctorWorkingHours) {
		return nil, domain.ErrForbidden
	}
	doctor, err := s.store.GetDoctor(ctx, doctorID)
	if err != nil {
		return nil, err
	}
	slotMinutes := slotMinutesFor(doctor)
	if len(intervals) > 50 {
		return nil, domain.Validation("a schedule can contain at most 50 intervals", nil)
	}

	hours := make([]domain.WorkingHour, 0, len(intervals))
	for _, interval := range intervals {
		if interval.Weekday < 0 || interval.Weekday > 6 {
			return nil, domain.ValidationField("intervals", "weekday must be between 0 and 6")
		}
		if interval.StartMinutes < 0 || interval.EndMinutes > 1440 || interval.StartMinutes >= interval.EndMinutes {
			return nil, domain.ValidationField("intervals", "each interval must satisfy 00:00 <= start < end <= 24:00")
		}
		if interval.EndMinutes-interval.StartMinutes < slotMinutes {
			return nil, domain.ValidationField("intervals", "each interval must be at least one slot long")
		}
		if interval.StartMinutes%slotMinutes != 0 || interval.EndMinutes%slotMinutes != 0 {
			return nil, domain.ValidationField("intervals", "interval boundaries must align to the slot size")
		}
		hours = append(hours, domain.WorkingHour{
			ID:           uuid.New(),
			DoctorID:     doctorID,
			Weekday:      interval.Weekday,
			StartMinutes: interval.StartMinutes,
			EndMinutes:   interval.EndMinutes,
		})
	}
	if err := validateNoOverlaps(hours); err != nil {
		return nil, err
	}
	if err := s.store.ReplaceWorkingHours(ctx, doctorID, hours, actor.UserID, meta); err != nil {
		return nil, err
	}
	return s.store.WorkingHours(ctx, doctorID)
}

func validateNoOverlaps(hours []domain.WorkingHour) error {
	sorted := make([]domain.WorkingHour, len(hours))
	copy(sorted, hours)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Weekday != sorted[j].Weekday {
			return sorted[i].Weekday < sorted[j].Weekday
		}
		return sorted[i].StartMinutes < sorted[j].StartMinutes
	})
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Weekday == sorted[i-1].Weekday && sorted[i].StartMinutes < sorted[i-1].EndMinutes {
			return domain.ValidationField("intervals", "intervals on the same weekday must not overlap")
		}
	}
	return nil
}

// ListBlocks returns schedule blocks for a doctor.
func (s *DoctorService) ListBlocks(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID, from, to *time.Time) ([]domain.ScheduleBlock, error) {
	if !canManageDoctorSchedule(actor, doctorID, domain.PermDoctorScheduleView) {
		return nil, domain.ErrForbidden
	}
	return s.store.ScheduleBlocks(ctx, doctorID, from, to)
}

// CreateBlock creates intentional unavailability.
func (s *DoctorService) CreateBlock(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID, start, end time.Time, reason string, meta domain.RequestMeta) (*domain.ScheduleBlock, error) {
	if !canManageDoctorSchedule(actor, doctorID, domain.PermDoctorBlockTime) {
		return nil, domain.ErrForbidden
	}
	start, end = start.UTC(), end.UTC()
	if start.Second() != 0 || start.Nanosecond() != 0 || end.Second() != 0 || end.Nanosecond() != 0 {
		return nil, domain.ValidationField("start_time", "must be aligned to a whole minute")
	}
	if !end.After(start) {
		return nil, domain.ValidationField("end_time", "must be after start_time")
	}
	if end.Sub(start) > 31*24*time.Hour {
		return nil, domain.ValidationField("end_time", "a single block cannot exceed 31 days")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, domain.ValidationField("reason", "is required")
	}
	if len(reason) > 200 {
		return nil, domain.ValidationField("reason", "must be at most 200 characters")
	}

	booked, err := s.store.ConfirmedSlotsInRange(ctx, doctorID, start, end)
	if err != nil {
		return nil, err
	}
	if overlapsAny(start, end, booked) {
		return nil, domain.NewError(domain.CodeConflict, "this time contains confirmed appointments; reschedule them before blocking the time")
	}

	return s.store.CreateScheduleBlock(ctx, domain.ScheduleBlock{
		DoctorID:  doctorID,
		StartTime: start,
		EndTime:   end,
		Reason:    reason,
	}, actor.UserID, meta)
}

// DeleteBlock removes one schedule block.
func (s *DoctorService) DeleteBlock(ctx context.Context, actor *domain.Actor, blockID uuid.UUID, meta domain.RequestMeta) error {
	block, err := s.store.GetScheduleBlock(ctx, blockID)
	if err != nil {
		return err
	}
	if !canManageDoctorSchedule(actor, block.DoctorID, domain.PermDoctorRemoveBlock) {
		return domain.ErrForbidden
	}
	return s.store.DeleteScheduleBlock(ctx, blockID, actor.UserID, meta)
}

func canManageDoctorSchedule(actor *domain.Actor, doctorID uuid.UUID, doctorPermission string) bool {
	return actor != nil && (actor.Can(domain.PermAdminWorkingHours) ||
		(actor.UserID == doctorID && containsString(actor.Roles, "doctor") && actor.Can(doctorPermission)))
}
