package service

import (
	"context"

	"curaclinic/backend/internal/domain"
)

// AuditService exposes the audit trail to authorised administrators.
type AuditService struct {
	store AuditStore
}

// NewAuditService creates the audit service.
func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

// List returns a filtered page of audit entries.
func (s *AuditService) List(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditLog, int64, error) {
	filter.Limit, filter.Offset = normalizePagination(filter.Limit, filter.Offset)
	return s.store.ListAuditLogs(ctx, filter)
}
