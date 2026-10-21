package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
)

// AvailabilityResult is the free-slot response for a doctor and date.
type AvailabilityResult struct {
	DoctorID    uuid.UUID
	Date        string
	Timezone    string
	SlotMinutes int
	Slots       []domain.Slot
}

// AvailabilityService generates dynamically filtered slots.
type AvailabilityService struct {
	store       AvailabilityStore
	holds       HoldLister
	clock       clock.Clock
	leadTime    time.Duration
	horizonDays int
}

// NewAvailabilityService creates the availability service.
func NewAvailabilityService(store AvailabilityStore, holds HoldLister, clk clock.Clock, leadTime time.Duration, horizonDays int) *AvailabilityService {
	return &AvailabilityService{
		store:       store,
		holds:       holds,
		clock:       clk,
		leadTime:    leadTime,
		horizonDays: horizonDays,
	}
}

// Availability returns the bookable slots for a doctor on a calendar date.
func (s *AvailabilityService) Availability(ctx context.Context, doctorID uuid.UUID, date string) (*AvailabilityResult, error) {
	doctor, location, err := s.loadDoctor(ctx, doctorID)
	if err != nil {
		return nil, err
	}
	slotMinutes := slotMinutesFor(doctor)

	day, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil {
		return nil, domain.ValidationField("date", "must use the YYYY-MM-DD format")
	}
	now := s.clock.Now()
	today := startOfDay(now.In(location))
	if day.Before(today) {
		return nil, domain.ValidationField("date", "cannot be in the past")
	}
	if day.After(today.AddDate(0, 0, s.horizonDays)) {
		return nil, domain.ValidationField("date", "is beyond the booking horizon")
	}

	weekday := int(day.Weekday())
	hours, err := s.store.WorkingHoursForWeekday(ctx, doctorID, weekday)
	if err != nil {
		return nil, err
	}

	dayStart := startOfDay(day)
	dayEnd := dayStart.AddDate(0, 0, 1)
	appointments, err := s.store.ConfirmedSlotsInRange(ctx, doctorID, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return nil, err
	}
	blocks, err := s.store.ScheduleBlocksInRange(ctx, doctorID, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return nil, err
	}
	held, err := s.holds.ListHeld(ctx, doctorID, day.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	cutoff := now.Add(s.leadTime)
	available := make([]domain.Slot, 0, len(hours)*4)
	for _, slot := range generateSlots(day, hours, slotMinutes, location) {
		if slot.StartTime.Before(cutoff) {
			continue
		}
		if overlapsAny(slot.StartTime, slot.EndTime, appointments) {
			continue
		}
		if overlapsBlocks(slot.StartTime, slot.EndTime, blocks) {
			continue
		}
		if _, isHeld := held[cache.HoldKey(doctorID, slot.StartTime, location)]; isHeld {
			continue
		}
		available = append(available, slot)
	}

	return &AvailabilityResult{
		DoctorID:    doctorID,
		Date:        day.Format("2006-01-02"),
		Timezone:    doctor.Timezone,
		SlotMinutes: slotMinutes,
		Slots:       available,
	}, nil
}

func (s *AvailabilityService) loadDoctor(ctx context.Context, doctorID uuid.UUID) (*domain.Doctor, *time.Location, error) {
	doctor, err := s.store.GetDoctor(ctx, doctorID)
	if err != nil {
		return nil, nil, err
	}
	if !doctor.IsAcceptingPatients {
		return nil, nil, domain.NewError(domain.CodeNotFound, "this doctor is not accepting appointments")
	}
	location, err := time.LoadLocation(doctor.Timezone)
	if err != nil {
		return nil, nil, domain.NewError(domain.CodeInternal, "the doctor's timezone is not configured correctly").WithCause(err)
	}
	return doctor, location, nil
}

func slotMinutesFor(doctor *domain.Doctor) int {
	if doctor.SlotMinutes >= 5 && doctor.SlotMinutes <= 180 {
		return doctor.SlotMinutes
	}
	return 30
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func generateSlots(day time.Time, hours []domain.WorkingHour, slotMinutes int, location *time.Location) []domain.Slot {
	if slotMinutes <= 0 {
		return nil
	}
	year, month, date := day.Date()
	slots := make([]domain.Slot, 0, len(hours)*4)
	for _, hour := range hours {
		for minutes := hour.StartMinutes; minutes+slotMinutes <= hour.EndMinutes; minutes += slotMinutes {
			hourOfDay := minutes / 60
			minuteOfHour := minutes % 60
			start := time.Date(year, month, date, hourOfDay, minuteOfHour, 0, 0, location)
			// Skip wall-clock times that do not exist because of a DST jump.
			if start.Hour() != hourOfDay || start.Minute() != minuteOfHour {
				continue
			}
			end := start.Add(time.Duration(slotMinutes) * time.Minute)
			slots = append(slots, domain.Slot{StartTime: start.UTC(), EndTime: end.UTC()})
		}
	}
	return slots
}

func overlapsAny(start, end time.Time, slots []domain.Slot) bool {
	for _, slot := range slots {
		if start.Before(slot.EndTime) && slot.StartTime.Before(end) {
			return true
		}
	}
	return false
}

func overlapsBlocks(start, end time.Time, blocks []domain.ScheduleBlock) bool {
	for _, block := range blocks {
		if start.Before(block.EndTime) && block.StartTime.Before(end) {
			return true
		}
	}
	return false
}
