package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// assignScopedRoles replaces the doctor/admin role assignments for a user. The
// statements live in queries/scoped.sql; only the scope-dependent
// orchestration remains in Go.
func assignScopedRoles(ctx context.Context, q *dbgen.Queries, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	if err := q.ClearDoctorRoles(ctx, userID); err != nil {
		return err
	}
	if err := q.ClearAdminRoles(ctx, userID); err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}

	rows, err := q.ListRoleScopesByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	scopes := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		scopes[row.RoleID] = row.Scope
	}

	for _, roleID := range roleIDs {
		scope, ok := scopes[roleID]
		if !ok {
			return pgx.ErrNoRows
		}
		switch scope {
		case "doctor":
			exists, err := q.DoctorIdentityExists(ctx, userID)
			if err != nil {
				return err
			}
			if !exists {
				return domain.Validation("doctor role requires a doctor record", nil)
			}
			if err := q.AssignDoctorRole(ctx, dbgen.AssignDoctorRoleParams{
				DoctorID:   userID,
				RoleID:     roleID,
				AssignedBy: &assignedBy,
			}); err != nil {
				return err
			}
		case "admin":
			if err := q.EnsureAdminIdentity(ctx, userID); err != nil {
				return err
			}
			if err := q.AssignAdminRole(ctx, dbgen.AssignAdminRoleParams{
				AdminID:    userID,
				RoleID:     roleID,
				AssignedBy: &assignedBy,
			}); err != nil {
				return err
			}
		default:
			return domain.Validation("unsupported role scope", nil)
		}
	}
	return nil
}

// setScopedPermissionOverride routes an override to the doctor or admin table
// based on the permission code prefix, creating the admin identity if needed.
func setScopedPermissionOverride(ctx context.Context, q *dbgen.Queries, userID, permissionID uuid.UUID, effect, note string, grantedBy uuid.UUID) (uuid.UUID, error) {
	permission, err := q.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return uuid.Nil, err
	}
	switch permissionScope(permission.Code) {
	case "doctor":
		return q.UpsertDoctorPermissionOverride(ctx, dbgen.UpsertDoctorPermissionOverrideParams{
			UserID:       userID,
			PermissionID: permissionID,
			Effect:       dbgen.OverrideEffect(effect),
			GrantedBy:    &grantedBy,
			Note:         optionalString(note),
		})
	case "admin":
		if err := q.EnsureAdminIdentity(ctx, userID); err != nil {
			return uuid.Nil, err
		}
		return q.UpsertAdminPermissionOverride(ctx, dbgen.UpsertAdminPermissionOverrideParams{
			UserID:       userID,
			PermissionID: permissionID,
			Effect:       dbgen.OverrideEffect(effect),
			GrantedBy:    &grantedBy,
			Note:         optionalString(note),
		})
	default:
		return uuid.Nil, domain.Validation("unsupported permission scope", nil)
	}
}

// deleteScopedPermissionOverride soft-deletes the override in both scoped
// tables; only one of them can hold the row.
func deleteScopedPermissionOverride(ctx context.Context, q *dbgen.Queries, userID, permissionID uuid.UUID) error {
	if err := q.DeleteDoctorPermissionOverride(ctx, dbgen.DeleteDoctorPermissionOverrideParams{
		UserID:       userID,
		PermissionID: permissionID,
	}); err != nil {
		return err
	}
	return q.DeleteAdminPermissionOverride(ctx, dbgen.DeleteAdminPermissionOverrideParams{
		UserID:       userID,
		PermissionID: permissionID,
	})
}
