-- name: ListEffectivePermissionCodes :many
SELECT DISTINCT p.code
FROM permissions p
WHERE p.deleted_at IS NULL
  AND (
    EXISTS (
        SELECT 1
        FROM role_permissions rp
        JOIN user_roles ur ON ur.role_id = rp.role_id AND ur.deleted_at IS NULL
        JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
        WHERE rp.permission_id = p.id
          AND rp.deleted_at IS NULL
          AND ur.user_id = sqlc.arg('user_id')
    )
    OR EXISTS (
        SELECT 1
        FROM user_permission_overrides allow_override
        WHERE allow_override.user_id = sqlc.arg('user_id')
          AND allow_override.permission_id = p.id
          AND allow_override.effect = 'allow'
          AND allow_override.deleted_at IS NULL
    )
  )
  AND NOT EXISTS (
    SELECT 1
    FROM user_permission_overrides deny_override
    WHERE deny_override.user_id = sqlc.arg('user_id')
      AND deny_override.permission_id = p.id
      AND deny_override.effect = 'deny'
      AND deny_override.deleted_at IS NULL
  )
ORDER BY p.code;

-- name: ListRoleNamesForUser :many
SELECT name FROM (
    SELECT r.name
    FROM roles r
    JOIN user_roles ur ON ur.role_id = r.id AND ur.deleted_at IS NULL
    WHERE ur.user_id = sqlc.arg('user_id') AND r.deleted_at IS NULL
    UNION
    SELECT 'patient'::text AS name
    FROM patients p
    WHERE p.user_id = sqlc.arg('user_id') AND p.deleted_at IS NULL
    UNION
    SELECT 'doctor'::text AS name
    FROM doctors d
    WHERE d.user_id = sqlc.arg('user_id') AND d.deleted_at IS NULL
    UNION
    SELECT 'admin'::text AS name
    FROM admins a
    WHERE a.user_id = sqlc.arg('user_id') AND a.deleted_at IS NULL
) identities
ORDER BY name;

-- name: ListRoles :many
SELECT * FROM roles WHERE deleted_at IS NULL ORDER BY name;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = sqlc.arg('name') AND deleted_at IS NULL;

-- name: CreateRole :one
INSERT INTO roles (name, description, is_system)
VALUES (sqlc.arg('name'), sqlc.narg('description'), false)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = sqlc.arg('name'), description = sqlc.narg('description'), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteRole :exec
UPDATE roles SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL AND is_system = false;

-- name: ListRolePermissionCodes :many
SELECT p.code
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at IS NULL
WHERE rp.role_id = sqlc.arg('role_id') AND rp.deleted_at IS NULL
ORDER BY p.code;

-- name: SoftDeleteRolePermissions :exec
UPDATE role_permissions SET deleted_at = now()
WHERE role_id = sqlc.arg('role_id') AND deleted_at IS NULL;

-- name: InsertRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES (sqlc.arg('role_id'), sqlc.arg('permission_id'))
ON CONFLICT DO NOTHING;

-- name: ListPermissions :many
SELECT * FROM permissions WHERE deleted_at IS NULL ORDER BY code;

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetPermissionByCode :one
SELECT * FROM permissions WHERE code = sqlc.arg('code') AND deleted_at IS NULL;

-- name: FindPermissionIDsByCodes :many
SELECT id FROM permissions
WHERE code = ANY(sqlc.arg('codes')::text[]) AND deleted_at IS NULL;

-- name: GetRoleScope :one
SELECT rs.scope FROM role_scopes rs
JOIN roles r ON r.id = rs.role_id AND r.deleted_at IS NULL
WHERE rs.role_id = sqlc.arg('id');

-- name: AssignDoctorRole :exec
INSERT INTO doctor_roles (doctor_id, role_id, assigned_by)
VALUES (sqlc.arg('doctor_id'), sqlc.arg('role_id'), sqlc.narg('assigned_by'))
ON CONFLICT DO NOTHING;

-- name: AssignAdminRole :exec
INSERT INTO admin_roles (admin_id, role_id, assigned_by)
VALUES (sqlc.arg('admin_id'), sqlc.arg('role_id'), sqlc.narg('assigned_by'))
ON CONFLICT DO NOTHING;

-- name: ListRoleIDsForUser :many
SELECT role_id FROM user_roles
WHERE user_id = sqlc.arg('user_id') AND deleted_at IS NULL;

-- name: ListOverridesForUser :many
SELECT o.id, o.user_id, o.permission_id, p.code AS permission_code, o.effect, o.note, o.created_at, o.updated_at
FROM user_permission_overrides o
JOIN permissions p ON p.id = o.permission_id AND p.deleted_at IS NULL
WHERE o.user_id = sqlc.arg('user_id') AND o.deleted_at IS NULL
ORDER BY p.code;

-- An active admin is an active user who holds the system admin role. The
-- admins identity row alone grants nothing, so it is not counted.
-- name: CountActiveAdmins :one
SELECT COUNT(DISTINCT a.user_id)::bigint AS total
FROM admins a
JOIN users u ON u.id = a.user_id AND u.deleted_at IS NULL
JOIN admin_roles ar ON ar.admin_id = a.user_id AND ar.deleted_at IS NULL
JOIN roles r ON r.id = ar.role_id AND r.deleted_at IS NULL AND r.is_system AND r.name = 'admin'
WHERE a.deleted_at IS NULL
  AND u.status = 'active';

-- name: UserHoldsAdminRole :one
SELECT EXISTS (
    SELECT 1
    FROM admin_roles ar
    JOIN roles r ON r.id = ar.role_id AND r.deleted_at IS NULL AND r.is_system AND r.name = 'admin'
    WHERE ar.admin_id = sqlc.arg('user_id') AND ar.deleted_at IS NULL
) AS holds;

-- Serializes changes that could remove the last active admin.
-- name: LockAdminGuard :exec
SELECT pg_advisory_xact_lock(hashtext('curaclinic.last_admin_guard'));
