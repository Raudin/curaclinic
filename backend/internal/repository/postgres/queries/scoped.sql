-- Scoped identity, role-scope and shared-profile statements. These used to be
-- hand-written in the generated dbgen package; keeping them here makes them
-- visible to the sqlc drift check.

-- name: CreatePatientIdentity :exec
INSERT INTO patients (user_id) VALUES (sqlc.arg('user_id'))
ON CONFLICT (user_id) DO NOTHING;

-- name: HasPatientIdentity :one
SELECT EXISTS (
    SELECT 1 FROM patients
    WHERE user_id = sqlc.arg('user_id') AND deleted_at IS NULL
);

-- name: RoleScope :one
SELECT rs.scope::text
FROM role_scopes rs
JOIN roles r ON r.id = rs.role_id
WHERE rs.role_id = sqlc.arg('role_id') AND r.deleted_at IS NULL;

-- name: ScopedRole :one
SELECT rs.scope::text
FROM role_scopes rs
JOIN roles r ON r.id = rs.role_id
WHERE rs.role_id = sqlc.arg('role_id') AND r.deleted_at IS NULL;

-- name: CreateScopedRole :one
WITH new_role AS (
    INSERT INTO roles (name, description, is_system)
    VALUES (sqlc.arg('name'), sqlc.narg('description'), false)
    RETURNING id
)
INSERT INTO role_scopes (role_id, scope)
SELECT new_role.id, sqlc.arg('scope')::permission_scope FROM new_role
RETURNING role_id;

-- name: SetRolePermissions :exec
INSERT INTO role_permissions (role_id, permission_id)
SELECT sqlc.arg('role_id'), p.id
FROM permissions p
JOIN role_scopes rs ON rs.role_id = sqlc.arg('role_id')
WHERE p.id = ANY(sqlc.arg('permission_ids')::uuid[])
  AND p.code LIKE rs.scope::text || '.%'
  AND p.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- name: UpdateRoleScopePermissions :execrows
UPDATE roles
SET name = sqlc.arg('name'), description = sqlc.narg('description'), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ClearDoctorRoles :exec
UPDATE doctor_roles SET deleted_at = now()
WHERE doctor_id = sqlc.arg('doctor_id') AND deleted_at IS NULL;

-- name: ClearAdminRoles :exec
UPDATE admin_roles SET deleted_at = now()
WHERE admin_id = sqlc.arg('admin_id') AND deleted_at IS NULL;

-- name: ListRoleScopesByIDs :many
SELECT rs.role_id, rs.scope::text AS scope
FROM role_scopes rs
JOIN roles r ON r.id = rs.role_id AND r.deleted_at IS NULL
WHERE rs.role_id = ANY(sqlc.arg('role_ids')::uuid[]);

-- name: DoctorIdentityExists :one
SELECT EXISTS (
    SELECT 1 FROM doctors
    WHERE user_id = sqlc.arg('user_id') AND deleted_at IS NULL
);

-- name: EnsureAdminIdentity :exec
INSERT INTO admins (user_id)
SELECT id FROM users WHERE id = sqlc.arg('user_id') AND deleted_at IS NULL
ON CONFLICT (user_id) DO NOTHING;

-- name: UpsertDoctorPermissionOverride :one
INSERT INTO doctor_permission_overrides (doctor_id, permission_id, effect, granted_by, note)
SELECT sqlc.arg('user_id'), sqlc.arg('permission_id'), sqlc.arg('effect')::override_effect,
       sqlc.arg('granted_by'), sqlc.narg('note')
WHERE EXISTS (
    SELECT 1 FROM doctors
    WHERE user_id = sqlc.arg('user_id') AND deleted_at IS NULL
)
ON CONFLICT (doctor_id, permission_id) WHERE deleted_at IS NULL
DO UPDATE SET effect = EXCLUDED.effect, granted_by = EXCLUDED.granted_by,
              note = EXCLUDED.note, updated_at = now()
RETURNING id;

-- name: UpsertAdminPermissionOverride :one
INSERT INTO admin_permission_overrides (admin_id, permission_id, effect, granted_by, note)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('permission_id'),
    sqlc.arg('effect')::override_effect,
    sqlc.arg('granted_by'),
    sqlc.narg('note')
)
ON CONFLICT (admin_id, permission_id) WHERE deleted_at IS NULL
DO UPDATE SET effect = EXCLUDED.effect, granted_by = EXCLUDED.granted_by,
              note = EXCLUDED.note, updated_at = now()
RETURNING id;

-- name: DeleteDoctorPermissionOverride :exec
UPDATE doctor_permission_overrides
SET deleted_at = now(), updated_at = now()
WHERE doctor_id = sqlc.arg('user_id')
  AND permission_id = sqlc.arg('permission_id')
  AND deleted_at IS NULL;

-- name: DeleteAdminPermissionOverride :exec
UPDATE admin_permission_overrides
SET deleted_at = now(), updated_at = now()
WHERE admin_id = sqlc.arg('user_id')
  AND permission_id = sqlc.arg('permission_id')
  AND deleted_at IS NULL;

-- name: SetUserAvatar :exec
UPDATE users
SET avatar_url = NULLIF(sqlc.arg('avatar_url')::text, ''), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SetSharedIdentity :exec
UPDATE users
SET date_of_birth = NULLIF(sqlc.arg('date_of_birth')::text, '')::date,
    avatar_url = NULLIF(sqlc.arg('avatar_url')::text, ''),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SharedIdentity :one
SELECT COALESCE(to_char(date_of_birth, 'YYYY-MM-DD'), '')::text AS date_of_birth, avatar_url
FROM users
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
