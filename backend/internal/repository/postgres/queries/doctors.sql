-- name: CreateDoctorProfile :one
INSERT INTO doctors (
    user_id, specialty, bio, experience_years, languages, timezone, slot_minutes, is_accepting_patients
)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('specialty'),
    sqlc.narg('bio'),
    sqlc.arg('experience_years'),
    sqlc.arg('languages')::text[],
    sqlc.arg('timezone'),
    sqlc.arg('slot_minutes'),
    sqlc.arg('is_accepting_patients')
)
RETURNING *;

-- name: GetDoctorProfileByUserID :one
SELECT dp.id AS profile_id, dp.user_id, dp.specialty, dp.bio, dp.experience_years, dp.languages,
       u.avatar_url, dp.timezone, dp.slot_minutes, dp.is_accepting_patients,
       u.first_name, u.last_name, u.email
FROM doctors dp
JOIN users u ON u.id = dp.user_id AND u.deleted_at IS NULL
WHERE dp.user_id = sqlc.arg('user_id') AND dp.deleted_at IS NULL;

-- name: UpdateDoctorProfile :one
UPDATE doctors
SET specialty = sqlc.arg('specialty'),
    bio = sqlc.narg('bio'),
    experience_years = sqlc.arg('experience_years'),
    languages = sqlc.arg('languages')::text[],
    timezone = sqlc.arg('timezone'),
    slot_minutes = sqlc.arg('slot_minutes'),
    is_accepting_patients = sqlc.arg('is_accepting_patients'),
    updated_at = now()
WHERE user_id = sqlc.arg('user_id') AND deleted_at IS NULL
RETURNING *;

-- name: ListDoctors :many
SELECT dp.id AS profile_id, dp.user_id, dp.specialty, dp.bio, dp.experience_years, dp.languages,
       u.avatar_url, dp.timezone, dp.slot_minutes, dp.is_accepting_patients,
       u.first_name, u.last_name, u.email
FROM doctors dp
JOIN users u ON u.id = dp.user_id AND u.deleted_at IS NULL AND u.status = 'active'
WHERE dp.deleted_at IS NULL
  AND (
    sqlc.narg('specialty')::text IS NULL
    OR dp.specialty ILIKE '%' || sqlc.narg('specialty')::text || '%'
  )
  AND (
    sqlc.narg('accepting')::boolean IS NULL
    OR dp.is_accepting_patients = sqlc.narg('accepting')
  )
ORDER BY u.last_name, u.first_name;
