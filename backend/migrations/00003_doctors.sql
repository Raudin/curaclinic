-- +goose Up
CREATE TABLE doctor_working_hours (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id uuid NOT NULL REFERENCES doctors (user_id) ON DELETE CASCADE,
    weekday smallint NOT NULL,
    start_time time NOT NULL,
    end_time time NOT NULL,
    created_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT doctor_working_hours_weekday_range CHECK (weekday BETWEEN 0 AND 6),
    CONSTRAINT doctor_working_hours_time_order CHECK (start_time < end_time)
);

CREATE UNIQUE INDEX doctor_working_hours_unique_active
    ON doctor_working_hours (doctor_id, weekday, start_time, end_time) WHERE deleted_at IS NULL;
CREATE INDEX doctor_working_hours_doctor_idx
    ON doctor_working_hours (doctor_id, weekday) WHERE deleted_at IS NULL;

CREATE TABLE schedule_blocks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id uuid NOT NULL REFERENCES doctors (user_id) ON DELETE CASCADE,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    reason text NOT NULL,
    created_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT schedule_blocks_time_order CHECK (start_time < end_time),
    CONSTRAINT schedule_blocks_reason_not_blank CHECK (length(btrim(reason)) > 0),
    CONSTRAINT schedule_blocks_no_overlap
        EXCLUDE USING gist (doctor_id WITH =, tstzrange(start_time, end_time, '[)') WITH &&)
        WHERE (deleted_at IS NULL)
);

CREATE INDEX schedule_blocks_doctor_idx ON schedule_blocks (doctor_id, start_time) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS schedule_blocks;
DROP TABLE IF EXISTS doctor_working_hours;
