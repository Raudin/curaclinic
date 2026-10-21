package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"curaclinic/backend/internal/requestctx"
)

// AccessLog logs one structured line per request.
func AccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		userID := ""
		if actor := requestctx.Actor(c); actor != nil {
			userID = actor.UserID.String()
		}

		log.Info("http request",
			zap.String("request_id", requestctx.RequestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", userID),
		)
	}
}

// Recovery converts panics into 500 responses and logs the stack.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("panic recovered",
					zap.String("request_id", requestctx.RequestID(c)),
					zap.Any("panic", recovered),
					zap.ByteString("stack", debug.Stack()),
				)
				if c.Writer.Written() {
					c.Abort()
					return
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "internal_error",
						"message":    "an unexpected error occurred",
						"request_id": requestctx.RequestID(c),
					},
				})
			}
		}()
		c.Next()
	}
}
