package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres"
)

func integrationRepo(t *testing.T) *postgres.Repo {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	t.Cleanup(pool.Close)
	return postgres.New(pool)
}

// createPatientUser creates an authenticated user with the patient identity
// (a patients row), which is what grants the built-in patient permissions.
func createPatientUser(t *testing.T, repo *postgres.Repo) *domain.User {
	t.Helper()
	userID := uuid.New()
	user, err := repo.CreateUserWithRoles(context.Background(), domain.NewUser{
		ID:           userID,
		Email:        fmt.Sprintf("test-patient-%s@curaclinic.test", uuid.NewString()),
		PasswordHash: "not-a-real-hash",
		FirstName:    "Test",
		LastName:     "Patient",
		Status:       domain.UserStatusActive,
	}, nil, userID, domain.RequestMeta{IP: "test"})
	require.NoError(t, err)
	return &user.User
}

// createDoctorUser creates a user, doctor profile and doctor role assignment.
func createDoctorUser(t *testing.T, repo *postgres.Repo) *domain.User {
	t.Helper()
	userID := uuid.New()
	doctor, err := repo.CreateDoctorAccount(context.Background(), domain.NewUser{
		ID:           userID,
		Email:        fmt.Sprintf("test-doctor-%s@curaclinic.test", uuid.NewString()),
		PasswordHash: "not-a-real-hash",
		FirstName:    "Test",
		LastName:     "Doctor",
		Status:       domain.UserStatusActive,
	}, "doctor", domain.NewDoctorProfile{
		UserID:              userID,
		Specialty:           "Family Medicine",
		Timezone:            "Africa/Nairobi",
		SlotMinutes:         30,
		IsAcceptingPatients: true,
	}, userID, domain.RequestMeta{IP: "test"})
	require.NoError(t, err)
	return &domain.User{
		ID:        doctor.UserID,
		Email:     doctor.Email,
		FirstName: doctor.FirstName,
		LastName:  doctor.LastName,
		Status:    domain.UserStatusActive,
	}
}

func TestEffectivePermissionsDenyWinsIntegration(t *testing.T) {
	repo := integrationRepo(t)
	ctx := context.Background()

	user := createDoctorUser(t, repo)
	permission, err := repo.GetPermissionByCode(ctx, domain.PermDoctorAppointmentsViewOwn)
	require.NoError(t, err)

	baseline, err := repo.EffectivePermissions(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, baseline, domain.PermDoctorAppointmentsViewOwn)

	_, err = repo.SetOverride(ctx, user.ID, permission.ID, "deny", "administrative hold", user.ID, domain.RequestMeta{})
	require.NoError(t, err)

	granted, err := repo.EffectivePermissions(ctx, user.ID)
	require.NoError(t, err)
	require.NotContains(t, granted, domain.PermDoctorAppointmentsViewOwn)

	reports, err := repo.GetPermissionByCode(ctx, domain.PermReportsView)
	require.NoError(t, err)
	_, err = repo.SetOverride(ctx, user.ID, reports.ID, "allow", "temporary reporting access", user.ID, domain.RequestMeta{})
	require.NoError(t, err)

	granted, err = repo.EffectivePermissions(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, granted, domain.PermReportsView)
	require.NotContains(t, granted, domain.PermDoctorAppointmentsViewOwn, "deny must win over the role grant")

	require.NoError(t, repo.DeleteOverride(ctx, user.ID, permission.ID, user.ID, domain.RequestMeta{}))
	granted, err = repo.EffectivePermissions(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, granted, domain.PermDoctorAppointmentsViewOwn)
}

