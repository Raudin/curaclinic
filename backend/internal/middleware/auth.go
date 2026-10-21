// Package middleware contains the HTTP middleware chain for the API.
package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/requestctx"
)

// UserLoader loads the authenticated user's current state.
type UserLoader interface {
	GetActiveUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	SessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

// AccessReader resolves roles and permissions.
type AccessReader interface {
	Access(ctx context.Context, userID uuid.UUID) (domain.UserAccess, error)
}

// Auth verifies the bearer token and loads fresh roles and permissions.
func Auth(tokens *token.Service, users UserLoader, access AccessReader, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			httpx.WriteError(c, log, domain.ErrUnauthorized)
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			httpx.WriteError(c, log, domain.NewError(domain.CodeUnauthorized, "authorization header must use the Bearer scheme"))
			return
		}

		claims, err := tokens.Verify(strings.TrimSpace(parts[1]))
		if err != nil {
			httpx.WriteError(c, log, domain.NewError(domain.CodeUnauthorized, "access token is invalid or expired").WithCause(err))
			return
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			httpx.WriteError(c, log, domain.ErrUnauthorized)
			return
		}

		sessionActive, err := users.SessionActive(c.Request.Context(), claims.SessionID)
		if err != nil {
			httpx.WriteError(c, log, err)
			return
		}
		if !sessionActive {
			httpx.WriteError(c, log, domain.NewError(domain.CodeUnauthorized, "this session has ended; please sign in again"))
			return
		}

		user, err := users.GetActiveUserByID(c.Request.Context(), userID)
		if err != nil {
			if domain.IsCode(err, domain.CodeNotFound) {
				httpx.WriteError(c, log, domain.NewError(domain.CodeUnauthorized, "this account is no longer active"))
				return
			}
			httpx.WriteError(c, log, err)
			return
		}
		accessSnapshot, err := access.Access(c.Request.Context(), userID)
		if err != nil {
			httpx.WriteError(c, log, err)
			return
		}

		permissions := make(map[string]struct{}, len(accessSnapshot.Permissions))
		for _, code := range accessSnapshot.Permissions {
			permissions[code] = struct{}{}
		}
		requestctx.SetActor(c, &domain.Actor{
			UserID:      user.ID,
			SessionID:   claims.SessionID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Roles:       accessSnapshot.Roles,
			Permissions: permissions,
		})
		c.Next()
	}
}
