package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// ValidateSlot enforces every booking rule for a single requested time.
func (s *AvailabilityService) ValidateSlot(ctx context.Context, doctorID uuid.UUID, start time.Time) (*domain.BookableSlot, error) {
	doctor, location, err := s.loadDoctor(ctx, doctorID)
	if err != nil {
		return nil, err
	}
	slotMinutes := slotMinutesFor(doctor)
	start = start.UTC()

	if start.Second() != 0 || start.Nanosecond() != 0 {
		return nil, domain.ValidationField("start_time", "must be aligned to a whole minute")
	}
	now := s.clock.Now()
	if start.Before(now.Add(s.leadTime)) {
		return nil, domain.ValidationField("start_time", "must be at least "+humanDuration(s.leadTime)+" from now")
	}

	local := start.In(location)
	day := startOfDay(local)
	today := startOfDay(now.In(location))
	if day.Before(today) {
		return nil, domain.ValidationField("start_time", "cannot be in the past")
	}
	if day.After(today.AddDate(0, 0, s.horizonDays)) {
		return nil, domain.ValidationField("start_time", "is beyond the booking horizon")
	}

	localMinutes := local.Hour()*60 + local.Minute()
	endLocalMinutes := localMinutes + slotMinutes
	hours, err := s.store.WorkingHoursForWeekday(ctx, doctorID, int(local.Weekday()))
	if err != nil {
		return nil, err
	}

	insideHours := false
	for _, hour := range hours {
		if localMinutes >= hour.StartMinutes &&
			endLocalMinutes <= hour.EndMinutes &&
			(localMinutes-hour.StartMinutes)%slotMinutes == 0 {
			insideHours = true
			break
		}
	}
	if !insideHours {
		return nil, domain.ValidationField("start_time", "is outside the doctor's working hours")
	}

	end := start.Add(time.Duration(slotMinutes) * time.Minute)

	blocks, err := s.store.ScheduleBlocksInRange(ctx, doctorID, start, end)
	if err != nil {
		return nil, err
	}
	if overlapsBlocks(start, end, blocks) {
		return nil, domain.ErrSlotUnavailable
	}

	appointments, err := s.store.ConfirmedSlotsInRange(ctx, doctorID, start, end)
	if err != nil {
		return nil, err
	}
	if overlapsAny(start, end, appointments) {
		return nil, domain.ErrSlotUnavailable
	}

	return &domain.BookableSlot{
		DoctorID:    doctorID,
		StartTime:   start,
		EndTime:     end,
		Timezone:    doctor.Timezone,
		SlotMinutes: slotMinutes,
	}, nil
}

func humanDuration(duration time.Duration) string {
	if duration%time.Minute == 0 {
		return duration.String()
	}
	return duration.Round(time.Second).String()
}
