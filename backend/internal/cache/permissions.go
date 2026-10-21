package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"curaclinic/backend/internal/domain"
)

const permissionVersionKey = "rbac:version"

// PermissionCache caches resolved roles and permissions per user.
type PermissionCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewPermissionCache creates a cache. A nil client disables caching.
func NewPermissionCache(rdb *redis.Client, ttl time.Duration) *PermissionCache {
	return &PermissionCache{rdb: rdb, ttl: ttl}
}

func (c *PermissionCache) version(ctx context.Context) (string, bool) {
	if c.rdb == nil {
		return "0", false
	}
	value, err := c.rdb.Get(ctx, permissionVersionKey).Result()
	if errors.Is(err, redis.Nil) {
		return "0", true
	}
	if err != nil {
		return "0", false
	}
	return value, true
}

func (c *PermissionCache) key(version string, userID uuid.UUID) string {
	return fmt.Sprintf("rbac:perms:%s:%s", version, userID)
}

func (c *PermissionCache) Get(ctx context.Context, userID uuid.UUID) (domain.UserAccess, string, bool) {
	version, ok := c.version(ctx)
	if !ok {
		return domain.UserAccess{}, "", false
	}
	raw, err := c.rdb.Get(ctx, c.key(version, userID)).Bytes()
	if err != nil {
		return domain.UserAccess{}, version, false
	}
	var access domain.UserAccess
	if err := json.Unmarshal(raw, &access); err != nil {
		return domain.UserAccess{}, version, false
	}
	return access, version, true
}

// Set stores an access snapshot under the version returned by Get.
func (c *PermissionCache) Set(ctx context.Context, version string, userID uuid.UUID, access domain.UserAccess) {
	if c.rdb == nil || version == "" {
		return
	}
	raw, err := json.Marshal(access)
	if err != nil {
		return
	}
	_ = c.rdb.Set(ctx, c.key(version, userID), raw, c.ttl).Err()
}

// Invalidate bumps the global RBAC version, effectively invalidating every
// cached permission set. Old entries expire through their TTL.
func (c *PermissionCache) Invalidate(ctx context.Context) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Incr(ctx, permissionVersionKey).Err()
}
