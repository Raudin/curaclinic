package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// EffectivePermissions resolves role grants plus explicit overrides using
// "deny wins, then allow, then role".
func (r *Repo) EffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.ListEffectivePermissionCodes(ctx, userID)
}

// RoleNames returns the names of the roles assigned to a user.
func (r *Repo) RoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.ListRoleNamesForUser(ctx, userID)
}

// UserRoleIDs returns the role ids currently assigned to a user.
func (r *Repo) UserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ListRoleIDsForUser(ctx, userID)
}

// ListRoles returns every role with its permission codes.
func (r *Repo) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]domain.Role, 0, len(rows))
	for _, row := range rows {
		codes, err := r.q.ListRolePermissionCodes(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		scope, err := r.q.RoleScope(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, domain.Role{
			ID:          row.ID,
			Name:        row.Name,
			Description: stringValue(row.Description),
			Scope:       scope,
			IsSystem:    row.IsSystem,
			Permissions: codes,
			CreatedAt:   row.CreatedAt,
		})
	}
	return roles, nil
}

// GetRoleByID returns one role.
func (r *Repo) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	row, err := r.q.GetRoleByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	codes, err := r.q.ListRolePermissionCodes(ctx, id)
	if err != nil {
		return nil, err
	}
	scope, err := r.q.RoleScope(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Role{
		ID:          row.ID,
		Name:        row.Name,
		Description: stringValue(row.Description),
		Scope:       scope,
		IsSystem:    row.IsSystem,
		Permissions: codes,
		CreatedAt:   row.CreatedAt,
	}, nil
}

// GetRoleByName returns one role by name.
func (r *Repo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	row, err := r.q.GetRoleByName(ctx, name)
	if err != nil {
		return nil, mapNotFound(err)
	}
	codes, err := r.q.ListRolePermissionCodes(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	scope, err := r.q.RoleScope(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return &domain.Role{
		ID:          row.ID,
		Name:        row.Name,
		Description: stringValue(row.Description),
		Scope:       scope,
		IsSystem:    row.IsSystem,
		Permissions: codes,
		CreatedAt:   row.CreatedAt,
	}, nil
}

// ListPermissions returns the permission catalog.
func (r *Repo) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	rows, err := r.q.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Permission{
			ID:          row.ID,
			Scope:       permissionScope(row.Code),
			Resource:    row.Resource,
			Action:      row.Action,
			Code:        row.Code,
			Description: stringValue(row.Description),
		})
	}
	return items, nil
}

// GetPermissionByID finds one permission.
func (r *Repo) GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	row, err := r.q.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &domain.Permission{
		ID:          row.ID,
		Scope:       permissionScope(row.Code),
		Resource:    row.Resource,
		Action:      row.Action,
		Code:        row.Code,
		Description: stringValue(row.Description),
	}, nil
}

// GetPermissionByCode finds one permission by its stable code.
func (r *Repo) GetPermissionByCode(ctx context.Context, code string) (*domain.Permission, error) {
	row, err := r.q.GetPermissionByCode(ctx, code)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &domain.Permission{
		ID:          row.ID,
		Scope:       permissionScope(row.Code),
		Resource:    row.Resource,
		Action:      row.Action,
		Code:        row.Code,
		Description: stringValue(row.Description),
	}, nil
}

func permissionScope(code string) string {
	if strings.HasPrefix(code, "doctor.") {
		return "doctor"
	}
	if strings.HasPrefix(code, "admin.") {
		return "admin"
	}
	return ""
}

// ListOverrides returns explicit permission overrides for a user.
func (r *Repo) ListOverrides(ctx context.Context, userID uuid.UUID) ([]domain.PermissionOverride, error) {
	rows, err := r.q.ListOverridesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.PermissionOverride, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PermissionOverride{
			ID:             row.ID,
			UserID:         row.UserID,
			PermissionID:   row.PermissionID,
			PermissionCode: row.PermissionCode,
			Effect:         string(row.Effect),
			Note:           stringValue(row.Note),
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return items, nil
}

// ActiveAdminCount returns how many active users hold the system admin role.
func (r *Repo) ActiveAdminCount(ctx context.Context) (int64, error) {
	return r.q.CountActiveAdmins(ctx)
}

// HoldsAdminRole reports whether a user is assigned the system admin role.
func (r *Repo) HoldsAdminRole(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.q.UserHoldsAdminRole(ctx, userID)
}
