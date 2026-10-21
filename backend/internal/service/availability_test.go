package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
	"curaclinic/backend/internal/service"
)

func availabilityFixture(t *testing.T) (*service.AvailabilityService, *fakeAvailabilityStore, uuid.UUID, *time.Location) {
	t.Helper()
	location, err := time.LoadLocation("Africa/Nairobi")
	require.NoError(t, err)
	doctorID := uuid.New()

	doctor := &domain.Doctor{
		UserID:              doctorID,
		FirstName:           "Naomi",
		LastName:            "Kamau",
		Timezone:            "Africa/Nairobi",
		SlotMinutes:         30,
		IsAcceptingPatients: true,
	}
	store := &fakeAvailabilityStore{
		doctor: doctor,
		hours: map[int][]domain.WorkingHour{
			2: {
				{DoctorID: doctorID, Weekday: 2, StartMinutes: 9 * 60, EndMinutes: 12 * 60},
				{DoctorID: doctorID, Weekday: 2, StartMinutes: 13 * 60, EndMinutes: 16 * 60},
			},
		},
		appointments: []domain.Slot{{
			StartTime: time.Date(2026, 10, 20, 10, 0, 0, 0, location).UTC(),
			EndTime:   time.Date(2026, 10, 20, 10, 30, 0, 0, location).UTC(),
		}},
		blocks: []domain.ScheduleBlock{{
			DoctorID:  doctorID,
			StartTime: time.Date(2026, 10, 20, 13, 0, 0, 0, location).UTC(),
			EndTime:   time.Date(2026, 10, 20, 13, 30, 0, 0, location).UTC(),
		}},
	}
	now := time.Date(2026, 10, 20, 5, 0, 0, 0, time.UTC) // 08:00 in Nairobi
	holds := &fakeHoldLister{held: map[string]struct{}{
		cache.HoldKey(doctorID, time.Date(2026, 10, 20, 11, 0, 0, 0, location).UTC(), location): {},
	}}

	svc := service.NewAvailabilityService(store, holds, clock.NewFake(now), 30*time.Minute, 90)
	return svc, store, doctorID, location
}

func TestAvailabilityGeneratesAndFiltersSlots(t *testing.T) {
	svc, _, doctorID, _ := availabilityFixture(t)
	result, err := svc.Availability(context.Background(), doctorID, "2026-10-20")
	require.NoError(t, err)
	require.Equal(t, "2026-10-20", result.Date)
	require.Equal(t, "Africa/Nairobi", result.Timezone)
	require.Equal(t, 30, result.SlotMinutes)

	require.Len(t, result.Slots, 9)
	require.Equal(t, "2026-10-20T06:00:00Z", result.Slots[0].StartTime.Format(time.RFC3339))
	require.Equal(t, "2026-10-20T06:30:00Z", result.Slots[1].StartTime.Format(time.RFC3339))
	require.Equal(t, "2026-10-20T07:30:00Z", result.Slots[2].StartTime.Format(time.RFC3339))
	require.Equal(t, "2026-10-20T08:30:00Z", result.Slots[3].StartTime.Format(time.RFC3339))
	require.Equal(t, "2026-10-20T10:30:00Z", result.Slots[4].StartTime.Format(time.RFC3339))
}

func TestAvailabilityRejectsPastAndUnboundedDates(t *testing.T) {
	svc, _, doctorID, _ := availabilityFixture(t)
	_, err := svc.Availability(context.Background(), doctorID, "2026-10-19")
	require.True(t, domain.IsCode(err, domain.CodeValidation))

	_, err = svc.Availability(context.Background(), doctorID, "2027-06-01")
	require.True(t, domain.IsCode(err, domain.CodeValidation))

	_, err = svc.Availability(context.Background(), doctorID, "20-10-2026")
	require.True(t, domain.IsCode(err, domain.CodeValidation))
}

func TestValidateSlotEnforcesBookingRules(t *testing.T) {
	svc, _, doctorID, location := availabilityFixture(t)
	ctx := context.Background()

	valid := time.Date(2026, 10, 20, 9, 0, 0, 0, location)
	slot, err := svc.ValidateSlot(ctx, doctorID, valid)
	require.NoError(t, err)
	require.Equal(t, valid.UTC(), slot.StartTime)
	require.Equal(t, valid.Add(30*time.Minute).UTC(), slot.EndTime)

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 20, 8, 15, 0, 0, location))
	require.True(t, domain.IsCode(err, domain.CodeValidation), "lead time must reject 08:15")

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 20, 9, 10, 0, 0, location))
	require.True(t, domain.IsCode(err, domain.CodeValidation), "unaligned slot must be rejected")

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 20, 10, 0, 0, 0, location))
	require.ErrorIs(t, err, domain.ErrSlotUnavailable, "booked slot must be rejected")

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 20, 13, 0, 0, 0, location))
	require.ErrorIs(t, err, domain.ErrSlotUnavailable, "blocked slot must be rejected")

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 20, 17, 0, 0, 0, location))
	require.True(t, domain.IsCode(err, domain.CodeValidation), "outside working hours must be rejected")

	_, err = svc.ValidateSlot(ctx, doctorID, time.Date(2026, 10, 25, 9, 0, 0, 0, location))
	require.True(t, domain.IsCode(err, domain.CodeValidation), "Sunday has no working hours")
}

func TestAvailabilitySkipsCutoffSlots(t *testing.T) {
	_, store, doctorID, location := availabilityFixture(t)
	held := &fakeHoldLister{}
	// Move "now" to 08:40 Nairobi: 09:00 is only 20 minutes away.
	store.doctor.Timezone = "Africa/Nairobi"
	svc := service.NewAvailabilityService(store, held, clock.NewFake(
		time.Date(2026, 10, 20, 5, 40, 0, 0, time.UTC),
	), 30*time.Minute, 90)

	result, err := svc.Availability(context.Background(), doctorID, "2026-10-20")
	require.NoError(t, err)
	for _, slot := range result.Slots {
		require.False(t, slot.StartTime.Before(time.Date(2026, 10, 20, 6, 10, 0, 0, time.UTC)))
	}
	require.NotEqual(t, time.Date(2026, 10, 20, 9, 0, 0, 0, location).UTC(), result.Slots[0].StartTime)
}
