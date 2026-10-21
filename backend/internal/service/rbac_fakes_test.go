package service_test

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

type fakeRBACStore struct {
	user            domain.User
	permissions     []string
	roles           []string
	rolesByUser     map[uuid.UUID][]string
	roleByID        map[uuid.UUID]*domain.Role
	permissionsByID map[uuid.UUID]*domain.Permission
	activeAdmins    int64
	adminHolders    map[uuid.UUID]bool
	onLoad          func()
	effectiveCalls  int
	assignedRoleIDs []uuid.UUID
	overrides       []domain.PermissionOverride
}

func (f *fakeRBACStore) GetUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	user := f.user
	if user.ID == uuid.Nil {
		user.ID = id
	}
	return &user, nil
}

func (f *fakeRBACStore) EffectivePermissions(_ context.Context, _ uuid.UUID) ([]string, error) {
	f.effectiveCalls++
	if f.onLoad != nil {
		f.onLoad()
	}
	return f.permissions, nil
}

func (f *fakeRBACStore) RoleNames(_ context.Context, userID uuid.UUID) ([]string, error) {
	if f.rolesByUser != nil {
		if roles, ok := f.rolesByUser[userID]; ok {
			return roles, nil
		}
	}
	return f.roles, nil
}

func (f *fakeRBACStore) UserRoleIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeRBACStore) ListRoles(_ context.Context) ([]domain.Role, error) { return nil, nil }

func (f *fakeRBACStore) GetRoleByID(_ context.Context, id uuid.UUID) (*domain.Role, error) {
	if role, ok := f.roleByID[id]; ok {
		return role, nil
	}
	return nil, domain.ErrNotFound
}

func (f *fakeRBACStore) ListPermissions(_ context.Context) ([]domain.Permission, error) {
	return nil, nil
}

func (f *fakeRBACStore) GetPermissionByID(_ context.Context, id uuid.UUID) (*domain.Permission, error) {
	if permission, ok := f.permissionsByID[id]; ok {
		return permission, nil
	}
	return nil, domain.ErrNotFound
}

func (f *fakeRBACStore) ListOverrides(_ context.Context, _ uuid.UUID) ([]domain.PermissionOverride, error) {
	return f.overrides, nil
}

func (f *fakeRBACStore) ActiveAdminCount(_ context.Context) (int64, error) {
	return f.activeAdmins, nil
}

func (f *fakeRBACStore) HoldsAdminRole(_ context.Context, userID uuid.UUID) (bool, error) {
	return f.adminHolders[userID], nil
}

func (f *fakeRBACStore) CreateRole(_ context.Context, _ string, _ string, _ string, _ []uuid.UUID, _ uuid.UUID, _ domain.RequestMeta) (*domain.Role, error) {
	return nil, nil
}

func (f *fakeRBACStore) UpdateRole(_ context.Context, _ uuid.UUID, _ string, _ string, _ []uuid.UUID, _ uuid.UUID, _ domain.RequestMeta) (*domain.Role, error) {
	return nil, nil
}

func (f *fakeRBACStore) DeleteRole(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ domain.RequestMeta) error {
	return nil
}

func (f *fakeRBACStore) AssignUserRoles(_ context.Context, _ uuid.UUID, roleIDs []uuid.UUID, _ uuid.UUID, _ domain.RequestMeta) error {
	f.assignedRoleIDs = roleIDs
	return nil
}

func (f *fakeRBACStore) SetOverride(_ context.Context, userID, permissionID uuid.UUID, effect, note string, _ uuid.UUID, _ domain.RequestMeta) (*domain.PermissionOverride, error) {
	override := domain.PermissionOverride{
		ID: uuid.New(), UserID: userID, PermissionID: permissionID,
		PermissionCode: "appointments.create_self", Effect: effect, Note: note,
	}
	f.overrides = append(f.overrides, override)
	return &override, nil
}

func (f *fakeRBACStore) DeleteOverride(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID, _ domain.RequestMeta) error {
	return nil
}
