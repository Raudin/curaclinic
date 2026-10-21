-- +goose Up
CREATE TYPE appointment_status AS ENUM ('confirmed', 'cancelled');
CREATE TYPE appointment_event_type AS ENUM ('created', 'cancelled', 'rescheduled');

CREATE TABLE appointments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id uuid NOT NULL REFERENCES doctors (user_id) ON DELETE RESTRICT,
    patient_id uuid NOT NULL REFERENCES patients (user_id) ON DELETE RESTRICT,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    reason text NOT NULL,
    status appointment_status NOT NULL DEFAULT 'confirmed',
    created_by uuid REFERENCES users (id) ON DELETE SET NULL,
    cancelled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT appointments_time_order CHECK (end_time > start_time),
    CONSTRAINT appointments_reason_not_blank CHECK (length(btrim(reason)) > 0),
    CONSTRAINT appointments_doctor_no_overlap
        EXCLUDE USING gist (doctor_id WITH =, tstzrange(start_time, end_time, '[)') WITH &&)
        WHERE (status = 'confirmed' AND deleted_at IS NULL),
    CONSTRAINT appointments_patient_no_overlap
        EXCLUDE USING gist (patient_id WITH =, tstzrange(start_time, end_time, '[)') WITH &&)
        WHERE (status = 'confirmed' AND deleted_at IS NULL)
);

CREATE INDEX appointments_doctor_time_idx ON appointments (doctor_id, start_time) WHERE deleted_at IS NULL;
CREATE INDEX appointments_patient_time_idx ON appointments (patient_id, start_time) WHERE deleted_at IS NULL;
CREATE INDEX appointments_status_time_idx ON appointments (status, start_time) WHERE deleted_at IS NULL;

CREATE TABLE appointment_events (
    id bigserial PRIMARY KEY,
    appointment_id uuid NOT NULL REFERENCES appointments (id) ON DELETE CASCADE,
    event_type appointment_event_type NOT NULL,
    from_start_time timestamptz,
    to_start_time timestamptz,
    event_reason text,
    created_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX appointment_events_appointment_idx ON appointment_events (appointment_id, created_at);

CREATE TABLE audit_logs (
    id bigserial PRIMARY KEY,
    actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    request_id text,
    ip text,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_resource_idx ON audit_logs (resource_type, resource_id, created_at DESC);
CREATE INDEX audit_logs_actor_idx ON audit_logs (actor_id, created_at DESC);
CREATE INDEX audit_logs_created_idx ON audit_logs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS appointment_events;
DROP TABLE IF EXISTS appointments;
DROP TYPE IF EXISTS appointment_event_type;
DROP TYPE IF EXISTS appointment_status;
