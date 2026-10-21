package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// AssignRoles replaces a user's roles, guarding the last active admin.
func (s *RBACService) AssignRoles(ctx context.Context, actor, userID uuid.UUID, roleIDs []uuid.UUID, meta domain.RequestMeta) error {
	if _, err := s.store.GetUserByID(ctx, userID); err != nil {
		return err
	}

	roleIDs = uniqueIDs(roleIDs)
	keepsAdmin := false
	for _, roleID := range roleIDs {
		role, err := s.store.GetRoleByID(ctx, roleID)
		if err != nil {
			if domain.IsCode(err, domain.CodeNotFound) {
				return domain.Validation("one of the selected roles does not exist", nil)
			}
			return err
		}
		if role.IsSystem && role.Name == "admin" {
			keepsAdmin = true
		}
	}

	// The repository re-checks this atomically; checking here first gives a
	// clear error without opening a transaction.
	holdsAdmin, err := s.store.HoldsAdminRole(ctx, userID)
	if err != nil {
		return err
	}
	if holdsAdmin && !keepsAdmin {
		count, err := s.store.ActiveAdminCount(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.NewError(domain.CodeConflict, "cannot remove the last active admin")
		}
	}

	if err := s.store.AssignUserRoles(ctx, userID, roleIDs, actor, meta); err != nil {
		return err
	}
	if err := s.Invalidate(ctx); err != nil {
		return fmt.Errorf("invalidate permission cache: %w", err)
	}
	return nil
}

// SetOverride grants or denies one permission for a user.
func (s *RBACService) SetOverride(ctx context.Context, actor, userID, permissionID uuid.UUID, effect, note string, meta domain.RequestMeta) (*domain.PermissionOverride, error) {
	if effect != string(domain.OverrideAllow) && effect != string(domain.OverrideDeny) {
		return nil, domain.ValidationField("effect", "must be allow or deny")
	}
	if _, err := s.store.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetPermissionByID(ctx, permissionID); err != nil {
		return nil, err
	}
	override, err := s.store.SetOverride(ctx, userID, permissionID, effect, note, actor, meta)
	if err != nil {
		return nil, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return nil, fmt.Errorf("invalidate permission cache: %w", err)
	}
	return override, nil
}

// DeleteOverride removes an explicit grant or deny.
func (s *RBACService) DeleteOverride(ctx context.Context, actor, userID, permissionID uuid.UUID, meta domain.RequestMeta) error {
	if _, err := s.store.GetUserByID(ctx, userID); err != nil {
		return err
	}
	if err := s.store.DeleteOverride(ctx, userID, permissionID, actor, meta); err != nil {
		return err
	}
	if err := s.Invalidate(ctx); err != nil {
		return fmt.Errorf("invalidate permission cache: %w", err)
	}
	return nil
}

// CreateRole creates a custom role.
func (s *RBACService) CreateRole(ctx context.Context, actor uuid.UUID, scope, name, description string, permissionIDs []uuid.UUID, meta domain.RequestMeta) (*domain.Role, error) {
	if scope != "doctor" && scope != "admin" {
		return nil, domain.ValidationField("scope", "must be doctor or admin")
	}
	if err := s.validatePermissions(ctx, scope, permissionIDs); err != nil {
		return nil, err
	}
	role, err := s.store.CreateRole(ctx, scope, name, description, uniqueIDs(permissionIDs), actor, meta)
	if err != nil {
		return nil, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole updates a role and its permission set.
func (s *RBACService) UpdateRole(ctx context.Context, actor, roleID uuid.UUID, name, description string, permissionIDs []uuid.UUID, meta domain.RequestMeta) (*domain.Role, error) {
	existing, err := s.store.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if existing.IsSystem && existing.Name != name {
		return nil, domain.Validation("system roles cannot be renamed", nil)
	}
	permissionIDs = uniqueIDs(permissionIDs)
	if validateErr := s.validatePermissions(ctx, existing.Scope, permissionIDs); validateErr != nil {
		return nil, validateErr
	}
	role, err := s.store.UpdateRole(ctx, roleID, name, description, permissionIDs, actor, meta)
	if err != nil {
		return nil, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return nil, err
	}
	return role, nil
}

// DeleteRole removes a non-system role.
func (s *RBACService) DeleteRole(ctx context.Context, actor, roleID uuid.UUID, meta domain.RequestMeta) error {
	if err := s.store.DeleteRole(ctx, roleID, actor, meta); err != nil {
		return err
	}
	return s.Invalidate(ctx)
}

func (s *RBACService) validatePermissions(ctx context.Context, scope string, permissionIDs []uuid.UUID) error {
	if len(permissionIDs) > 200 {
		return domain.Validation("too many permissions selected", nil)
	}
	for _, permissionID := range permissionIDs {
		permission, err := s.store.GetPermissionByID(ctx, permissionID)
		if err != nil {
			if domain.IsCode(err, domain.CodeNotFound) {
				return domain.Validation("one of the selected permissions does not exist", nil)
			}
			return err
		}
		if !strings.HasPrefix(permission.Code, scope+".") {
			return domain.Validation("role permissions must match the role scope", nil)
		}
	}
	return nil
}
