// Package service contains the domain use cases behind the HTTP handlers.
package service

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
)

// RBACService resolves and administers roles and permissions.
type RBACService struct {
	store RBACStore
	cache *cache.PermissionCache
}

// NewRBACService creates the authorization service.
func NewRBACService(store RBACStore, permissionCache *cache.PermissionCache) *RBACService {
	return &RBACService{store: store, cache: permissionCache}
}

// Access returns the effective roles and permissions for a user, using the
// versioned Redis cache when available.
func (s *RBACService) Access(ctx context.Context, userID uuid.UUID) (domain.UserAccess, error) {
	access, version, ok := s.cache.Get(ctx, userID)
	if ok {
		return access, nil
	}
	permissions, err := s.store.EffectivePermissions(ctx, userID)
	if err != nil {
		return domain.UserAccess{}, err
	}
	roles, err := s.store.RoleNames(ctx, userID)
	if err != nil {
		return domain.UserAccess{}, err
	}
	if containsString(roles, "patient") {
		permissions = mergePermissions(permissions, domain.PatientPermissionCodes)
	}
	access = domain.UserAccess{Roles: roles, Permissions: permissions}
	s.cache.Set(ctx, version, userID, access)
	return access, nil
}

// AccessDetail returns effective access plus explicit overrides.
func (s *RBACService) AccessDetail(ctx context.Context, userID uuid.UUID) (domain.UserAccess, []domain.PermissionOverride, error) {
	if _, err := s.store.GetUserByID(ctx, userID); err != nil {
		return domain.UserAccess{}, nil, err
	}
	access, err := s.Access(ctx, userID)
	if err != nil {
		return domain.UserAccess{}, nil, err
	}
	overrides, err := s.store.ListOverrides(ctx, userID)
	if err != nil {
		return domain.UserAccess{}, nil, err
	}
	return access, overrides, nil
}

// Invalidate clears every cached access snapshot.
func (s *RBACService) Invalidate(ctx context.Context) error {
	return s.cache.Invalidate(ctx)
}

// ListRoles returns the role catalog.
func (s *RBACService) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.store.ListRoles(ctx)
}

// ListPermissions returns the permission catalog.
func (s *RBACService) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.store.ListPermissions(ctx)
}
