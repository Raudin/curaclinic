// Package router assembles Gin middleware and routes.
package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"curaclinic/backend/internal/config"
	"curaclinic/backend/internal/handlers"
	"curaclinic/backend/internal/middleware"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/service"
)

// Deps are the infrastructure and service dependencies required by the router.
type Deps struct {
	Config       *config.Config
	Logger       *zap.Logger
	DB           *pgxpool.Pool
	Redis        *redis.Client
	Tokens       *token.Service
	UserLoader   middleware.UserLoader
	Auth         *service.AuthService
	RBAC         *service.RBACService
	Users        *service.UserService
	Doctors      *service.DoctorService
	Availability *service.AvailabilityService
	Holds        *service.HoldService
	Appointments *service.AppointmentService
	Audit        *service.AuditService
}

// New builds the HTTP engine with the production middleware order:
// request id -> recovery -> access log -> CORS -> rate limit -> auth -> RBAC.
func New(deps Deps) *gin.Engine {
	engine := gin.New()
	// Only trust forwarding headers from explicitly configured proxies so
	// rate limiting and audit logs cannot be spoofed through X-Forwarded-For.
	if len(deps.Config.TrustedProxies) == 0 {
		_ = engine.SetTrustedProxies(nil)
	} else if err := engine.SetTrustedProxies(deps.Config.TrustedProxies); err != nil {
		deps.Logger.Warn("invalid TRUSTED_PROXIES; disabling proxy headers", zap.Error(err))
		_ = engine.SetTrustedProxies(nil)
	}
	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.AccessLog(deps.Logger),
		middleware.CORS(deps.Config.CORSAllowedOrigins),
		middleware.RateLimit(deps.Redis, deps.Logger, middleware.RateLimitConfig{
			Scope:  "global",
			Limit:  deps.Config.RateLimitGlobalPerMin,
			Window: time.Minute,
		}),
	)

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "not_found",
				"message": "route not found",
			},
		})
	})

	api := engine.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.GET("/ready", readiness(deps))

	h := handlers.New(
		deps.Config,
		deps.Logger,
		deps.Auth,
		deps.RBAC,
		deps.Users,
		deps.Doctors,
		deps.Availability,
		deps.Holds,
		deps.Appointments,
		deps.Audit,
	)

	registerPublicAuthRoutes(api, h, deps)
	registerRoutes(api, h, deps)

	return engine
}

func require(deps Deps, code string) gin.HandlerFunc {
	return middleware.RequirePermission(deps.Logger, code)
}

func requireAny(deps Deps, codes ...string) gin.HandlerFunc {
	return middleware.RequireAnyPermission(deps.Logger, codes...)
}

func readiness(deps Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		checks := gin.H{"database": "ok", "redis": "ok"}
		healthy := true

		if err := deps.DB.Ping(ctx); err != nil {
			checks["database"] = "unavailable"
			healthy = false
			deps.Logger.Warn("readiness: database ping failed", zap.Error(err))
		}
		if deps.Redis == nil {
			checks["redis"] = "disabled"
		} else if deps.Redis.Ping(ctx).Err() != nil {
			checks["redis"] = "unavailable"
			healthy = false
		}

		status := "ok"
		httpStatus := http.StatusOK
		if !healthy {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
		c.JSON(httpStatus, gin.H{"status": status, "checks": checks})
	}
}
