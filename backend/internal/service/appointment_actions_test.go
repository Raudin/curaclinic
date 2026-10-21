package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/service"
)

func TestCancelEnforcesOwnershipAndStatus(t *testing.T) {
	svc, store, _, _, doctorID, patientID, slot := appointmentFixture(t)
	ctx := context.Background()
	appointment := domain.Appointment{
		ID: uuid.New(), DoctorID: doctorID, PatientID: patientID,
		StartTime: slot.StartTime, EndTime: slot.EndTime, Status: domain.AppointmentStatusConfirmed,
	}
	store.appointments[appointment.ID] = appointment

	_, err := svc.Cancel(ctx, testActor(patientID, domain.PermAppointmentsViewOwn), appointment.ID, "sick", domain.RequestMeta{})
	require.ErrorIs(t, err, domain.ErrForbidden)

	cancelled, err := svc.Cancel(ctx, testActor(patientID, domain.PermAppointmentsCancelOwn), appointment.ID, "sick", domain.RequestMeta{})
	require.NoError(t, err)
	require.Equal(t, domain.AppointmentStatusCancelled, cancelled.Status)

	_, err = svc.Cancel(ctx, testActor(patientID, domain.PermAppointmentsCancelOwn), appointment.ID, "again", domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeConflict))
}

func TestRescheduleAcquiresAndReleasesHold(t *testing.T) {
	svc, store, holds, validator, doctorID, patientID, slot := appointmentFixture(t)
	ctx := context.Background()
	appointment := domain.Appointment{
		ID: uuid.New(), DoctorID: doctorID, PatientID: patientID,
		StartTime: slot.StartTime, EndTime: slot.EndTime, Status: domain.AppointmentStatusConfirmed,
	}
	store.appointments[appointment.ID] = appointment

	newStart := slot.StartTime.Add(2 * time.Hour)
	validator.slot = &domain.BookableSlot{
		DoctorID: doctorID, StartTime: newStart, EndTime: newStart.Add(30 * time.Minute),
		Timezone: "Africa/Nairobi", SlotMinutes: 30,
	}

	updated, err := svc.Reschedule(ctx, testActor(patientID, domain.PermAppointmentsRescheduleOwn), appointment.ID, newStart, "travel", domain.RequestMeta{})
	require.NoError(t, err)
	require.Equal(t, newStart, updated.StartTime)
	require.Len(t, holds.created, 1)
	require.Equal(t, patientID, holds.created[0].PatientID)
	require.Len(t, holds.released, 1)
}

func TestRescheduleRejectsSameTime(t *testing.T) {
	svc, store, _, _, doctorID, patientID, slot := appointmentFixture(t)
	appointment := domain.Appointment{
		ID: uuid.New(), DoctorID: doctorID, PatientID: patientID,
		StartTime: slot.StartTime, EndTime: slot.EndTime, Status: domain.AppointmentStatusConfirmed,
	}
	store.appointments[appointment.ID] = appointment

	_, err := svc.Reschedule(context.Background(), testActor(patientID, domain.PermAppointmentsRescheduleOwn), appointment.ID, slot.StartTime, "", domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeValidation))
}

func TestBookManualValidatesPatientAndSlot(t *testing.T) {
	svc, store, _, validator, doctorID, patientID, slot := appointmentFixture(t)
	store.patient = domain.User{ID: patientID, FirstName: "Amelia", LastName: "Hart", Status: domain.UserStatusInactive}
	_, err := svc.BookManual(context.Background(), testActor(uuid.New(), domain.PermAdminAppointmentsCreate), service.BookManualInput{
		DoctorID: doctorID, PatientID: patientID, StartTime: slot.StartTime, Reason: "follow-up",
	}, domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeValidation))

	store.patient.Status = domain.UserStatusActive
	validator.err = domain.ErrSlotUnavailable
	_, err = svc.BookManual(context.Background(), testActor(uuid.New(), domain.PermAdminAppointmentsCreate), service.BookManualInput{
		DoctorID: doctorID, PatientID: patientID, StartTime: slot.StartTime, Reason: "follow-up",
	}, domain.RequestMeta{})
	require.ErrorIs(t, err, domain.ErrSlotUnavailable)
}

func TestListForcesPatientScopeWithoutViewAny(t *testing.T) {
	svc, store, _, _, _, patientID, _ := appointmentFixture(t)
	actor := testActor(patientID, domain.PermAppointmentsViewOwn)
	actor.Roles = []string{"patient"}
	_, _, err := svc.List(context.Background(), actor, service.ListAppointmentsInput{Scope: "all"})
	require.NoError(t, err)
	require.NotNil(t, store.lastFilter.PatientID)
	require.Equal(t, patientID, *store.lastFilter.PatientID)
	require.Nil(t, store.lastFilter.DoctorID)
}

func TestListScopesDoctorsToTheirOwnDiary(t *testing.T) {
	svc, store, _, _, doctorID, _, _ := appointmentFixture(t)
	actor := testActor(doctorID, domain.PermDoctorAppointmentsViewOwn)
	actor.Roles = []string{"doctor"}
	_, _, err := svc.List(context.Background(), actor, service.ListAppointmentsInput{Scope: "upcoming"})
	require.NoError(t, err)
	require.NotNil(t, store.lastFilter.DoctorID)
	require.Equal(t, doctorID, *store.lastFilter.DoctorID)
	require.Nil(t, store.lastFilter.PatientID)
	require.NotNil(t, store.lastFilter.Status)
	require.Equal(t, domain.AppointmentStatusConfirmed, *store.lastFilter.Status)
	// Upcoming filters on end time so in-progress appointments stay listed.
	require.Nil(t, store.lastFilter.FromTime)
	require.NotNil(t, store.lastFilter.EndsAfter)
}

func TestListDoctorViewRequiresDoctorPermission(t *testing.T) {
	svc, _, _, _, doctorID, _, _ := appointmentFixture(t)
	// A doctor whose doctor.appointments.view_own is denied but who can still
	// view their own patient appointments.
	actor := testActor(doctorID, domain.PermAppointmentsViewOwn)
	actor.Roles = []string{"doctor", "patient"}
	_, _, err := svc.List(context.Background(), actor, service.ListAppointmentsInput{As: "doctor"})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCancelAndRescheduleRejectStartedAppointments(t *testing.T) {
	svc, store, _, _, doctorID, patientID, _ := appointmentFixture(t)
	ctx := context.Background()
	start := time.Date(2026, 10, 19, 6, 0, 0, 0, time.UTC) // before the fixture clock
	appointment := domain.Appointment{
		ID: uuid.New(), DoctorID: doctorID, PatientID: patientID,
		StartTime: start, EndTime: start.Add(30 * time.Minute), Status: domain.AppointmentStatusConfirmed,
	}
	store.appointments[appointment.ID] = appointment

	_, err := svc.Cancel(ctx, testActor(patientID, domain.PermAppointmentsCancelOwn), appointment.ID, "missed", domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeConflict))

	_, err = svc.Reschedule(ctx, testActor(patientID, domain.PermAppointmentsRescheduleOwn), appointment.ID, start.Add(48*time.Hour), "missed", domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeConflict))
}
