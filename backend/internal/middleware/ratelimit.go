package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/httpx"
)

var rateLimitScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return current
`)

// RateLimitConfig describes one fixed-window limiter.
type RateLimitConfig struct {
	Scope    string
	Limit    int
	Window   time.Duration
	Identity func(*gin.Context) string
}

// RateLimit enforces a Redis fixed-window limit. It fails open when Redis is
// unavailable so a cache outage cannot take the API down.
func RateLimit(rdb *redis.Client, log *zap.Logger, cfg RateLimitConfig) gin.HandlerFunc {
	identity := cfg.Identity
	if identity == nil {
		identity = func(c *gin.Context) string { return c.ClientIP() }
	}

	return func(c *gin.Context) {
		if rdb == nil || cfg.Limit <= 0 || cfg.Window <= 0 {
			c.Next()
			return
		}

		now := time.Now().UTC()
		bucket := now.UnixNano() / int64(cfg.Window)
		key := fmt.Sprintf("rl:%s:%s:%d", cfg.Scope, identity(c), bucket)
		resetAt := time.Unix(0, (bucket+1)*int64(cfg.Window))

		ctx, cancel := requestTimeout(c, 500*time.Millisecond)
		defer cancel()

		count, err := rateLimitScript.Run(ctx, rdb, []string{key}, cfg.Window.Milliseconds()).Int64()
		if err != nil {
			log.Warn("rate limiter unavailable; allowing request", zap.Error(err))
			c.Next()
			return
		}

		remaining := int64(cfg.Limit) - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if count > int64(cfg.Limit) {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			httpx.WriteError(c, log, domain.ErrRateLimited)
			return
		}
		c.Next()
	}
}
