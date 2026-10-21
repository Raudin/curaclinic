package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
)

func TestMemoryHoldLifecycle(t *testing.T) {
	location := nairobi(t)
	fakeClock := clock.NewFake(time.Date(2026, 6, 20, 6, 0, 0, 0, time.UTC))
	store := cache.NewMemoryHoldStore(fakeClock, 5*time.Minute)
	ctx := context.Background()
	doctorID := uuid.New()
	patientID := uuid.New()
	start := time.Date(2026, 6, 20, 9, 30, 0, 0, location)

	hold, err := store.Create(ctx, cache.HoldInput{
		DoctorID: doctorID, PatientID: patientID, StartUTC: start, Location: location,
	})
	require.NoError(t, err)
	require.Equal(t, "hold:"+doctorID.String()+":2026-06-20:0930", hold.Key)

	_, err = store.Create(ctx, cache.HoldInput{
		DoctorID: doctorID, PatientID: uuid.New(), StartUTC: start, Location: location,
	})
	require.ErrorIs(t, err, domain.ErrSlotUnavailable)

	held, err := store.ListHeld(ctx, doctorID, "2026-06-20")
	require.NoError(t, err)
	require.Contains(t, held, hold.Key)

	got, err := store.Get(ctx, doctorID, start, location)
	require.NoError(t, err)
	require.Equal(t, hold.Token, got.Token)

	require.ErrorIs(t, store.Release(ctx, hold.Token, uuid.New()), domain.ErrHoldNotOwned)
	require.NoError(t, store.Release(ctx, hold.Token, patientID))

	got, err = store.Get(ctx, doctorID, start, location)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestMemoryHoldExpiry(t *testing.T) {
	location := nairobi(t)
	fakeClock := clock.NewFake(time.Date(2026, 6, 20, 6, 0, 0, 0, time.UTC))
	store := cache.NewMemoryHoldStore(fakeClock, 5*time.Minute)
	ctx := context.Background()
	doctorID := uuid.New()
	start := time.Date(2026, 6, 20, 10, 0, 0, 0, location)

	hold, err := store.Create(ctx, cache.HoldInput{
		DoctorID: doctorID, PatientID: uuid.New(), StartUTC: start, Location: location,
	})
	require.NoError(t, err)
	require.NotEmpty(t, hold.Token)

	fakeClock.Advance(5*time.Minute + time.Second)

	got, err := store.Get(ctx, doctorID, start, location)
	require.NoError(t, err)
	require.Nil(t, got)

	held, err := store.ListHeld(ctx, doctorID, "2026-06-20")
	require.NoError(t, err)
	require.Empty(t, held)
}
