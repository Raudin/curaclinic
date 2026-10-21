package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// CreateRole inserts a scoped role with its permission set.
func (r *Repo) CreateRole(ctx context.Context, scope, name, description string, permissionIDs []uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) (*domain.Role, error) {
	var roleID uuid.UUID
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		id, err := q.CreateScopedRole(ctx, dbgen.CreateScopedRoleParams{
			Scope:       dbgen.PermissionScope(scope),
			Name:        name,
			Description: optionalString(description),
		})
		if err != nil {
			return mapUniqueConflict(err, "a role with this name already exists in that scope")
		}
		roleID = id
		if err := q.SetRolePermissions(ctx, dbgen.SetRolePermissionsParams{
			RoleID:        id,
			PermissionIds: permissionIDs,
		}); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditRoleCreated,
			ResourceType: domain.ResourceRole,
			ResourceID:   id.String(),
			Metadata:     map[string]any{"scope": scope, "name": name, "permission_ids": permissionIDs},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return r.GetRoleByID(ctx, roleID)
}

// UpdateRole updates a role and replaces its scoped permissions.
func (r *Repo) UpdateRole(ctx context.Context, id uuid.UUID, name, description string, permissionIDs []uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) (*domain.Role, error) {
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		affected, err := q.UpdateRoleScopePermissions(ctx, dbgen.UpdateRoleScopePermissionsParams{
			ID:          id,
			Name:        name,
			Description: optionalString(description),
		})
		if err != nil {
			return mapNotFound(err)
		}
		if affected == 0 {
			return domain.ErrNotFound
		}
		if err := q.SoftDeleteRolePermissions(ctx, id); err != nil {
			return err
		}
		if err := q.SetRolePermissions(ctx, dbgen.SetRolePermissionsParams{
			RoleID:        id,
			PermissionIds: permissionIDs,
		}); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditRoleUpdated,
			ResourceType: domain.ResourceRole,
			ResourceID:   id.String(),
			Metadata:     map[string]any{"name": name, "permission_ids": permissionIDs},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return r.GetRoleByID(ctx, id)
}

// DeleteRole soft-deletes a non-system role.
func (r *Repo) DeleteRole(ctx context.Context, id uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		role, err := q.GetRoleByID(ctx, id)
		if err != nil {
			return mapNotFound(err)
		}
		if role.IsSystem {
			return domain.Validation("system roles cannot be deleted", nil)
		}
		if err := q.SoftDeleteRole(ctx, id); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditRoleDeleted,
			ResourceType: domain.ResourceRole,
			ResourceID:   id.String(),
			Metadata:     map[string]any{"name": role.Name},
			Meta:         meta,
		})
	})
}
