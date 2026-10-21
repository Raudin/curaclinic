// Package logger builds the application's structured logger.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a production JSON logger outside development and a console
// logger during development.
func New(env, level string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid LOG_LEVEL %q: %w", level, err)
	}

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}
