package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"curaclinic/backend/internal/domain"
)

// Release removes a hold after verifying ownership.
func (s *HoldStore) Release(ctx context.Context, token string, userID uuid.UUID) error {
	result, err := holdReleaseScript.Run(ctx, s.rdb, []string{holdTokenKey(token)}, userID.String()).Int()
	if err != nil {
		return fmt.Errorf("release slot hold: %w", err)
	}
	switch result {
	case 0:
		return domain.ErrHoldExpired
	case -1:
		return domain.ErrHoldNotOwned
	default:
		return nil
	}
}

// ListHeld returns the Redis hold keys currently indexed for a doctor date.
func (s *HoldStore) ListHeld(ctx context.Context, doctorID uuid.UUID, date string) (map[string]struct{}, error) {
	indexKey := holdIndexKey(doctorID, date)
	now := s.clock.Now().Unix()
	if err := s.rdb.ZRemRangeByScore(ctx, indexKey, "-inf", strconv.FormatInt(now, 10)).Err(); err != nil {
		return nil, fmt.Errorf("prune expired holds: %w", err)
	}
	members, err := s.rdb.ZRange(ctx, indexKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("list slot holds: %w", err)
	}
	if len(members) == 0 {
		return map[string]struct{}{}, nil
	}

	pipe := s.rdb.Pipeline()
	exists := make([]*redis.IntCmd, len(members))
	for i, member := range members {
		exists[i] = pipe.Exists(ctx, member)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("verify slot holds: %w", err)
	}

	held := make(map[string]struct{}, len(members))
	stale := make([]string, 0)
	for i, member := range members {
		if exists[i].Val() > 0 {
			held[member] = struct{}{}
		} else {
			stale = append(stale, member)
		}
	}
	if len(stale) > 0 {
		_ = s.rdb.ZRem(ctx, indexKey, stale).Err()
	}
	return held, nil
}
