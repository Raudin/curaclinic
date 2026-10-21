package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// ListUsers returns a page of users and the total count.
func (r *Repo) ListUsers(ctx context.Context, search, role string, limit, offset int) ([]domain.UserSummary, int64, error) {
	params := dbgen.ListUsersParams{
		Search: optionalString(search),
		Role:   optionalString(role),
		Limit:  int32FromInt(limit),
		Offset: int32FromInt(offset),
	}
	rows, err := r.q.ListUsers(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.UserSummary, 0, len(rows))
	for _, row := range rows {
		item := mapUserSummary(row)
		if fillErr := r.fillSharedIdentity(ctx, &item.User); fillErr != nil {
			return nil, 0, fillErr
		}
		items = append(items, item)
	}
	total, err := r.q.CountUsers(ctx, dbgen.CountUsersParams{Search: params.Search, Role: params.Role})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateUserProfile updates names and phone.
func (r *Repo) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName, phone, dateOfBirth, avatarURL string) (*domain.User, error) {
	row, err := r.q.UpdateUserProfile(ctx, dbgen.UpdateUserProfileParams{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     optionalString(phone),
	})
	if err != nil {
		return nil, mapNotFound(err)
	}
	if err := r.q.SetSharedIdentity(ctx, dbgen.SetSharedIdentityParams{
		ID:          id,
		DateOfBirth: dateOfBirth,
		AvatarUrl:   avatarURL,
	}); err != nil {
		return nil, err
	}
	user := mapUser(row)
	user.DateOfBirth = dateOfBirth
	user.AvatarURL = avatarURL
	return &user, nil
}

// UpdateUserStatus activates or deactivates a user.
func (r *Repo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status string, actorID uuid.UUID, meta domain.RequestMeta) (*domain.User, error) {
	var user domain.User
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		if err := guardLastAdmin(ctx, q, "cannot deactivate the last active admin", func() error {
			row, err := q.UpdateUserStatus(ctx, dbgen.UpdateUserStatusParams{
				ID:     id,
				Status: dbgen.UserStatus(status),
			})
			if err != nil {
				return mapNotFound(err)
			}
			user = mapUser(row)
			return nil
		}); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actorID,
			Action:       domain.AuditUserStatusUpdated,
			ResourceType: domain.ResourceUser,
			ResourceID:   id.String(),
			Metadata:     map[string]any{"status": status},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePasswordHash replaces the stored password hash.
func (r *Repo) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	if err := r.q.UpdatePassword(ctx, dbgen.UpdatePasswordParams{ID: id, PasswordHash: hash}); err != nil {
		return mapNotFound(err)
	}
	return nil
}
