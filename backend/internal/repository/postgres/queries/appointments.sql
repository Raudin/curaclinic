-- name: CreateAppointment :one
INSERT INTO appointments (doctor_id, patient_id, start_time, end_time, reason, created_by)
VALUES (
    sqlc.arg('doctor_id'),
    sqlc.arg('patient_id'),
    sqlc.arg('start_time'),
    sqlc.arg('end_time'),
    sqlc.arg('reason'),
    sqlc.narg('created_by')
)
RETURNING *;

-- name: GetAppointmentByID :one
SELECT a.id, a.doctor_id, a.patient_id, a.start_time, a.end_time, a.reason, a.status,
       a.created_by, a.cancelled_at, a.created_at, a.updated_at,
       d.first_name AS doctor_first_name, d.last_name AS doctor_last_name, d.email AS doctor_email,
       p.first_name AS patient_first_name, p.last_name AS patient_last_name, p.email AS patient_email,
       COALESCE(dp.specialty, '')::text AS doctor_specialty
FROM appointments a
JOIN users d ON d.id = a.doctor_id
JOIN users p ON p.id = a.patient_id
LEFT JOIN doctors dp ON dp.user_id = a.doctor_id AND dp.deleted_at IS NULL
WHERE a.id = sqlc.arg('id') AND a.deleted_at IS NULL;

-- name: ListAppointments :many
SELECT a.id, a.doctor_id, a.patient_id, a.start_time, a.end_time, a.reason, a.status,
       a.created_by, a.cancelled_at, a.created_at, a.updated_at,
       d.first_name AS doctor_first_name, d.last_name AS doctor_last_name, d.email AS doctor_email,
       p.first_name AS patient_first_name, p.last_name AS patient_last_name, p.email AS patient_email,
       COALESCE(dp.specialty, '')::text AS doctor_specialty
FROM appointments a
JOIN users d ON d.id = a.doctor_id
JOIN users p ON p.id = a.patient_id
LEFT JOIN doctors dp ON dp.user_id = a.doctor_id AND dp.deleted_at IS NULL
WHERE a.deleted_at IS NULL
  AND (sqlc.narg('patient_id')::uuid IS NULL OR a.patient_id = sqlc.narg('patient_id'))
  AND (sqlc.narg('doctor_id')::uuid IS NULL OR a.doctor_id = sqlc.narg('doctor_id'))
  AND (sqlc.narg('status')::text IS NULL OR a.status::text = sqlc.narg('status'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR a.start_time >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR a.end_time <= sqlc.narg('to_time'))
  AND (sqlc.narg('ends_after')::timestamptz IS NULL OR a.end_time > sqlc.narg('ends_after'))
ORDER BY
  CASE WHEN sqlc.arg('ascending')::boolean THEN a.start_time END ASC,
  CASE WHEN NOT sqlc.arg('ascending')::boolean THEN a.start_time END DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAppointments :one
SELECT COUNT(*)::bigint AS total
FROM appointments a
WHERE a.deleted_at IS NULL
  AND (sqlc.narg('patient_id')::uuid IS NULL OR a.patient_id = sqlc.narg('patient_id'))
  AND (sqlc.narg('doctor_id')::uuid IS NULL OR a.doctor_id = sqlc.narg('doctor_id'))
  AND (sqlc.narg('status')::text IS NULL OR a.status::text = sqlc.narg('status'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR a.start_time >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR a.end_time <= sqlc.narg('to_time'))
  AND (sqlc.narg('ends_after')::timestamptz IS NULL OR a.end_time > sqlc.narg('ends_after'));

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = sqlc.arg('status')::appointment_status,
    cancelled_at = CASE WHEN sqlc.arg('status') = 'cancelled' THEN now() ELSE NULL END,
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL AND status = 'confirmed'
RETURNING *;

-- name: UpdateAppointmentTimes :one
UPDATE appointments
SET start_time = sqlc.arg('start_time'),
    end_time = sqlc.arg('end_time'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL AND status = 'confirmed'
RETURNING *;

-- name: ListConfirmedAppointmentsInRange :many
SELECT start_time, end_time
FROM appointments
WHERE doctor_id = sqlc.arg('doctor_id')
  AND status = 'confirmed'
  AND deleted_at IS NULL
  AND start_time < sqlc.arg('range_end')
  AND end_time > sqlc.arg('range_start')
ORDER BY start_time;

-- name: CreateAppointmentEvent :one
INSERT INTO appointment_events (appointment_id, event_type, from_start_time, to_start_time, event_reason, created_by)
VALUES (
    sqlc.arg('appointment_id'),
    sqlc.arg('event_type')::appointment_event_type,
    sqlc.narg('from_start_time'),
    sqlc.narg('to_start_time'),
    sqlc.narg('event_reason'),
    sqlc.narg('created_by')
)
RETURNING *;

-- name: ListAppointmentEvents :many
SELECT e.id, e.appointment_id, e.event_type, e.from_start_time, e.to_start_time, e.event_reason,
       e.created_by, e.created_at,
       COALESCE(NULLIF(btrim(u.first_name || ' ' || u.last_name), ''), 'System')::text AS actor_name
FROM appointment_events e
LEFT JOIN users u ON u.id = e.created_by
WHERE e.appointment_id = sqlc.arg('appointment_id')
ORDER BY e.created_at ASC, e.id ASC;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (actor_id, action, resource_type, resource_id, metadata, request_id, ip, user_agent)
VALUES (
    sqlc.narg('actor_id'),
    sqlc.arg('action'),
    sqlc.arg('resource_type'),
    sqlc.narg('resource_id'),
    sqlc.arg('metadata')::jsonb,
    sqlc.narg('request_id'),
    sqlc.narg('ip'),
    sqlc.narg('user_agent')
);

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE (sqlc.narg('resource_type')::text IS NULL OR resource_type = sqlc.narg('resource_type'))
  AND (sqlc.narg('resource_id')::text IS NULL OR resource_id = sqlc.narg('resource_id'))
  AND (sqlc.narg('actor_id')::uuid IS NULL OR actor_id = sqlc.narg('actor_id'))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAuditLogs :one
SELECT COUNT(*)::bigint AS total FROM audit_logs
WHERE (sqlc.narg('resource_type')::text IS NULL OR resource_type = sqlc.narg('resource_type'))
  AND (sqlc.narg('resource_id')::text IS NULL OR resource_id = sqlc.narg('resource_id'))
  AND (sqlc.narg('actor_id')::uuid IS NULL OR actor_id = sqlc.narg('actor_id'));
