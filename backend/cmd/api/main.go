package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/config"
	"curaclinic/backend/internal/db"
	"curaclinic/backend/internal/logger"
	"curaclinic/backend/internal/platform/clock"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/repository/postgres"
	"curaclinic/backend/internal/router"
	"curaclinic/backend/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("could not load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	appLogger, err := logger.New(cfg.Env, cfg.LogLevel)
	if err != nil {
		log.Fatalf("logger error: %v", err)
	}
	defer func() { _ = appLogger.Sync() }()

	gin.SetMode(cfg.GinMode)

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStartup()

	pool, err := db.NewPostgres(startupCtx, cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer pool.Close()

	var rdb *redis.Client
	if strings.EqualFold(cfg.RedisAddr, "memory") {
		appLogger.Warn("Redis is disabled; using in-memory slot holds and database-backed permissions (development only)")
	} else {
		rdb, err = cache.NewRedis(startupCtx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisTLS)
		if err != nil {
			appLogger.Fatal("redis connection failed", zap.Error(err))
		}
		defer func() { _ = rdb.Close() }()
	}

	repo := postgres.New(pool)
	appClock := clock.Real
	permissionCache := cache.NewPermissionCache(rdb, cfg.RBACCacheTTL)
	rbacService := service.NewRBACService(repo, permissionCache)
	tokenService := token.New(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)

	authService := service.NewAuthService(repo, rbacService, tokenService, cfg.BcryptCost, cfg.RefreshTokenTTL, cfg.SessionIdleTTL, appClock)
	userService := service.NewUserService(repo, cfg.BcryptCost)
	doctorService := service.NewDoctorService(repo, cfg.ClinicTimezone, cfg.BcryptCost)
	var slotHolds service.SlotHoldStore
	if rdb != nil {
		slotHolds = cache.NewHoldStore(rdb, appClock, cfg.HoldTTL)
	} else {
		slotHolds = cache.NewMemoryHoldStore(appClock, cfg.HoldTTL)
	}
	availabilityService := service.NewAvailabilityService(repo, slotHolds, appClock, cfg.BookingLeadTime, cfg.BookingHorizonDays)
	holdService := service.NewHoldService(availabilityService, slotHolds)
	appointmentService := service.NewAppointmentService(repo, availabilityService, slotHolds, appClock)
	auditService := service.NewAuditService(repo)

	engine := router.New(router.Deps{
		Config:       cfg,
		Logger:       appLogger,
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

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info("CuraClinic API listening", zap.String("addr", server.Addr), zap.String("env", cfg.Env))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		appLogger.Fatal("API server stopped unexpectedly", zap.Error(err))
	case <-shutdownCtx.Done():
		appLogger.Info("shutdown signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("graceful shutdown failed", zap.Error(err))
		_ = server.Close()
	}
	appLogger.Info("API stopped")
}
