-- name: ListWorkingHoursForWeekday :many
SELECT id, doctor_id, weekday,
       (EXTRACT(EPOCH FROM start_time) / 60)::int AS start_minutes,
       (EXTRACT(EPOCH FROM end_time) / 60)::int AS end_minutes
FROM doctor_working_hours
WHERE doctor_id = sqlc.arg('doctor_id')
  AND weekday = sqlc.arg('weekday')
  AND deleted_at IS NULL
ORDER BY start_time;

-- name: ListWorkingHours :many
SELECT id, doctor_id, weekday,
       (EXTRACT(EPOCH FROM start_time) / 60)::int AS start_minutes,
       (EXTRACT(EPOCH FROM end_time) / 60)::int AS end_minutes
FROM doctor_working_hours
WHERE doctor_id = sqlc.arg('doctor_id') AND deleted_at IS NULL
ORDER BY weekday, start_time;

-- name: SoftDeleteWorkingHours :exec
UPDATE doctor_working_hours
SET deleted_at = now(), updated_at = now()
WHERE doctor_id = sqlc.arg('doctor_id') AND deleted_at IS NULL;

-- name: InsertWorkingHour :exec
INSERT INTO doctor_working_hours (doctor_id, weekday, start_time, end_time, created_by)
VALUES (
    sqlc.arg('doctor_id'),
    sqlc.arg('weekday'),
    make_interval(secs => sqlc.arg('start_seconds')::int)::time,
    make_interval(secs => sqlc.arg('end_seconds')::int)::time,
    sqlc.narg('created_by')
);

-- name: CreateScheduleBlock :one
INSERT INTO schedule_blocks (doctor_id, start_time, end_time, reason, created_by)
VALUES (
    sqlc.arg('doctor_id'),
    sqlc.arg('start_time'),
    sqlc.arg('end_time'),
    sqlc.arg('reason'),
    sqlc.narg('created_by')
)
RETURNING *;

-- name: GetScheduleBlockByID :one
SELECT * FROM schedule_blocks
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SoftDeleteScheduleBlock :exec
UPDATE schedule_blocks
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListScheduleBlocksInRange :many
SELECT * FROM schedule_blocks
WHERE doctor_id = sqlc.arg('doctor_id')
  AND deleted_at IS NULL
  AND start_time < sqlc.arg('range_end')
  AND end_time > sqlc.arg('range_start')
ORDER BY start_time;

-- name: ListScheduleBlocks :many
SELECT * FROM schedule_blocks
WHERE doctor_id = sqlc.arg('doctor_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR end_time > sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR start_time < sqlc.narg('to_time'))
ORDER BY start_time;
