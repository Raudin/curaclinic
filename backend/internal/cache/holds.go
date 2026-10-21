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
	"curaclinic/backend/internal/platform/clock"
)

// Hold is a temporary reservation of a doctor slot.
type Hold struct {
	Token     string    `json:"token"`
	DoctorID  uuid.UUID `json:"doctor_id"`
	PatientID uuid.UUID `json:"patient_id"`
	StartUTC  time.Time `json:"start_utc"`
	Date      string    `json:"date"`
	Key       string    `json:"key"`
	IndexKey  string    `json:"index_key"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// HoldInput creates a hold for a slot.
type HoldInput struct {
	DoctorID  uuid.UUID
	PatientID uuid.UUID
	StartUTC  time.Time
	Location  *time.Location
}

// HoldStore manages Redis slot holds and their daily indexes.
type HoldStore struct {
	rdb        *redis.Client
	clock      clock.Clock
	defaultTTL time.Duration
}

// NewHoldStore creates a hold store.
func NewHoldStore(rdb *redis.Client, clk clock.Clock, ttl time.Duration) *HoldStore {
	return &HoldStore{rdb: rdb, clock: clk, defaultTTL: ttl}
}

var holdCreateScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
  return 0
end
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
redis.call('ZADD', KEYS[2], ARGV[3], KEYS[1])
redis.call('SET', KEYS[3], KEYS[1], 'PX', ARGV[2])
return 1
`)

var holdReleaseScript = redis.NewScript(`
local key = redis.call('GET', KEYS[1])
if not key then
  return 0
end
local raw = redis.call('GET', key)
if not raw then
  redis.call('DEL', KEYS[1])
  return 0
end
local ok, hold = pcall(cjson.decode, raw)
if not ok then
  redis.call('DEL', KEYS[1])
  redis.call('DEL', key)
  return 0
end
if hold.patient_id ~= ARGV[1] then
  return -1
end
redis.call('DEL', key)
redis.call('DEL', KEYS[1])
redis.call('ZREM', hold.index_key, hold.key)
return 1
`)

// HoldKey builds the Redis key for a slot in the doctor's timezone.
func HoldKey(doctorID uuid.UUID, start time.Time, location *time.Location) string {
	local := start.In(location)
	return fmt.Sprintf("hold:%s:%s:%s", doctorID, local.Format("2006-01-02"), local.Format("1504"))
}

func holdIndexKey(doctorID uuid.UUID, date string) string {
	return fmt.Sprintf("holds:%s:%s", doctorID, date)
}

func holdTokenKey(token string) string { return "hold:token:" + token }

// Create atomically reserves a slot. It returns domain.ErrSlotUnavailable when
// another hold already exists.
func (s *HoldStore) Create(ctx context.Context, input HoldInput) (*Hold, error) {
	now := s.clock.Now()
	local := input.StartUTC.In(input.Location)
	date := local.Format("2006-01-02")
	key := HoldKey(input.DoctorID, input.StartUTC, input.Location)
	indexKey := holdIndexKey(input.DoctorID, date)
	token := uuid.NewString()
	expiresAt := now.Add(s.defaultTTL)

	hold := Hold{
		Token:     token,
		DoctorID:  input.DoctorID,
		PatientID: input.PatientID,
		StartUTC:  input.StartUTC.UTC(),
		Date:      date,
		Key:       key,
		IndexKey:  indexKey,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	payload, err := json.Marshal(hold)
	if err != nil {
		return nil, fmt.Errorf("marshal hold: %w", err)
	}

	created, err := holdCreateScript.Run(ctx, s.rdb,
		[]string{key, indexKey, holdTokenKey(token)},
		string(payload),
		s.defaultTTL.Milliseconds(),
		expiresAt.Unix(),
	).Int()
	if err != nil {
		return nil, fmt.Errorf("create slot hold: %w", err)
	}
	if created == 0 {
		return nil, domain.ErrSlotUnavailable
	}
	return &hold, nil
}

// Get returns the active hold for a slot, or nil when none exists.
func (s *HoldStore) Get(ctx context.Context, doctorID uuid.UUID, start time.Time, location *time.Location) (*Hold, error) {
	raw, err := s.rdb.Get(ctx, HoldKey(doctorID, start, location)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read slot hold: %w", err)
	}
	var hold Hold
	if err := json.Unmarshal(raw, &hold); err != nil {
		return nil, fmt.Errorf("decode slot hold: %w", err)
	}
	if !hold.ExpiresAt.After(s.clock.Now()) {
		return nil, nil
	}
	return &hold, nil
}
