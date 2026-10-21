package cache

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/clock"
)

type MemoryHoldStore struct {
	mu     sync.Mutex
	holds  map[string]Hold
	tokens map[string]string
	index  map[string]map[string]struct{}
	clock  clock.Clock
	ttl    time.Duration
}

// NewMemoryHoldStore creates an in-memory hold store.
func NewMemoryHoldStore(clk clock.Clock, ttl time.Duration) *MemoryHoldStore {
	return &MemoryHoldStore{
		holds:  make(map[string]Hold),
		tokens: make(map[string]string),
		index:  make(map[string]map[string]struct{}),
		clock:  clk,
		ttl:    ttl,
	}
}

// Create reserves a slot when it is free.
func (s *MemoryHoldStore) Create(_ context.Context, input HoldInput) (*Hold, error) {
	now := s.clock.Now()
	local := input.StartUTC.In(input.Location)
	key := HoldKey(input.DoctorID, input.StartUTC, input.Location)
	indexKey := holdIndexKey(input.DoctorID, local.Format("2006-01-02"))
	token := uuid.NewString()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)

	if existing, ok := s.holds[key]; ok && existing.ExpiresAt.After(now) {
		return nil, domain.ErrSlotUnavailable
	}

	hold := Hold{
		Token:     token,
		DoctorID:  input.DoctorID,
		PatientID: input.PatientID,
		StartUTC:  input.StartUTC.UTC(),
		Date:      local.Format("2006-01-02"),
		Key:       key,
		IndexKey:  indexKey,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
	s.holds[key] = hold
	s.tokens[token] = key
	if s.index[indexKey] == nil {
		s.index[indexKey] = make(map[string]struct{})
	}
	s.index[indexKey][key] = struct{}{}

	clone := hold
	return &clone, nil
}

// Get returns the active hold for a slot, or nil.
func (s *MemoryHoldStore) Get(_ context.Context, doctorID uuid.UUID, start time.Time, location *time.Location) (*Hold, error) {
	now := s.clock.Now()
	key := HoldKey(doctorID, start, location)

	s.mu.Lock()
	defer s.mu.Unlock()
	hold, ok := s.holds[key]
	if !ok {
		return nil, nil
	}
	if !hold.ExpiresAt.After(now) {
		s.removeLocked(key)
		return nil, nil
	}
	clone := hold
	return &clone, nil
}

// Release removes a hold after verifying ownership.
func (s *MemoryHoldStore) Release(_ context.Context, token string, userID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(s.clock.Now())

	key, ok := s.tokens[token]
	if !ok {
		return domain.ErrHoldExpired
	}
	hold, ok := s.holds[key]
	if !ok {
		delete(s.tokens, token)
		return domain.ErrHoldExpired
	}
	if hold.PatientID != userID {
		return domain.ErrHoldNotOwned
	}
	s.removeLocked(key)
	return nil
}

// ListHeld returns the hold keys currently indexed for a doctor date.
func (s *MemoryHoldStore) ListHeld(_ context.Context, doctorID uuid.UUID, date string) (map[string]struct{}, error) {
	now := s.clock.Now()
	indexKey := holdIndexKey(doctorID, date)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)

	held := make(map[string]struct{}, len(s.index[indexKey]))
	for key := range s.index[indexKey] {
		held[key] = struct{}{}
	}
	return held, nil
}

func (s *MemoryHoldStore) pruneLocked(now time.Time) {
	for key, hold := range s.holds {
		if !hold.ExpiresAt.After(now) {
			s.removeLocked(key)
		}
	}
}

func (s *MemoryHoldStore) removeLocked(key string) {
	hold, ok := s.holds[key]
	if !ok {
		return
	}
	delete(s.holds, key)
	delete(s.tokens, hold.Token)
	if index, ok := s.index[hold.IndexKey]; ok {
		delete(index, key)
		if len(index) == 0 {
			delete(s.index, hold.IndexKey)
		}
	}
}
