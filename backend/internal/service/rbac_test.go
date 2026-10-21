package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/service"
)

func newPermissionCache(t *testing.T) (*cache.PermissionCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return cache.NewPermissionCache(client, time.Minute), mr
}

func TestRBAccessCachesUntilInvalidated(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	store := &fakeRBACStore{
		permissions: []string{domain.PermAppointmentsCreateSelf},
		roles:       []string{"patient"},
	}
	svc := service.NewRBACService(store, permissionCache)
	userID := uuid.New()

	first, err := svc.Access(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, []string{"patient"}, first.Roles)
	require.Equal(t, 1, store.effectiveCalls)

	second, err := svc.Access(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, 1, store.effectiveCalls, "second read must come from cache")

	require.NoError(t, svc.Invalidate(context.Background()))
	_, err = svc.Access(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, 2, store.effectiveCalls, "invalidation must bypass the old cache entry")
}

func TestRBACAccessDoesNotCacheSnapshotLoadedBeforeInvalidation(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	store := &fakeRBACStore{roles: []string{"admin"}}
	svc := service.NewRBACService(store, permissionCache)
	userID := uuid.New()

	// An admin change lands while the first request is reading the database.
	store.onLoad = func() {
		store.onLoad = nil
		require.NoError(t, svc.Invalidate(context.Background()))
	}
	_, err := svc.Access(context.Background(), userID)
	require.NoError(t, err)

	_, err = svc.Access(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, 2, store.effectiveCalls, "a snapshot read before invalidation must not be cached under the new version")
}

func TestAssignRolesProtectsLastAdmin(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	adminRoleID := uuid.New()
	patientRoleID := uuid.New()
	userID := uuid.New()
	store := &fakeRBACStore{
		rolesByUser:  map[uuid.UUID][]string{userID: {"admin"}},
		adminHolders: map[uuid.UUID]bool{userID: true},
		roleByID:     map[uuid.UUID]*domain.Role{patientRoleID: {ID: patientRoleID, Name: "patient"}},
		activeAdmins: 1,
	}
	svc := service.NewRBACService(store, permissionCache)

	err := svc.AssignRoles(context.Background(), uuid.New(), userID, []uuid.UUID{patientRoleID}, domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeConflict))
	require.Nil(t, store.assignedRoleIDs)

	store.activeAdmins = 2
	err = svc.AssignRoles(context.Background(), uuid.New(), userID, []uuid.UUID{patientRoleID}, domain.RequestMeta{})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{patientRoleID}, store.assignedRoleIDs)
	require.NotEqual(t, adminRoleID, store.assignedRoleIDs[0])
}

func TestAssignRolesIgnoresAdminIdentityWithoutAdminRole(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	receptionistRoleID := uuid.New()
	userID := uuid.New()
	// The user has an admin identity (so RoleNames reports "admin") but only a
	// custom admin-scope role, so changing their roles cannot remove the last
	// holder of the system admin role.
	store := &fakeRBACStore{
		rolesByUser:  map[uuid.UUID][]string{userID: {"admin", "receptionist"}},
		roleByID:     map[uuid.UUID]*domain.Role{receptionistRoleID: {ID: receptionistRoleID, Name: "receptionist"}},
		activeAdmins: 1,
	}
	svc := service.NewRBACService(store, permissionCache)

	err := svc.AssignRoles(context.Background(), uuid.New(), userID, []uuid.UUID{receptionistRoleID}, domain.RequestMeta{})
	require.NoError(t, err)
}

func TestAssignRolesValidatesRoleIDs(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	store := &fakeRBACStore{roleByID: map[uuid.UUID]*domain.Role{}}
	svc := service.NewRBACService(store, permissionCache)

	err := svc.AssignRoles(context.Background(), uuid.New(), uuid.New(), []uuid.UUID{uuid.New()}, domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeValidation))
}

func TestSetOverrideValidatesEffect(t *testing.T) {
	permissionCache, _ := newPermissionCache(t)
	store := &fakeRBACStore{}
	svc := service.NewRBACService(store, permissionCache)

	_, err := svc.SetOverride(context.Background(), uuid.New(), uuid.New(), uuid.New(), "maybe", "", domain.RequestMeta{})
	require.True(t, domain.IsCode(err, domain.CodeValidation))
}
