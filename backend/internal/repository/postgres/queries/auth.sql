-- name: CreateSession :one
INSERT INTO auth_sessions (id, user_id, family_id, refresh_token_hash, expires_at, ip, user_agent)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('user_id'),
    sqlc.arg('family_id'),
    sqlc.arg('refresh_token_hash'),
    sqlc.arg('expires_at'),
    sqlc.narg('ip'),
    sqlc.narg('user_agent')
)
RETURNING *;

-- name: GetSessionByHash :one
SELECT * FROM auth_sessions
WHERE refresh_token_hash = sqlc.arg('refresh_token_hash');

-- name: RevokeSession :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE id = sqlc.arg('id') AND revoked_at IS NULL;

-- name: RevokeSessionAndLink :execrows
UPDATE auth_sessions
SET revoked_at = now(), replaced_by_session_id = sqlc.arg('replaced_by_session_id')
WHERE id = sqlc.arg('id') AND revoked_at IS NULL;

-- name: GetUserPasswordHash :one
SELECT password_hash FROM users
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: RevokeSessionFamily :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE family_id = sqlc.arg('family_id') AND revoked_at IS NULL;

-- name: RevokeUserSessions :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE user_id = sqlc.arg('user_id') AND revoked_at IS NULL;

-- name: TouchSession :exec
UPDATE auth_sessions
SET last_used_at = now()
WHERE id = sqlc.arg('id') AND revoked_at IS NULL;

-- A session's access tokens stay valid while its refresh-token family still has
-- a live session. Rotation keeps the family alive; logout, reuse detection and
-- password changes revoke it.
-- name: SessionFamilyActive :one
SELECT EXISTS (
    SELECT 1
    FROM auth_sessions s
    JOIN auth_sessions live ON live.family_id = s.family_id
    WHERE s.id = sqlc.arg('id')
      AND live.revoked_at IS NULL
      AND live.expires_at > now()
) AS active;
