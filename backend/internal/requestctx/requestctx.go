// Package requestctx shares request-scoped values between middleware,
// handlers and the HTTP error writer without importing each other.
package requestctx

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

const (
	keyRequestID = "curaclinic.request_id"
	keyActor     = "curaclinic.actor"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// EnsureRequestID returns the incoming request id when it is safe to echo,
// otherwise it generates a new UUID.
func EnsureRequestID(c *gin.Context) string {
	if incoming := c.GetHeader("X-Request-ID"); requestIDPattern.MatchString(incoming) {
		return incoming
	}
	return uuid.NewString()
}

// SetRequestID stores the request id on the Gin context.
func SetRequestID(c *gin.Context, id string) { c.Set(keyRequestID, id) }

// RequestID returns the request id, or an empty string when unset.
func RequestID(c *gin.Context) string {
	if value, ok := c.Get(keyRequestID); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

// SetActor stores the authenticated actor.
func SetActor(c *gin.Context, actor *domain.Actor) { c.Set(keyActor, actor) }

// Actor returns the authenticated actor, if any.
func Actor(c *gin.Context) *domain.Actor {
	if value, ok := c.Get(keyActor); ok {
		if actor, ok := value.(*domain.Actor); ok {
			return actor
		}
	}
	return nil
}

// Meta builds audit metadata for the current request.
func Meta(c *gin.Context) domain.RequestMeta {
	return domain.RequestMeta{
		RequestID: RequestID(c),
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}
