package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/service"
)

const refreshTokenValue = "test-refresh-token-value"

type fakeAuthStore struct {
	session         *domain.Session
	user            *domain.User
	revoked         []uuid.UUID
	revokedFamilies []uuid.UUID
	rotateCalls     int
	rotatedSession  *domain.NewSession
}

func (f *fakeAuthStore) GetUserByEmail(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeAuthStore) GetActiveUserByID(context.Context, uuid.UUID) (*domain.User, error) {
	return f.user, nil
}

func (f *fakeAuthStore) GetPasswordHash(context.Context, uuid.UUID) (string, error) {
	return "", nil
}

func (f *fakeAuthStore) TouchLastLogin(context.Context, uuid.UUID) error { return nil }

func (f *fakeAuthStore) UpdatePasswordHash(context.Context, uuid.UUID, string) error { return nil }

func (f *fakeAuthStore) CreateSession(context.Context, domain.NewSession) (*domain.Session, error) {
	return nil, nil
}

func (f *fakeAuthStore) GetSessionByHash(_ context.Context, hash string) (*domain.Session, error) {
	if f.session == nil || f.session.RefreshTokenHash != hash {
		return nil, domain.ErrNotFound
	}
	return f.session, nil
}

func (f *fakeAuthStore) RotateSession(_ context.Context, _ uuid.UUID, input domain.NewSession) error {
	f.rotateCalls++
	f.rotatedSession = &input
	return nil
}

func (f *fakeAuthStore) RevokeSession(_ context.Context, id uuid.UUID) error {
	f.revoked = append(f.revoked, id)
	return nil
}

func (f *fakeAuthStore) RevokeSessionFamily(_ context.Context, familyID uuid.UUID) error {
	f.revokedFamilies = append(f.revokedFamilies, familyID)
	return nil
}

func (f *fakeAuthStore) RevokeUserSessions(context.Context, uuid.UUID) error { return nil }

func (f *fakeAuthStore) GetUserSummary(_ context.Context, _ uuid.UUID) (*domain.UserSummary, error) {
	return &domain.UserSummary{User: *f.user}, nil
}

func (f *fakeAuthStore) EnsurePatientIdentity(context.Context, uuid.UUID) error { return nil }

func (f *fakeAuthStore) GetRoleByName(context.Context, string) (*domain.Role, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeAuthStore) CreateUserWithRoles(context.Context, domain.NewUser, []uuid.UUID, uuid.UUID, domain.RequestMeta) (*domain.UserSummary, error) {
	return nil, nil
}

type fakeAccessReader struct{}

func (fakeAccessReader) Access(context.Context, uuid.UUID) (domain.UserAccess, error) {
	return domain.UserAccess{Roles: []string{"patient"}}, nil
}

func newAuthService(t *testing.T, store *fakeAuthStore, now time.Time) *service.AuthService {
	t.Helper()
	tokens := token.New("test-secret-value-that-is-long-enough", "curaclinic-api", "curaclinic-web", 15*time.Minute)
	return service.NewAuthService(store, fakeAccessReader{}, tokens, 10, 24*time.Hour, 30*time.Minute, clock.NewFake(now))
}

func newFakeStore(now time.Time, lastUsed *time.Time) *fakeAuthStore {
	userID := uuid.New()
	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           userID,
		FamilyID:         uuid.New(),
		RefreshTokenHash: token.HashRefreshToken(refreshTokenValue),
		ExpiresAt:        now.Add(24 * time.Hour),
		LastUsedAt:       lastUsed,
		CreatedAt:        now.Add(-2 * time.Hour),
	}
	return &fakeAuthStore{
		session: session,
		user: &domain.User{
			ID:        userID,
			Email:     "patient@curaclinic.local",
			FirstName: "Test",
			LastName:  "Patient",
			Status:    domain.UserStatusActive,
		},
	}
}

func TestRefreshRejectsIdleSession(t *testing.T) {
	now := time.Date(2026, time.January, 15, 9, 0, 0, 0, time.UTC)
	lastUsed := now.Add(-31 * time.Minute)
	store := newFakeStore(now, &lastUsed)
	svc := newAuthService(t, store, now)

	_, err := svc.Refresh(context.Background(), refreshTokenValue, domain.RequestMeta{})

	require.Error(t, err)
	require.True(t, domain.IsCode(err, domain.CodeUnauthorized), err.Error())
	require.Equal(t, []uuid.UUID{store.session.ID}, store.revoked)
	require.Zero(t, store.rotateCalls)
}

func TestRefreshRejectsIdleSessionWithoutLastUsedAt(t *testing.T) {
	now := time.Date(2026, time.January, 15, 9, 0, 0, 0, time.UTC)
	store := newFakeStore(now, nil)
	store.session.CreatedAt = now.Add(-31 * time.Minute)
	svc := newAuthService(t, store, now)

	_, err := svc.Refresh(context.Background(), refreshTokenValue, domain.RequestMeta{})

	require.Error(t, err)
	require.True(t, domain.IsCode(err, domain.CodeUnauthorized), err.Error())
	require.Equal(t, []uuid.UUID{store.session.ID}, store.revoked)
	require.Zero(t, store.rotateCalls)
}

func TestRefreshRotatesActiveSession(t *testing.T) {
	now := time.Date(2026, time.January, 15, 9, 0, 0, 0, time.UTC)
	lastUsed := now.Add(-10 * time.Minute)
	store := newFakeStore(now, &lastUsed)
	svc := newAuthService(t, store, now)

	result, err := svc.Refresh(context.Background(), refreshTokenValue, domain.RequestMeta{})

	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.NotEqual(t, refreshTokenValue, result.RefreshToken)
	require.Equal(t, int64((15 * time.Minute).Seconds()), result.ExpiresIn)
	require.Equal(t, 1, store.rotateCalls)
	require.NotNil(t, store.rotatedSession)
	require.Equal(t, store.session.FamilyID, store.rotatedSession.FamilyID)
	require.Equal(t, now.Add(24*time.Hour), store.rotatedSession.ExpiresAt)
	require.Empty(t, store.revoked)
}

func TestRefreshAllowsSessionAtIdleBoundary(t *testing.T) {
	now := time.Date(2026, time.January, 15, 9, 0, 0, 0, time.UTC)
	lastUsed := now.Add(-30 * time.Minute)
	store := newFakeStore(now, &lastUsed)
	svc := newAuthService(t, store, now)

	_, err := svc.Refresh(context.Background(), refreshTokenValue, domain.RequestMeta{})

	require.NoError(t, err)
	require.Equal(t, 1, store.rotateCalls)
}

func TestRefreshRejectsRevokedSessionAndRevokesFamily(t *testing.T) {
	now := time.Date(2026, time.January, 15, 9, 0, 0, 0, time.UTC)
	lastUsed := now.Add(-time.Minute)
	store := newFakeStore(now, &lastUsed)
	revokedAt := now.Add(-time.Minute)
	store.session.RevokedAt = &revokedAt
	svc := newAuthService(t, store, now)

	_, err := svc.Refresh(context.Background(), refreshTokenValue, domain.RequestMeta{})

	require.Error(t, err)
	require.True(t, domain.IsCode(err, domain.CodeUnauthorized), err.Error())
	require.Equal(t, []uuid.UUID{store.session.FamilyID}, store.revokedFamilies)
	require.Zero(t, store.rotateCalls)
	require.Empty(t, store.revoked)
}
