package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

type fakeAvailabilityStore struct {
	doctor       *domain.Doctor
	hours        map[int][]domain.WorkingHour
	appointments []domain.Slot
	blocks       []domain.ScheduleBlock
}

func (f *fakeAvailabilityStore) GetDoctor(_ context.Context, _ uuid.UUID) (*domain.Doctor, error) {
	if f.doctor == nil {
		return nil, domain.ErrNotFound
	}
	return f.doctor, nil
}

func (f *fakeAvailabilityStore) WorkingHoursForWeekday(_ context.Context, _ uuid.UUID, weekday int) ([]domain.WorkingHour, error) {
	return f.hours[weekday], nil
}

func (f *fakeAvailabilityStore) ConfirmedSlotsInRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]domain.Slot, error) {
	return f.appointments, nil
}

func (f *fakeAvailabilityStore) ScheduleBlocksInRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]domain.ScheduleBlock, error) {
	return f.blocks, nil
}

type fakeHoldLister struct {
	held map[string]struct{}
}

func (f *fakeHoldLister) ListHeld(_ context.Context, _ uuid.UUID, _ string) (map[string]struct{}, error) {
	if f.held == nil {
		return map[string]struct{}{}, nil
	}
	return f.held, nil
}
