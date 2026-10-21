-- +goose Up
INSERT INTO permissions (resource, action, code, description) VALUES
    ('doctor_profile', 'update_own', 'doctor.profile.update_own', 'Update own doctor profile'),
    ('appointments', 'view_own', 'doctor.appointments.view_own', 'View appointments assigned to this doctor'),
    ('appointments', 'create_for_patient', 'doctor.appointments.create_for_patient', 'Book an appointment for a patient'),
    ('appointments', 'cancel_own', 'doctor.appointments.cancel_own', 'Cancel appointments assigned to this doctor'),
    ('appointments', 'reschedule_own', 'doctor.appointments.reschedule_own', 'Reschedule appointments assigned to this doctor'),
    ('patients', 'view', 'doctor.patients.view', 'View patients when managing appointments'),
    ('schedule', 'view_own', 'doctor.schedule.view_own', 'View own schedule'),
    ('schedule', 'define_working_hours_own', 'doctor.schedule.define_working_hours_own', 'Define own recurring working hours'),
    ('schedule', 'block_time_own', 'doctor.schedule.block_time_own', 'Block own time for leave, lunch or other reasons'),
    ('schedule', 'remove_block_own', 'doctor.schedule.remove_block_own', 'Remove own schedule blocks'),
    ('users', 'create', 'admin.users.create', 'Create doctor, patient and admin records'),
    ('users', 'view', 'admin.users.view', 'View doctor, patient and admin records'),
    ('users', 'update_profiles', 'admin.users.update_profiles', 'Update user profiles including own'),
    ('users', 'assign_roles', 'admin.users.assign_roles', 'Assign roles to doctor and admin records'),
    ('roles', 'manage', 'admin.roles.manage', 'Create and update roles'),
    ('permissions', 'manage', 'admin.permissions.manage', 'Manage permissions and role permission sets'),
    ('schedule', 'define_working_hours_any', 'admin.schedule.define_working_hours_any', 'Define recurring working hours for doctors'),
    ('audit', 'view', 'admin.audit.view', 'View audit records'),
    ('appointments', 'view_any', 'admin.appointments.view_any', 'View patient and doctor appointments'),
    ('appointments', 'create_any', 'admin.appointments.create_any', 'Book appointments for patients with doctors'),
    ('appointments', 'cancel_any', 'admin.appointments.cancel_any', 'Cancel patient and doctor appointments'),
    ('appointments', 'reschedule_any', 'admin.appointments.reschedule_any', 'Reschedule patient and doctor appointments')
ON CONFLICT DO NOTHING;

INSERT INTO roles (name, description, is_system) VALUES
    ('doctor', 'Standard doctor permissions', true),
    ('admin', 'Clinic administration permissions', true)
ON CONFLICT DO NOTHING;

INSERT INTO role_scopes (role_id, scope)
SELECT id, name::permission_scope FROM roles WHERE name IN ('doctor', 'admin') AND deleted_at IS NULL
ON CONFLICT (role_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN role_scopes rs ON rs.role_id = r.id
JOIN permissions p ON p.code LIKE rs.scope::text || '.%'
WHERE r.name IN ('doctor', 'admin')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles WHERE name IN ('doctor', 'admin');
DELETE FROM permissions WHERE code LIKE 'doctor.%' OR code LIKE 'admin.%';
