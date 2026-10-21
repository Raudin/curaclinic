package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
	"curaclinic/backend/internal/platform/password"
	"curaclinic/backend/internal/platform/token"
)

// AuthResult is returned by login and refresh.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         domain.UserSummary
	Roles        []string
	Permissions  []string
}

// Account is the current-user payload.
type Account struct {
	User        domain.UserSummary
	Roles       []string
	Permissions []string
}

// AuthService implements login, refresh, logout and password management.
type AuthService struct {
	store      AuthStore
	rbac       AccessReader
	tokens     *token.Service
	bcryptCost int
	refreshTTL time.Duration
	idleTTL    time.Duration
	clock      clock.Clock

	dummyHashOnce sync.Once
	dummyHash     string
}

// NewAuthService creates the authentication service.
func NewAuthService(store AuthStore, rbac AccessReader, tokens *token.Service, bcryptCost int, refreshTTL, idleTTL time.Duration, clk clock.Clock) *AuthService {
	return &AuthService{
		store:      store,
		rbac:       rbac,
		tokens:     tokens,
		bcryptCost: bcryptCost,
		refreshTTL: refreshTTL,
		idleTTL:    idleTTL,
		clock:      clk,
	}
}

// Login verifies credentials and starts a new refresh-token family.
func (s *AuthService) Login(ctx context.Context, email, plainPassword string, meta domain.RequestMeta) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || plainPassword == "" {
		return nil, domain.NewError(domain.CodeUnauthorized, "invalid email or password")
	}

	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if domain.IsCode(err, domain.CodeNotFound) {
			// Spend the same bcrypt work as a real check so response time
			// does not reveal which emails are registered.
			password.Verify(s.timingHash(), plainPassword)
			return nil, domain.ErrUnauthorized.WithCause(errors.New("unknown email"))
		}
		return nil, err
	}

	hash, err := s.store.GetPasswordHash(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if !password.Verify(hash, plainPassword) {
		return nil, domain.NewError(domain.CodeUnauthorized, "invalid email or password")
	}
	if user.Status != domain.UserStatusActive {
		return nil, domain.NewError(domain.CodeForbidden, "this account is inactive; contact the clinic administrator")
	}

	if err := s.store.TouchLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}
	return s.startSession(ctx, *user, uuid.New(), meta)
}

// timingHash returns a throwaway bcrypt hash at the configured cost, used to
// equalize login timing for unknown emails.
func (s *AuthService) timingHash() string {
	s.dummyHashOnce.Do(func() {
		hash, err := password.Hash("curaclinic-timing-placeholder", s.bcryptCost)
		if err == nil {
			s.dummyHash = hash
		}
	})
	return s.dummyHash
}

func (s *AuthService) startSession(ctx context.Context, user domain.User, familyID uuid.UUID, meta domain.RequestMeta) (*AuthResult, error) {
	now := s.clock.Now()
	refreshToken, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	sessionID := uuid.New()
	if _, sessionErr := s.store.CreateSession(ctx, domain.NewSession{
		ID:               sessionID,
		UserID:           user.ID,
		FamilyID:         familyID,
		RefreshTokenHash: token.HashRefreshToken(refreshToken),
		ExpiresAt:        now.Add(s.refreshTTL),
		IP:               meta.IP,
		UserAgent:        meta.UserAgent,
	}); sessionErr != nil {
		return nil, sessionErr
	}

	accessToken, expiresAt, err := s.tokens.IssueAccess(user.ID, sessionID, now)
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
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiresAt.Sub(now).Seconds()),
		User:         *summary,
		Roles:        access.Roles,
		Permissions:  access.Permissions,
	}, nil
}
