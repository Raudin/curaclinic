// Package cache contains the Redis-backed stores used for temporary holds,
// permission caching and rate limiting.
package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates and verifies a Redis client.
func NewRedis(ctx context.Context, addr, password string, database int, useTLS bool) (*redis.Client, error) {
	options := &redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              database,
		MaxRetries:      2,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        20,
		MinIdleConns:    2,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	if useTLS {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(options)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
