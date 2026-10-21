package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// AssignUserRoles replaces the complete role set for a user.
func (r *Repo) AssignUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID, meta domain.RequestMeta) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		if err := guardLastAdmin(ctx, q, "cannot remove the last active admin", func() error {
			return assignScopedRoles(ctx, q, userID, roleIDs, assignedBy)
		}); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &assignedBy,
			Action:       domain.AuditRolesAssigned,
			ResourceType: domain.ResourceUser,
			ResourceID:   userID.String(),
			Metadata:     map[string]any{"role_ids": roleIDs},
			Meta:         meta,
		})
	})
}

// SetOverride creates or updates a user permission override.
func (r *Repo) SetOverride(ctx context.Context, userID, permissionID uuid.UUID, effect, note string, grantedBy uuid.UUID, meta domain.RequestMeta) (*domain.PermissionOverride, error) {
	var override domain.PermissionOverride
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		id, err := setScopedPermissionOverride(ctx, q, userID, permissionID, effect, note, grantedBy)
		if err != nil {
			return mapWriteError(err)
		}
		permission, err := q.GetPermissionByID(ctx, permissionID)
		if err != nil {
			return mapNotFound(err)
		}
		override = domain.PermissionOverride{
			ID:             id,
			UserID:         userID,
			PermissionID:   permissionID,
			PermissionCode: permission.Code,
			Effect:         effect,
			Note:           note,
		}
		rows, err := q.ListOverridesForUser(ctx, userID)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.PermissionID == permissionID {
				override.CreatedAt = row.CreatedAt
				override.UpdatedAt = row.UpdatedAt
				break
			}
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &grantedBy,
			Action:       domain.AuditOverrideSet,
			ResourceType: domain.ResourceUser,
			ResourceID:   userID.String(),
			Metadata:     map[string]any{"permission": permission.Code, "effect": effect, "note": note},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &override, nil
}

// DeleteOverride removes a user permission override.
func (r *Repo) DeleteOverride(ctx context.Context, userID, permissionID uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		permission, err := q.GetPermissionByID(ctx, permissionID)
		if err != nil {
			return mapNotFound(err)
		}
		if err := deleteScopedPermissionOverride(ctx, q, userID, permissionID); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditOverrideRemoved,
			ResourceType: domain.ResourceUser,
			ResourceID:   userID.String(),
			Metadata:     map[string]any{"permission": permission.Code},
			Meta:         meta,
		})
	})
}

// guardLastAdmin runs change inside a transaction while holding the admin-guard
// lock, and rolls it back if it would leave the clinic without an active admin.
// Concurrent removals serialize on the lock, so they cannot each pass the check
// and together remove every admin.
func guardLastAdmin(ctx context.Context, q *dbgen.Queries, message string, change func() error) error {
	if err := q.LockAdminGuard(ctx); err != nil {
		return err
	}
	before, err := q.CountActiveAdmins(ctx)
	if err != nil {
		return err
	}
	if changeErr := change(); changeErr != nil {
		return changeErr
	}
	after, err := q.CountActiveAdmins(ctx)
	if err != nil {
		return err
	}
	if before > 0 && after == 0 {
		return domain.NewError(domain.CodeConflict, message)
	}
	return nil
}
