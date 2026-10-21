package middleware

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/requestctx"
)

// RequestID echoes a valid X-Request-ID or generates a new one.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := requestctx.EnsureRequestID(c)
		requestctx.SetRequestID(c, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
