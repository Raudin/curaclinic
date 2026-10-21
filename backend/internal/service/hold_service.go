package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"curaclinic/backend/internal/cache"
	"curaclinic/backend/internal/domain"
)

// HoldService creates and releases temporary slot reservations.
type HoldService struct {
	validator SlotValidator
	holds     SlotHoldStore
}

// NewHoldService creates the hold service.
func NewHoldService(validator SlotValidator, holds SlotHoldStore) *HoldService {
	return &HoldService{validator: validator, holds: holds}
}

// Create validates a slot and reserves it for the caller.
func (s *HoldService) Create(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID, start time.Time) (*cache.Hold, error) {
	slot, err := s.validator.ValidateSlot(ctx, doctorID, start)
	if err != nil {
		return nil, err
	}
	return reserveSlot(ctx, s.holds, actor.UserID, slot)
}

// Release removes a hold owned by the caller.
func (s *HoldService) Release(ctx context.Context, actor *domain.Actor, token string) error {
	return s.holds.Release(ctx, token, actor.UserID)
}
