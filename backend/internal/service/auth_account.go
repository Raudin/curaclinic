package service

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
)

// Me returns the current user's profile, roles and permissions.
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*Account, error) {
	user, err := s.store.GetUserSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	access, err := s.rbac.Access(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Account{User: *user, Roles: access.Roles, Permissions: access.Permissions}, nil
}

// ChangePassword verifies the current password, stores the new one and revokes
// every active session so a stolen token cannot outlive the change.
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	hash, err := s.store.GetPasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if !password.Verify(hash, currentPassword) {
		return domain.ValidationField("current_password", "is incorrect")
	}
	if currentPassword == newPassword {
		return domain.ValidationField("new_password", "must be different from the current password")
	}
	if validateErr := password.Validate(newPassword); validateErr != nil {
		return domain.ValidationField("new_password", validateErr.Error())
	}
	newHash, err := password.Hash(newPassword, s.bcryptCost)
	if err != nil {
		return err
	}
	if err := s.store.UpdatePasswordHash(ctx, userID, newHash); err != nil {
		return err
	}
	return s.store.RevokeUserSessions(ctx, userID)
}
