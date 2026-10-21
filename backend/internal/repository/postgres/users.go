package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

func mapUser(row dbgen.User) domain.User {
	return domain.User{
		ID:        row.ID,
		Email:     row.Email,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Phone:     stringValue(row.Phone),
		Status:    string(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (r *Repo) fillSharedIdentity(ctx context.Context, user *domain.User) error {
	identity, err := r.q.SharedIdentity(ctx, user.ID)
	if err != nil {
		return err
	}
	if identity.DateOfBirth != "" {
		user.DateOfBirth = identity.DateOfBirth
	}
	if identity.AvatarUrl != nil {
		user.AvatarURL = *identity.AvatarUrl
	}
	return nil
}

func mapUserSummary(row dbgen.ListUsersRow) domain.UserSummary {
	return domain.UserSummary{
		User: domain.User{
			ID:        row.ID,
			Email:     row.Email,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Phone:     stringValue(row.Phone),
			Status:    string(row.Status),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Roles: row.Roles,
	}
}

func mapUserWithRolesRow(row dbgen.GetUserWithRolesRow) domain.UserSummary {
	return domain.UserSummary{
		User: domain.User{
			ID:        row.ID,
			Email:     row.Email,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Phone:     stringValue(row.Phone),
			Status:    string(row.Status),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Roles: row.Roles,
	}
}

// CreateUser inserts a user.
func (r *Repo) CreateUser(ctx context.Context, input domain.NewUser) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, dbgen.CreateUserParams{
		ID:           input.ID,
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        optionalString(input.Phone),
		Status:       dbgen.UserStatus(input.Status),
	})
	if err != nil {
		return nil, mapUniqueConflict(err, "a user with this email already exists")
	}
	user := mapUser(row)
	if err := r.q.SetSharedIdentity(ctx, dbgen.SetSharedIdentityParams{
		ID:          row.ID,
		DateOfBirth: input.DateOfBirth,
		AvatarUrl:   input.AvatarURL,
	}); err != nil {
		return nil, err
	}
	if err := r.fillSharedIdentity(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUserWithRoles inserts a user and assigns roles atomically.
func (r *Repo) CreateUserWithRoles(ctx context.Context, input domain.NewUser, roleIDs []uuid.UUID, assignedBy uuid.UUID, meta domain.RequestMeta) (*domain.UserSummary, error) {
	var created domain.User
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.CreateUser(ctx, dbgen.CreateUserParams{
			ID:           input.ID,
			Email:        input.Email,
			PasswordHash: input.PasswordHash,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
			Phone:        optionalString(input.Phone),
			Status:       dbgen.UserStatus(input.Status),
		})
		if err != nil {
			return mapUniqueConflict(err, "a user with this email already exists")
		}
		created = mapUser(row)
		if err := q.SetSharedIdentity(ctx, dbgen.SetSharedIdentityParams{
			ID:          row.ID,
			DateOfBirth: input.DateOfBirth,
			AvatarUrl:   input.AvatarURL,
		}); err != nil {
			return err
		}
		created.DateOfBirth = input.DateOfBirth
		created.AvatarURL = input.AvatarURL
		if len(roleIDs) == 0 {
			if err := q.CreatePatientIdentity(ctx, row.ID); err != nil {
				return err
			}
		} else if err := assignScopedRoles(ctx, q, row.ID, roleIDs, assignedBy); err != nil {
			return mapNotFound(err)
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &assignedBy,
			Action:       domain.AuditUserCreated,
			ResourceType: domain.ResourceUser,
			ResourceID:   row.ID.String(),
			Metadata:     map[string]any{"email": row.Email, "role_ids": roleIDs, "patient_identity": len(roleIDs) == 0},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	roles, err := r.RoleNames(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	return &domain.UserSummary{User: created, Roles: roles}, nil
}

// GetUserByEmail finds a user by normalized email.
func (r *Repo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapNotFound(err)
	}
	user := mapUser(row)
	if err := r.fillSharedIdentity(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a non-deleted user.
func (r *Repo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	user := mapUser(row)
	if err := r.fillSharedIdentity(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetActiveUserByID finds a non-deleted, active user.
func (r *Repo) GetActiveUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetActiveUserByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	user := mapUser(row)
	if err := r.fillSharedIdentity(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserSummary returns a user with their roles.
func (r *Repo) GetUserSummary(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error) {
	row, err := r.q.GetUserWithRoles(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	summary := mapUserWithRolesRow(row)
	if err := r.fillSharedIdentity(ctx, &summary.User); err != nil {
		return nil, err
	}
	return &summary, nil
}

// EnsurePatientIdentity links patient access to an existing authenticated user.
func (r *Repo) EnsurePatientIdentity(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.CreatePatientIdentity(ctx, userID); err != nil {
		return mapWriteError(err)
	}
	return nil
}

// TouchLastLogin records a successful login.
func (r *Repo) TouchLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.q.UpdateLastLogin(ctx, id)
}
