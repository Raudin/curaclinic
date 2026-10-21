// Package postgres implements the persistence layer with pgx and sqlc-generated
// queries. One Repo value satisfies the narrow interfaces declared by services.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"curaclinic/backend/internal/db"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// Repo is the PostgreSQL repository.
type Repo struct {
	pool *pgxpool.Pool
	q    *dbgen.Queries
}

// New creates a repository backed by the pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, q: dbgen.New(pool)}
}

// Ping verifies the database connection.
func (r *Repo) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }

// WithTx runs fn inside a transaction.
func (r *Repo) WithTx(ctx context.Context, fn func(q *dbgen.Queries) error) error {
	return db.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(dbgen.New(tx))
	})
}

func (r *Repo) audit(ctx context.Context, q *dbgen.Queries, entry domain.AuditEntry) error {
	if entry.Action == "" {
		return nil
	}
	metadata := entry.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	return q.InsertAuditLog(ctx, dbgen.InsertAuditLogParams{
		ActorID:      entry.ActorID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   optionalString(entry.ResourceID),
		Metadata:     raw,
		RequestID:    optionalString(entry.Meta.RequestID),
		Ip:           optionalString(entry.Meta.IP),
		UserAgent:    optionalString(entry.Meta.UserAgent),
	})
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23P01":
			return domain.ErrSlotUnavailable.WithCause(err)
		case "23503":
			return domain.Validation("a referenced record does not exist", nil).WithCause(err)
		}
	}
	return err
}

func mapUniqueConflict(err error, message string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.NewError(domain.CodeConflict, message).WithCause(err)
	}
	return err
}

func stringSlice(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}

func mapBlockError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
		return domain.NewError(domain.CodeConflict, "the schedule block overlaps an existing block").WithCause(err)
	}
	return err
}

// int32FromInt narrows a pagination value after clamping it to the range the
// database accepts.
func int32FromInt(value int) int32 {
	switch {
	case value > math.MaxInt32:
		return math.MaxInt32
	case value < math.MinInt32:
		return math.MinInt32
	default:
		return int32(value) //nolint:gosec // clamped to the int32 range above
	}
}

// weekdayInt16 narrows a weekday that the service validates as 0-6.
func weekdayInt16(value int) int16 {
	switch {
	case value < 0:
		return 0
	case value > 6:
		return 6
	default:
		return int16(value) //nolint:gosec // clamped to 0-6 above
	}
}

func mapAppointmentWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewError(domain.CodeConflict, "the appointment is no longer active").WithCause(err)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23P01":
			return domain.ErrSlotUnavailable.WithCause(err)
		case "23503":
			return domain.Validation("doctor or patient does not exist", nil).WithCause(err)
		}
	}
	return err
}
