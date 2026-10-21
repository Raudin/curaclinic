package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
)

// Register creates a patient account and starts a session for it.
func (s *AuthService) Register(ctx context.Context, firstName, lastName, email, phone, dateOfBirth, plainPassword string, meta domain.RequestMeta) (*AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	firstName, lastName, err = normalizeNames(firstName, lastName)
	if err != nil {
		return nil, err
	}
	dateOfBirth, err = normalizeDateOfBirth(dateOfBirth)
	if err != nil {
		return nil, err
	}
	if validateErr := password.Validate(plainPassword); validateErr != nil {
		return nil, domain.ValidationField("password", validateErr.Error())
	}
	hash, err := password.Hash(plainPassword, s.bcryptCost)
	if err != nil {
		return nil, err
	}

	existing, err := s.store.GetUserByEmail(ctx, normalizedEmail)
	if err == nil {
		existingHash, hashErr := s.store.GetPasswordHash(ctx, existing.ID)
		if hashErr != nil {
			return nil, hashErr
		}
		if !password.Verify(existingHash, plainPassword) {
			return nil, domain.NewError(domain.CodeUnauthorized, "sign in to the existing account before adding patient access")
		}
		if existing.Status != domain.UserStatusActive {
			return nil, domain.NewError(domain.CodeForbidden, "this account is inactive; contact the clinic administrator")
		}
		if identityErr := s.store.EnsurePatientIdentity(ctx, existing.ID); identityErr != nil {
			return nil, identityErr
		}
		if invalidator, ok := s.rbac.(interface{ Invalidate(context.Context) error }); ok {
			if invalidateErr := invalidator.Invalidate(ctx); invalidateErr != nil {
				return nil, invalidateErr
			}
		}
		return s.startSession(ctx, *existing, uuid.New(), meta)
	}
	if !domain.IsCode(err, domain.CodeNotFound) {
		return nil, err
	}

	userID := uuid.New()
	summary, err := s.store.CreateUserWithRoles(ctx, domain.NewUser{
		ID:           userID,
		Email:        normalizedEmail,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        strings.TrimSpace(phone),
		DateOfBirth:  dateOfBirth,
		Status:       domain.UserStatusActive,
	}, nil, userID, meta)
	if err != nil {
		return nil, err
	}
	return s.startSession(ctx, summary.User, uuid.New(), meta)
}
