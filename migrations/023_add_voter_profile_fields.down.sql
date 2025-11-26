-- Rollback: Remove profile fields from voters and users tables

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_voters_updated_at ON voters;
DROP FUNCTION IF EXISTS update_voters_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_voters_email;
DROP INDEX IF EXISTS idx_voters_updated_at;
DROP INDEX IF EXISTS idx_users_last_login;

-- Remove columns from users table
ALTER TABLE users
DROP COLUMN IF EXISTS last_login_at,
DROP COLUMN IF EXISTS login_count;

-- Remove columns from voters table
ALTER TABLE voters
DROP COLUMN IF EXISTS email,
DROP COLUMN IF EXISTS phone,
DROP COLUMN IF EXISTS photo_url,
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF NOT EXISTS voting_method_preference,
DROP COLUMN IF EXISTS updated_at;
