-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, first_name, last_name, phone, status)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('email'),
    sqlc.arg('password_hash'),
    sqlc.arg('first_name'),
    sqlc.arg('last_name'),
    sqlc.narg('phone'),
    sqlc.arg('status')
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = sqlc.arg('email') AND deleted_at IS NULL;

-- name: GetActiveUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg('id') AND status = 'active' AND deleted_at IS NULL;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = sqlc.arg('password_hash'), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    phone = sqlc.narg('phone'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET status = sqlc.arg('status')::user_status, updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: ListUsers :many
SELECT u.id, u.email, u.first_name, u.last_name, u.phone, u.status, u.created_at, u.updated_at,
       COALESCE(ARRAY(
         SELECT identity.role_name FROM (
           SELECT r.name AS role_name
           FROM user_roles ur2 JOIN roles r ON r.id = ur2.role_id AND r.deleted_at IS NULL
           WHERE ur2.user_id = u.id AND ur2.deleted_at IS NULL
           UNION SELECT 'patient'::text WHERE EXISTS (SELECT 1 FROM patients px WHERE px.user_id = u.id AND px.deleted_at IS NULL)
           UNION SELECT 'doctor'::text WHERE EXISTS (SELECT 1 FROM doctors dx WHERE dx.user_id = u.id AND dx.deleted_at IS NULL)
           UNION SELECT 'admin'::text WHERE EXISTS (SELECT 1 FROM admins ax WHERE ax.user_id = u.id AND ax.deleted_at IS NULL)
         ) identity ORDER BY identity.role_name
       ), '{}')::text[] AS roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id AND ur.deleted_at IS NULL
LEFT JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
WHERE u.deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR u.email ILIKE '%' || sqlc.narg('search')::text || '%'
    OR u.first_name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR u.last_name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
  AND (
    sqlc.narg('role')::text IS NULL
    OR (sqlc.narg('role')::text = 'patient' AND EXISTS (SELECT 1 FROM patients px WHERE px.user_id = u.id AND px.deleted_at IS NULL))
    OR EXISTS (
        SELECT 1
        FROM user_roles ur_filter
        JOIN roles r_filter ON r_filter.id = ur_filter.role_id AND r_filter.deleted_at IS NULL
        WHERE ur_filter.user_id = u.id
          AND ur_filter.deleted_at IS NULL
          AND r_filter.name = sqlc.narg('role')::text
    )
  )
GROUP BY u.id
ORDER BY u.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountUsers :one
SELECT COUNT(*)::bigint AS total
FROM users u
WHERE u.deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR u.email ILIKE '%' || sqlc.narg('search')::text || '%'
    OR u.first_name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR u.last_name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
  AND (
    sqlc.narg('role')::text IS NULL
    OR (sqlc.narg('role')::text = 'patient' AND EXISTS (SELECT 1 FROM patients px WHERE px.user_id = u.id AND px.deleted_at IS NULL))
    OR EXISTS (
        SELECT 1
        FROM user_roles ur_filter
        JOIN roles r_filter ON r_filter.id = ur_filter.role_id AND r_filter.deleted_at IS NULL
        WHERE ur_filter.user_id = u.id
          AND ur_filter.deleted_at IS NULL
          AND r_filter.name = sqlc.narg('role')::text
    )
  );

-- name: GetUserWithRoles :one
SELECT u.id, u.email, u.first_name, u.last_name, u.phone, u.status, u.created_at, u.updated_at,
       COALESCE(ARRAY(
         SELECT identity.role_name FROM (
           SELECT r.name AS role_name
           FROM user_roles ur2 JOIN roles r ON r.id = ur2.role_id AND r.deleted_at IS NULL
           WHERE ur2.user_id = u.id AND ur2.deleted_at IS NULL
           UNION SELECT 'patient'::text WHERE EXISTS (SELECT 1 FROM patients px WHERE px.user_id = u.id AND px.deleted_at IS NULL)
           UNION SELECT 'doctor'::text WHERE EXISTS (SELECT 1 FROM doctors dx WHERE dx.user_id = u.id AND dx.deleted_at IS NULL)
           UNION SELECT 'admin'::text WHERE EXISTS (SELECT 1 FROM admins ax WHERE ax.user_id = u.id AND ax.deleted_at IS NULL)
         ) identity ORDER BY identity.role_name
       ), '{}')::text[] AS roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id AND ur.deleted_at IS NULL
LEFT JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
WHERE u.id = sqlc.arg('id') AND u.deleted_at IS NULL
GROUP BY u.id;

-- name: FindUserIDsByPermission :many
SELECT DISTINCT ur.user_id
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id AND rp.deleted_at IS NULL
WHERE ur.role_id = sqlc.arg('role_id') AND ur.deleted_at IS NULL;
