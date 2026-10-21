package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
)

func newHoldStore(t *testing.T) (*cache.HoldStore, *miniredis.Miniredis, *clock.Fake) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	fakeClock := clock.NewFake(time.Date(2026, 6, 20, 6, 0, 0, 0, time.UTC))
	return cache.NewHoldStore(client, fakeClock, 5*time.Minute), mr, fakeClock
}

func nairobi(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Africa/Nairobi")
	require.NoError(t, err)
	return location
}

func TestHoldCreateIsAtomicAndIndexed(t *testing.T) {
	store, _, _ := newHoldStore(t)
	ctx := context.Background()
	location := nairobi(t)
	doctorID := uuid.New()
	patientID := uuid.New()
	start := time.Date(2026, 6, 20, 9, 30, 0, 0, location)

	hold, err := store.Create(ctx, cache.HoldInput{
		DoctorID:  doctorID,
		PatientID: patientID,
		StartUTC:  start,
		Location:  location,
	})
	require.NoError(t, err)
	require.NotNil(t, hold)
	require.Equal(t, patientID, hold.PatientID)
	require.Equal(t, "2026-06-20", hold.Date)
	require.Equal(t, "hold:"+doctorID.String()+":2026-06-20:0930", hold.Key)

	_, err = store.Create(ctx, cache.HoldInput{
		DoctorID:  doctorID,
		PatientID: uuid.New(),
		StartUTC:  start,
		Location:  location,
	})
	require.ErrorIs(t, err, domain.ErrSlotUnavailable)

	held, err := store.ListHeld(ctx, doctorID, "2026-06-20")
	require.NoError(t, err)
	require.Contains(t, held, hold.Key)
}

func TestHoldOwnershipAndRelease(t *testing.T) {
	store, _, _ := newHoldStore(t)
	ctx := context.Background()
	location := nairobi(t)
	doctorID := uuid.New()
	patientID := uuid.New()
	start := time.Date(2026, 6, 20, 10, 0, 0, 0, location)

	hold, err := store.Create(ctx, cache.HoldInput{
		DoctorID: doctorID, PatientID: patientID, StartUTC: start, Location: location,
	})
	require.NoError(t, err)

	err = store.Release(ctx, hold.Token, uuid.New())
	require.ErrorIs(t, err, domain.ErrHoldNotOwned)

	require.NoError(t, store.Release(ctx, hold.Token, patientID))

	got, err := store.Get(ctx, doctorID, start, location)
	require.NoError(t, err)
	require.Nil(t, got)

	held, err := store.ListHeld(ctx, doctorID, "2026-06-20")
	require.NoError(t, err)
	require.Empty(t, held)

	err = store.Release(ctx, hold.Token, patientID)
	require.True(t, errors.Is(err, domain.ErrHoldExpired))
}

func TestHoldExpiresAndIndexIsPruned(t *testing.T) {
	store, mr, fakeClock := newHoldStore(t)
	ctx := context.Background()
	location := nairobi(t)
	doctorID := uuid.New()

	start := time.Date(2026, 6, 20, 11, 0, 0, 0, location)
	hold, err := store.Create(ctx, cache.HoldInput{
		DoctorID: doctorID, PatientID: uuid.New(), StartUTC: start, Location: location,
	})
	require.NoError(t, err)

	mr.FastForward(5*time.Minute + time.Second)
	fakeClock.Advance(5*time.Minute + time.Second)

	got, err := store.Get(ctx, doctorID, start, location)
	require.NoError(t, err)
	require.Nil(t, got)

	held, err := store.ListHeld(ctx, doctorID, "2026-06-20")
	require.NoError(t, err)
	require.Empty(t, held)
	require.NotEmpty(t, hold.Token)
}
