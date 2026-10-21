package postgres

import (
	"context"
	"encoding/json"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// RecordAudit writes an audit entry outside a larger transaction.
func (r *Repo) RecordAudit(ctx context.Context, entry domain.AuditEntry) error {
	return r.audit(ctx, r.q, entry)
}

// ListAuditLogs returns a filtered page of audit entries and the total count.
func (r *Repo) ListAuditLogs(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditLog, int64, error) {
	params := dbgen.ListAuditLogsParams{
		ResourceType: filter.ResourceType,
		ResourceID:   filter.ResourceID,
		ActorID:      filter.ActorID,
		Limit:        int32FromInt(filter.Limit),
		Offset:       int32FromInt(filter.Offset),
	}
	rows, err := r.q.ListAuditLogs(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.AuditLog, 0, len(rows))
	for _, row := range rows {
		metadata := map[string]any{}
		if len(row.Metadata) > 0 {
			_ = json.Unmarshal(row.Metadata, &metadata)
		}
		items = append(items, domain.AuditLog{
			ID:           row.ID,
			ActorID:      row.ActorID,
			Action:       row.Action,
			ResourceType: row.ResourceType,
			ResourceID:   stringValue(row.ResourceID),
			Metadata:     metadata,
			RequestID:    stringValue(row.RequestID),
			IP:           stringValue(row.Ip),
			UserAgent:    stringValue(row.UserAgent),
			CreatedAt:    row.CreatedAt,
		})
	}
	total, err := r.q.CountAuditLogs(ctx, dbgen.CountAuditLogsParams{
		ResourceType: filter.ResourceType,
		ResourceID:   filter.ResourceID,
		ActorID:      filter.ActorID,
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
