package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
)

// RequirePermission aborts with 403 unless the actor holds the permission.
func RequirePermission(log *zap.Logger, code string) gin.HandlerFunc {
	return RequireAnyPermission(log, code)
}

// RequireAnyPermission aborts with 403 unless the actor holds at least one of
// the permissions.
func RequireAnyPermission(log *zap.Logger, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := requestctx.Actor(c)
		if actor == nil {
			httpx.WriteError(c, log, domain.ErrUnauthorized)
			return
		}
		if !actor.CanAny(codes...) {
			httpx.WriteError(c, log, domain.ErrForbidden)
			return
		}
		c.Next()
	}
}
