// Package config loads and validates runtime configuration from the environment.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config contains every setting the API needs to start.
type Config struct {
	Env                string
	Port               string
	GinMode            string
	CORSAllowedOrigins []string
	TrustedProxies     []string

	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	RedisTLS        bool
	JWTSecret       string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	SessionIdleTTL  time.Duration

	ClinicTimezone     string
	BookingLeadTime    time.Duration
	BookingHorizonDays int
	HoldTTL            time.Duration
	RBACCacheTTL       time.Duration
	BcryptCost         int

	LogLevel              string
	RateLimitGlobalPerMin int
	RateLimitAuthPer15Min int
}

// Load reads .env (when present) and the process environment. Environment
// variables always win over .env values.
func Load() (*Config, error) {
	// Loading .env is optional; missing files are not an error.
	_ = godotenv.Load()

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("API_PORT", "8080")
	v.SetDefault("GIN_MODE", "debug")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	v.SetDefault("TRUSTED_PROXIES", "")
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_TLS", false)
	v.SetDefault("JWT_ISSUER", "curaclinic-api")
	v.SetDefault("JWT_AUDIENCE", "curaclinic-web")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "720h")
	v.SetDefault("JWT_SESSION_IDLE_TTL", "30m")
	v.SetDefault("CLINIC_TIMEZONE", "Africa/Nairobi")
	v.SetDefault("BOOKING_LEAD_TIME", "30m")
	v.SetDefault("BOOKING_HORIZON_DAYS", 90)
	v.SetDefault("HOLD_TTL", "5m")
	v.SetDefault("RBAC_CACHE_TTL", "60s")
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("RATE_LIMIT_GLOBAL_PER_MIN", 240)
	v.SetDefault("RATE_LIMIT_AUTH_PER_15MIN", 20)

	cfg := &Config{
		Env:                   strings.ToLower(strings.TrimSpace(v.GetString("APP_ENV"))),
		Port:                  strings.TrimSpace(v.GetString("API_PORT")),
		GinMode:               strings.TrimSpace(v.GetString("GIN_MODE")),
		CORSAllowedOrigins:    splitCSV(v.GetString("CORS_ALLOWED_ORIGINS")),
		TrustedProxies:        splitCSV(v.GetString("TRUSTED_PROXIES")),
		DatabaseURL:           strings.TrimSpace(v.GetString("DATABASE_URL")),
		RedisAddr:             strings.TrimSpace(v.GetString("REDIS_ADDR")),
		RedisPassword:         v.GetString("REDIS_PASSWORD"),
		RedisDB:               v.GetInt("REDIS_DB"),
		RedisTLS:              v.GetBool("REDIS_TLS"),
		JWTSecret:             v.GetString("JWT_SECRET"),
		JWTIssuer:             strings.TrimSpace(v.GetString("JWT_ISSUER")),
		JWTAudience:           strings.TrimSpace(v.GetString("JWT_AUDIENCE")),
		AccessTokenTTL:        v.GetDuration("JWT_ACCESS_TTL"),
		RefreshTokenTTL:       v.GetDuration("JWT_REFRESH_TTL"),
		SessionIdleTTL:        v.GetDuration("JWT_SESSION_IDLE_TTL"),
		ClinicTimezone:        strings.TrimSpace(v.GetString("CLINIC_TIMEZONE")),
		BookingLeadTime:       v.GetDuration("BOOKING_LEAD_TIME"),
		BookingHorizonDays:    v.GetInt("BOOKING_HORIZON_DAYS"),
		HoldTTL:               v.GetDuration("HOLD_TTL"),
		RBACCacheTTL:          v.GetDuration("RBAC_CACHE_TTL"),
		BcryptCost:            v.GetInt("BCRYPT_COST"),
		LogLevel:              strings.ToLower(strings.TrimSpace(v.GetString("LOG_LEVEL"))),
		RateLimitGlobalPerMin: v.GetInt("RATE_LIMIT_GLOBAL_PER_MIN"),
		RateLimitAuthPer15Min: v.GetInt("RATE_LIMIT_AUTH_PER_15MIN"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate fails fast on unsafe or unusable configuration.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}
	if c.IsProduction() && strings.EqualFold(c.RedisAddr, "memory") {
		return fmt.Errorf("REDIS_ADDR=memory is only allowed outside production; configure a real Redis instance")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.JWTIssuer == "" || c.JWTAudience == "" {
		return fmt.Errorf("JWT_ISSUER and JWT_AUDIENCE are required")
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TTL and JWT_REFRESH_TTL must be positive durations")
	}
	if c.SessionIdleTTL <= 0 {
		return fmt.Errorf("JWT_SESSION_IDLE_TTL must be a positive duration")
	}
	if c.SessionIdleTTL > c.RefreshTokenTTL {
		return fmt.Errorf("JWT_SESSION_IDLE_TTL cannot exceed JWT_REFRESH_TTL")
	}
	if _, err := time.LoadLocation(c.ClinicTimezone); err != nil {
		return fmt.Errorf("CLINIC_TIMEZONE: %w", err)
	}
	if c.BookingLeadTime < 0 {
		return fmt.Errorf("BOOKING_LEAD_TIME cannot be negative")
	}
	if c.BookingHorizonDays <= 0 || c.BookingHorizonDays > 730 {
		return fmt.Errorf("BOOKING_HORIZON_DAYS must be between 1 and 730")
	}
	if c.HoldTTL < 30*time.Second || c.HoldTTL > 30*time.Minute {
		return fmt.Errorf("HOLD_TTL must be between 30s and 30m")
	}
	if c.BcryptCost < 10 || c.BcryptCost > 14 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}
	if c.RateLimitGlobalPerMin <= 0 || c.RateLimitAuthPer15Min <= 0 {
		return fmt.Errorf("rate limit values must be positive")
	}
	switch c.GinMode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("GIN_MODE must be debug, release or test, got %q", c.GinMode)
	}
	return nil
}

// IsProduction reports whether the API is running with production settings.
func (c *Config) IsProduction() bool { return c.Env == "production" }

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