func TestAppointmentOverlapConstraintIntegration(t *testing.T) {
	repo := integrationRepo(t)
	ctx := context.Background()

	doctor := createDoctorUser(t, repo)
	patientOne := createPatientUser(t, repo)
	patientTwo := createPatientUser(t, repo)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Minute)
	end := start.Add(30 * time.Minute)

	first, err := repo.CreateAppointment(ctx, doctor.ID, patientOne.ID, start, end, "initial visit", patientOne.ID, domain.RequestMeta{})
	require.NoError(t, err)
	require.Equal(t, domain.AppointmentStatusConfirmed, first.Status)

	_, err = repo.CreateAppointment(ctx, doctor.ID, patientTwo.ID, start, end, "double booking", patientTwo.ID, domain.RequestMeta{})
	require.ErrorIs(t, err, domain.ErrSlotUnavailable, "the database must reject overlapping doctor bookings")

	_, err = repo.CancelAppointment(ctx, *first, "patient cancelled", patientOne.ID, domain.RequestMeta{})
	require.NoError(t, err)

	_, err = repo.RescheduleAppointment(ctx, *first, start.Add(time.Hour), end.Add(time.Hour), "move", patientOne.ID, domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeConflict), "a cancelled appointment cannot be rescheduled")

	second, err := repo.CreateAppointment(ctx, doctor.ID, patientTwo.ID, start, end, "taken after cancellation", patientTwo.ID, domain.RequestMeta{})
	require.NoError(t, err, "a cancelled appointment must free the slot")
	require.Equal(t, patientTwo.ID, second.PatientID)

	events, err := repo.AppointmentEvents(ctx, first.ID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, domain.EventCreated, events[0].EventType)
	require.Equal(t, domain.EventCancelled, events[1].EventType)
}

func TestSessionRotationIntegration(t *testing.T) {
	repo := integrationRepo(t)
	ctx := context.Background()

	user := createPatientUser(t, repo)
	familyID := uuid.New()
	oldHash := "old-" + uuid.NewString()
	session, err := repo.CreateSession(ctx, domain.NewSession{
		ID:               uuid.New(),
		UserID:           user.ID,
		FamilyID:         familyID,
		RefreshTokenHash: oldHash,
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	newHash := "new-" + uuid.NewString()
	newSessionID := uuid.New()
	require.NoError(t, repo.RotateSession(ctx, session.ID, domain.NewSession{
		ID:               newSessionID,
		UserID:           user.ID,
		FamilyID:         familyID,
		RefreshTokenHash: newHash,
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	}))

	oldSession, err := repo.GetSessionByHash(ctx, oldHash)
	require.NoError(t, err)
	require.NotNil(t, oldSession.RevokedAt)

	replacement, err := repo.GetSessionByHash(ctx, newHash)
	require.NoError(t, err)
	require.Equal(t, newSessionID, replacement.ID)
	require.Nil(t, replacement.RevokedAt)

	err = repo.RotateSession(ctx, session.ID, domain.NewSession{
		ID:               uuid.New(),
		UserID:           user.ID,
		FamilyID:         familyID,
		RefreshTokenHash: "reuse-" + uuid.NewString(),
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	require.True(t, domain.IsCode(err, domain.CodeUnauthorized), "reusing a rotated token must fail")
}

func TestRoleReassignmentAfterSoftDeleteIntegration(t *testing.T) {
	repo := integrationRepo(t)
	ctx := context.Background()

	user := createDoctorUser(t, repo)
	require.NoError(t, repo.AssignUserRoles(ctx, user.ID, []uuid.UUID{}, user.ID, domain.RequestMeta{}))

	permission, err := repo.GetPermissionByCode(ctx, domain.PermDoctorScheduleView)
	require.NoError(t, err)
	role, err := repo.CreateRole(ctx, "doctor", "locum-"+uuid.NewString(), "Locum doctor role",
		[]uuid.UUID{permission.ID}, user.ID, domain.RequestMeta{})
	require.NoError(t, err)

	require.NoError(t, repo.AssignUserRoles(ctx, user.ID, []uuid.UUID{role.ID}, user.ID, domain.RequestMeta{}))
	roles, err := repo.RoleNames(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, roles, role.Name)
}
