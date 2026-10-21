-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TYPE user_status AS ENUM ('active', 'inactive');

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    phone text,
    date_of_birth date,
    avatar_url text,
    status user_status NOT NULL DEFAULT 'active',
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    CONSTRAINT users_email_not_blank CHECK (length(btrim(email)) > 0),
    CONSTRAINT users_names_not_blank CHECK (length(btrim(first_name)) > 0 AND length(btrim(last_name)) > 0)
);

CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX users_status_idx ON users (status) WHERE deleted_at IS NULL;

CREATE TABLE patients (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE doctors (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    specialty text NOT NULL,
    bio text,
    experience_years integer NOT NULL DEFAULT 0,
    languages text[] NOT NULL DEFAULT '{}',
    timezone text NOT NULL DEFAULT 'Africa/Nairobi',
    slot_minutes integer NOT NULL DEFAULT 30,
    is_accepting_patients boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT doctors_specialty_not_blank CHECK (length(btrim(specialty)) > 0),
    CONSTRAINT doctors_experience_non_negative CHECK (experience_years >= 0),
    CONSTRAINT doctors_slot_minutes_range CHECK (slot_minutes BETWEEN 5 AND 180)
);
CREATE INDEX doctors_accepting_idx ON doctors (is_accepting_patients) WHERE deleted_at IS NULL;

CREATE TABLE admins (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family_id uuid NOT NULL,
    refresh_token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    replaced_by_session_id uuid REFERENCES auth_sessions (id) ON DELETE SET NULL,
    ip text,
    user_agent text,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX auth_sessions_refresh_hash_key ON auth_sessions (refresh_token_hash);
CREATE INDEX auth_sessions_user_idx ON auth_sessions (user_id);
CREATE INDEX auth_sessions_active_user_idx ON auth_sessions (user_id) WHERE revoked_at IS NULL;
CREATE INDEX auth_sessions_family_idx ON auth_sessions (family_id);

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS doctors;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
