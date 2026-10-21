package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
)

type fakeSlotValidator struct {
	slot *domain.BookableSlot
	err  error
}

func (f *fakeSlotValidator) ValidateSlot(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.BookableSlot, error) {
	return f.slot, f.err
}

type fakeSlotHoldStore struct {
	existing  *cache.Hold
	created   []cache.HoldInput
	released  []string
	createErr error
}

func (f *fakeSlotHoldStore) Get(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Location) (*cache.Hold, error) {
	return f.existing, nil
}

func (f *fakeSlotHoldStore) Create(_ context.Context, input cache.HoldInput) (*cache.Hold, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = append(f.created, input)
	return &cache.Hold{
		Token:     uuid.NewString(),
		DoctorID:  input.DoctorID,
		PatientID: input.PatientID,
		StartUTC:  input.StartUTC,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, nil
}

func (f *fakeSlotHoldStore) Release(_ context.Context, token string, _ uuid.UUID) error {
	f.released = append(f.released, token)
	return nil
}

func (f *fakeSlotHoldStore) ListHeld(_ context.Context, _ uuid.UUID, _ string) (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

type fakeAppointmentStore struct {
	appointments map[uuid.UUID]domain.Appointment
	patient      domain.User
	created      []domain.Appointment
	lastFilter   domain.AppointmentFilter
}

func newFakeAppointmentStore() *fakeAppointmentStore {
	return &fakeAppointmentStore{appointments: map[uuid.UUID]domain.Appointment{}}
}

func (f *fakeAppointmentStore) GetUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	user := f.patient
	if user.ID == uuid.Nil {
		user = domain.User{ID: id, FirstName: "Patient", LastName: "One", Status: domain.UserStatusActive}
	}
	return &user, nil
}

func (f *fakeAppointmentStore) GetAppointment(_ context.Context, id uuid.UUID) (*domain.Appointment, error) {
	appointment, ok := f.appointments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &appointment, nil
}

func (f *fakeAppointmentStore) ListAppointments(_ context.Context, filter domain.AppointmentFilter) ([]domain.Appointment, int64, error) {
	f.lastFilter = filter
	return nil, 0, nil
}

func (f *fakeAppointmentStore) CreateAppointment(_ context.Context, doctorID, patientID uuid.UUID, start, end time.Time, reason string, actor uuid.UUID, _ domain.RequestMeta) (*domain.Appointment, error) {
	appointment := domain.Appointment{
		ID:        uuid.New(),
		DoctorID:  doctorID,
		PatientID: patientID,
		StartTime: start,
		EndTime:   end,
		Reason:    reason,
		Status:    domain.AppointmentStatusConfirmed,
		CreatedBy: &actor,
		Doctor:    domain.DoctorSummary{PersonSummary: domain.PersonSummary{ID: doctorID, FirstName: "Naomi", LastName: "Kamau"}, Specialty: "Family Medicine"},
		Patient:   domain.PersonSummary{ID: patientID, FirstName: "Amelia", LastName: "Hart"},
	}
	f.appointments[appointment.ID] = appointment
	f.created = append(f.created, appointment)
	return &appointment, nil
}

func (f *fakeAppointmentStore) CancelAppointment(_ context.Context, appointment domain.Appointment, _ string, _ uuid.UUID, _ domain.RequestMeta) (*domain.Appointment, error) {
	now := time.Now().UTC()
	appointment.Status = domain.AppointmentStatusCancelled
	appointment.CancelledAt = &now
	f.appointments[appointment.ID] = appointment
	return &appointment, nil
}

func (f *fakeAppointmentStore) RescheduleAppointment(_ context.Context, appointment domain.Appointment, start, end time.Time, _ string, _ uuid.UUID, _ domain.RequestMeta) (*domain.Appointment, error) {
	appointment.StartTime = start
	appointment.EndTime = end
	f.appointments[appointment.ID] = appointment
	return &appointment, nil
}

func (f *fakeAppointmentStore) AppointmentEvents(_ context.Context, _ uuid.UUID) ([]domain.AppointmentEvent, error) {
	return nil, nil
}
