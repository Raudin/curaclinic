package router_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/config"
	"curaclinic/backend/internal/platform/clock"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/repository/postgres"
	"curaclinic/backend/internal/router"
	"curaclinic/backend/internal/service"
)

type testApp struct {
	engine       *gin.Engine
	repo         *postgres.Repo
	doctorID     uuid.UUID
	patientID    uuid.UUID
	patientEmail string
	date         string
	slot         time.Time
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := &config.Config{
		Env:                   "test",
		GinMode:               gin.TestMode,
		CORSAllowedOrigins:    []string{"http://localhost"},
		JWTSecret:             strings.Repeat("s", 40),
		JWTIssuer:             "curaclinic-api",
		JWTAudience:           "curaclinic-web",
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       24 * time.Hour,
		SessionIdleTTL:        30 * time.Minute,
		ClinicTimezone:        "Africa/Nairobi",
		BookingLeadTime:       30 * time.Minute,
		BookingHorizonDays:    90,
		HoldTTL:               5 * time.Minute,
		RBACCacheTTL:          time.Minute,
		BcryptCost:            10,
		RateLimitGlobalPerMin: 10000,
		RateLimitAuthPer15Min: 1000,
	}

	repo := postgres.New(pool)
	appClock := clock.Real
	permissionCache := cache.NewPermissionCache(rdb, cfg.RBACCacheTTL)
	rbacService := service.NewRBACService(repo, permissionCache)
	tokenService := token.New(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)
	authService := service.NewAuthService(repo, rbacService, tokenService, cfg.BcryptCost, cfg.RefreshTokenTTL, cfg.SessionIdleTTL, appClock)
	userService := service.NewUserService(repo, cfg.BcryptCost)
	doctorService := service.NewDoctorService(repo, cfg.ClinicTimezone, cfg.BcryptCost)
	holdStore := cache.NewHoldStore(rdb, appClock, cfg.HoldTTL)
	availabilityService := service.NewAvailabilityService(repo, holdStore, appClock, cfg.BookingLeadTime, cfg.BookingHorizonDays)
	holdService := service.NewHoldService(availabilityService, holdStore)
	appointmentService := service.NewAppointmentService(repo, availabilityService, holdStore, appClock)
	auditService := service.NewAuditService(repo)

	engine := router.New(router.Deps{
		Config:       cfg,
		Logger:       zap.NewNop(),
		DB:           pool,
		Redis:        rdb,
		Tokens:       tokenService,
		UserLoader:   repo,
		Auth:         authService,
		RBAC:         rbacService,
		Users:        userService,
		Doctors:      doctorService,
		Availability: availabilityService,
		Holds:        holdService,
		Appointments: appointmentService,
		Audit:        auditService,
	})

	doctorID, patientID, patientEmail, date, slot := seedFlowData(ctx, t, repo, cfg)
	return &testApp{
		engine:       engine,
		repo:         repo,
		doctorID:     doctorID,
		patientID:    patientID,
		patientEmail: patientEmail,
		date:         date,
		slot:         slot,
	}
}
