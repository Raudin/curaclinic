package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/token"
)

// Refresh rotates a refresh token and issues a new access token.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta domain.RequestMeta) (*AuthResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, domain.NewError(domain.CodeUnauthorized, "refresh token is required")
	}

	session, err := s.store.GetSessionByHash(ctx, token.HashRefreshToken(refreshToken))
	if err != nil {
		if domain.IsCode(err, domain.CodeNotFound) {
			return nil, domain.NewError(domain.CodeUnauthorized, "refresh token is invalid")
		}
		return nil, err
	}
	if session.RevokedAt != nil {
		// A rotated token was presented again: assume theft and revoke the family.
		_ = s.store.RevokeSessionFamily(ctx, session.FamilyID)
		return nil, domain.NewError(domain.CodeUnauthorized, "this session has been revoked; please sign in again")
	}
	if !session.ExpiresAt.After(s.clock.Now()) {
		_ = s.store.RevokeSession(ctx, session.ID)
		return nil, domain.NewError(domain.CodeUnauthorized, "refresh token has expired; please sign in again")
	}

	now := s.clock.Now()
	lastSeen := session.CreatedAt
	if session.LastUsedAt != nil {
		lastSeen = *session.LastUsedAt
	}
	if now.Sub(lastSeen) > s.idleTTL {
		_ = s.store.RevokeSession(ctx, session.ID)
		return nil, domain.NewError(domain.CodeUnauthorized, "session expired after inactivity; please sign in again")
	}

	user, err := s.store.GetActiveUserByID(ctx, session.UserID)
	if err != nil {
		if domain.IsCode(err, domain.CodeNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	nextRefresh, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	nextSessionID := uuid.New()
	if rotateErr := s.store.RotateSession(ctx, session.ID, domain.NewSession{
		ID:               nextSessionID,
		UserID:           session.UserID,
		FamilyID:         session.FamilyID,
		RefreshTokenHash: token.HashRefreshToken(nextRefresh),
		ExpiresAt:        now.Add(s.refreshTTL),
		IP:               meta.IP,
		UserAgent:        meta.UserAgent,
	}); rotateErr != nil {
		return nil, rotateErr
	}

	accessToken, expiresAt, err := s.tokens.IssueAccess(user.ID, nextSessionID, now)
	if err != nil {
		return nil, err
	}
	summary, err := s.store.GetUserSummary(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	access, err := s.rbac.Access(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: nextRefresh,
		ExpiresIn:    int64(expiresAt.Sub(now).Seconds()),
		User:         *summary,
		Roles:        access.Roles,
		Permissions:  access.Permissions,
	}, nil
}

// Logout revokes the presented refresh token. Unknown tokens are ignored.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	session, err := s.store.GetSessionByHash(ctx, token.HashRefreshToken(refreshToken))
	if err != nil {
		if domain.IsCode(err, domain.CodeNotFound) {
			return nil
		}
		return err
	}
	if session.RevokedAt != nil {
		return nil
	}
	return s.store.RevokeSession(ctx, session.ID)
}
