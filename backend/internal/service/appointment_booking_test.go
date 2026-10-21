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

func testActor(id uuid.UUID, permissions ...string) *domain.Actor {
	set := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		set[permission] = struct{}{}
	}
	return &domain.Actor{UserID: id, Permissions: set}
}

func appointmentFixture(t *testing.T) (*service.AppointmentService, *fakeAppointmentStore, *fakeSlotHoldStore, *fakeSlotValidator, uuid.UUID, uuid.UUID, *domain.BookableSlot) {
	t.Helper()
	store := newFakeAppointmentStore()
	doctorID := uuid.New()
	patientID := uuid.New()
	start := time.Date(2026, 10, 20, 6, 0, 0, 0, time.UTC)
	slot := &domain.BookableSlot{
		DoctorID:    doctorID,
		StartTime:   start,
		EndTime:     start.Add(30 * time.Minute),
		Timezone:    "Africa/Nairobi",
		SlotMinutes: 30,
	}
	validator := &fakeSlotValidator{slot: slot}
	holds := &fakeSlotHoldStore{}
	svc := service.NewAppointmentService(store, validator, holds, clock.NewFake(time.Date(2026, 10, 20, 4, 0, 0, 0, time.UTC)))
	return svc, store, holds, validator, doctorID, patientID, slot
}

func TestBookForSelfRequiresOwnedHold(t *testing.T) {
	svc, store, holds, _, doctorID, patientID, slot := appointmentFixture(t)
	ctx := context.Background()
	actor := testActor(patientID, domain.PermAppointmentsCreateSelf)
	meta := domain.RequestMeta{RequestID: "req-1", IP: "127.0.0.1"}

	_, err := svc.BookForSelf(ctx, actor, service.BookForSelfInput{
		DoctorID: doctorID, StartTime: slot.StartTime, Reason: "checkup", HoldToken: "token",
	}, meta)
	require.ErrorIs(t, err, domain.ErrHoldExpired)

	holds.existing = &cache.Hold{Token: "token", DoctorID: doctorID, PatientID: uuid.New(), StartUTC: slot.StartTime}
	_, err = svc.BookForSelf(ctx, actor, service.BookForSelfInput{
		DoctorID: doctorID, StartTime: slot.StartTime, Reason: "checkup", HoldToken: "token",
	}, meta)
	require.ErrorIs(t, err, domain.ErrHoldNotOwned)

	holds.existing = &cache.Hold{Token: "other-token", DoctorID: doctorID, PatientID: patientID, StartUTC: slot.StartTime}
	_, err = svc.BookForSelf(ctx, actor, service.BookForSelfInput{
		DoctorID: doctorID, StartTime: slot.StartTime, Reason: "checkup", HoldToken: "token",
	}, meta)
	require.ErrorIs(t, err, domain.ErrHoldNotOwned)

	holds.existing = &cache.Hold{Token: "token", DoctorID: doctorID, PatientID: patientID, StartUTC: slot.StartTime}
	appointment, err := svc.BookForSelf(ctx, actor, service.BookForSelfInput{
		DoctorID: doctorID, StartTime: slot.StartTime, Reason: "checkup", HoldToken: "token",
	}, meta)
	require.NoError(t, err)
	require.Equal(t, patientID, appointment.PatientID)
	require.Len(t, store.created, 1)
	require.Equal(t, []string{"token"}, holds.released)
}

func TestBookForSelfRejectsInvalidSlot(t *testing.T) {
	svc, _, _, validator, doctorID, patientID, slot := appointmentFixture(t)
	validator.err = domain.ErrSlotUnavailable
	_, err := svc.BookForSelf(context.Background(), testActor(patientID, domain.PermAppointmentsCreateSelf), service.BookForSelfInput{
		DoctorID: doctorID, StartTime: slot.StartTime, Reason: "checkup", HoldToken: "token",
	}, domain.RequestMeta{})
	require.ErrorIs(t, err, domain.ErrSlotUnavailable)
}
