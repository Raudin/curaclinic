package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

func mapSession(row dbgen.AuthSession) domain.Session {
	return domain.Session{
		ID:               row.ID,
		UserID:           row.UserID,
		FamilyID:         row.FamilyID,
		RefreshTokenHash: row.RefreshTokenHash,
		ExpiresAt:        row.ExpiresAt,
		RevokedAt:        row.RevokedAt,
		LastUsedAt:       row.LastUsedAt,
		CreatedAt:        row.CreatedAt,
	}
}

// CreateSession stores a new refresh-token session.
func (r *Repo) CreateSession(ctx context.Context, input domain.NewSession) (*domain.Session, error) {
	row, err := r.q.CreateSession(ctx, dbgen.CreateSessionParams{
		ID:               input.ID,
		UserID:           input.UserID,
		FamilyID:         input.FamilyID,
		RefreshTokenHash: input.RefreshTokenHash,
		ExpiresAt:        input.ExpiresAt,
		Ip:               optionalString(input.IP),
		UserAgent:        optionalString(input.UserAgent),
	})
	if err != nil {
		return nil, err
	}
	session := mapSession(row)
	return &session, nil
}

// GetSessionByHash finds a session by the hash of its refresh token.
func (r *Repo) GetSessionByHash(ctx context.Context, hash string) (*domain.Session, error) {
	row, err := r.q.GetSessionByHash(ctx, hash)
	if err != nil {
		return nil, mapNotFound(err)
	}
	session := mapSession(row)
	return &session, nil
}

// RotateSession stores the replacement session, then revokes and links the old
// one in the same transaction. Creating first satisfies the self-referencing
// foreign key on replaced_by_session_id.
func (r *Repo) RotateSession(ctx context.Context, oldSessionID uuid.UUID, input domain.NewSession) error {
	return r.WithTx(ctx, func(q *dbgen.Queries) error {
		_, err := q.CreateSession(ctx, dbgen.CreateSessionParams{
			ID:               input.ID,
			UserID:           input.UserID,
			FamilyID:         input.FamilyID,
			RefreshTokenHash: input.RefreshTokenHash,
			ExpiresAt:        input.ExpiresAt,
			Ip:               optionalString(input.IP),
			UserAgent:        optionalString(input.UserAgent),
		})
		if err != nil {
			return err
		}
		affected, err := q.RevokeSessionAndLink(ctx, dbgen.RevokeSessionAndLinkParams{
			ID:                  oldSessionID,
			ReplacedBySessionID: &input.ID,
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			return domain.NewError(domain.CodeUnauthorized, "this refresh token has already been used")
		}
		// Stamp activity on the replacement session so idle expiry is measured
		// from the last successful refresh.
		return q.TouchSession(ctx, input.ID)
	})
}

// SessionActive reports whether access tokens issued for a session are still
// honored: its refresh-token family must have a live session.
func (r *Repo) SessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return r.q.SessionFamilyActive(ctx, sessionID)
}

// GetPasswordHash returns the stored bcrypt hash for a user.
func (r *Repo) GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	hash, err := r.q.GetUserPasswordHash(ctx, userID)
	if err != nil {
		return "", mapNotFound(err)
	}
	return hash, nil
}

// RevokeSession revokes one session.
func (r *Repo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return r.q.RevokeSession(ctx, id)
}

// RevokeSessionFamily revokes a refresh-token family after reuse detection.
func (r *Repo) RevokeSessionFamily(ctx context.Context, familyID uuid.UUID) error {
	return r.q.RevokeSessionFamily(ctx, familyID)
}

// RevokeUserSessions revokes every active session for a user.
func (r *Repo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.q.RevokeUserSessions(ctx, userID)
}
