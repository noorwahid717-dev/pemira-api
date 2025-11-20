-- +goose Down
-- Rollback lecturer and staff support

-- Drop indexes
DROP INDEX IF EXISTS idx_user_accounts_staff_id;
DROP INDEX IF EXISTS idx_user_accounts_lecturer_id;
DROP INDEX IF EXISTS idx_staff_members_unit;
DROP INDEX IF EXISTS ux_staff_members_nip;
DROP INDEX IF EXISTS idx_lecturers_department;
DROP INDEX IF EXISTS idx_lecturers_faculty;
DROP INDEX IF EXISTS ux_lecturers_nidn;

-- Drop columns from user_accounts
ALTER TABLE user_accounts
    DROP COLUMN IF EXISTS staff_id,
    DROP COLUMN IF EXISTS lecturer_id;

-- Drop tables
DROP TABLE IF EXISTS staff_members;
DROP TABLE IF EXISTS lecturers;

-- Note: Cannot remove values from enum type in PostgreSQL
-- user_role enum will still contain LECTURER and STAFF values
-- To fully remove, would need to recreate the enum type
