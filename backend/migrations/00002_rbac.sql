-- +goose Up
CREATE TYPE override_effect AS ENUM ('allow', 'deny');
CREATE TYPE permission_scope AS ENUM ('doctor', 'admin');

CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text,
    is_system boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT roles_name_not_blank CHECK (length(btrim(name)) > 0)
);

CREATE UNIQUE INDEX roles_name_active_key ON roles (name) WHERE deleted_at IS NULL;

CREATE TABLE role_scopes (
    role_id uuid PRIMARY KEY REFERENCES roles (id) ON DELETE CASCADE,
    scope permission_scope NOT NULL
);

CREATE TABLE permissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource text NOT NULL,
    action text NOT NULL,
    code text NOT NULL,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT permissions_code_not_blank CHECK (length(btrim(code)) > 0),
    CONSTRAINT permissions_resource_not_blank CHECK (length(btrim(resource)) > 0),
    CONSTRAINT permissions_action_not_blank CHECK (length(btrim(action)) > 0)
);

CREATE UNIQUE INDEX permissions_code_active_key ON permissions (code) WHERE deleted_at IS NULL;
CREATE INDEX permissions_resource_idx ON permissions (resource) WHERE deleted_at IS NULL;

CREATE TABLE role_permissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX role_permissions_active_key ON role_permissions (role_id, permission_id) WHERE deleted_at IS NULL;
CREATE INDEX role_permissions_permission_idx ON role_permissions (permission_id) WHERE deleted_at IS NULL;

CREATE TABLE doctor_roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id uuid NOT NULL REFERENCES doctors (user_id) ON DELETE CASCADE,
    role_id uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    assigned_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX doctor_roles_active_key ON doctor_roles (doctor_id, role_id) WHERE deleted_at IS NULL;
CREATE INDEX doctor_roles_role_idx ON doctor_roles (role_id) WHERE deleted_at IS NULL;

CREATE TABLE admin_roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id uuid NOT NULL REFERENCES admins (user_id) ON DELETE CASCADE,
    role_id uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    assigned_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX admin_roles_active_key ON admin_roles (admin_id, role_id) WHERE deleted_at IS NULL;
CREATE INDEX admin_roles_role_idx ON admin_roles (role_id) WHERE deleted_at IS NULL;

CREATE TABLE doctor_permission_overrides (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id uuid NOT NULL REFERENCES doctors (user_id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    effect override_effect NOT NULL,
    granted_by uuid REFERENCES users (id) ON DELETE SET NULL,
    note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX doctor_permission_overrides_active_key
    ON doctor_permission_overrides (doctor_id, permission_id) WHERE deleted_at IS NULL;
CREATE INDEX doctor_permission_overrides_doctor_idx
    ON doctor_permission_overrides (doctor_id) WHERE deleted_at IS NULL;

CREATE TABLE admin_permission_overrides (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id uuid NOT NULL REFERENCES admins (user_id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    effect override_effect NOT NULL,
    granted_by uuid REFERENCES users (id) ON DELETE SET NULL,
    note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX admin_permission_overrides_active_key
    ON admin_permission_overrides (admin_id, permission_id) WHERE deleted_at IS NULL;
CREATE INDEX admin_permission_overrides_admin_idx
    ON admin_permission_overrides (admin_id) WHERE deleted_at IS NULL;

-- Compatibility read views expose a user's combined doctor/admin access to
-- existing account summaries. Writes are routed through scoped repository methods.
CREATE VIEW user_roles AS
SELECT id, doctor_id AS user_id, role_id, assigned_by, created_at, deleted_at FROM doctor_roles
UNION ALL
SELECT id, admin_id AS user_id, role_id, assigned_by, created_at, deleted_at FROM admin_roles;

CREATE VIEW user_permission_overrides AS
SELECT id, doctor_id AS user_id, permission_id, effect, granted_by, note, created_at, updated_at, deleted_at
FROM doctor_permission_overrides
UNION ALL
SELECT id, admin_id AS user_id, permission_id, effect, granted_by, note, created_at, updated_at, deleted_at
FROM admin_permission_overrides;

-- +goose Down
DROP VIEW IF EXISTS user_permission_overrides;
DROP VIEW IF EXISTS user_roles;
DROP TABLE IF EXISTS admin_permission_overrides;
DROP TABLE IF EXISTS doctor_permission_overrides;
DROP TABLE IF EXISTS admin_roles;
DROP TABLE IF EXISTS doctor_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS role_scopes;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TYPE IF EXISTS override_effect;
DROP TYPE IF EXISTS permission_scope;
