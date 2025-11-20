-- +goose Down
-- Drop triggers
DROP TRIGGER IF EXISTS update_candidates_updated_at ON candidates;
DROP TRIGGER IF EXISTS update_user_accounts_updated_at ON user_accounts;

-- Drop tables
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS user_accounts;

-- Drop enums
DROP TYPE IF EXISTS candidate_status;
DROP TYPE IF EXISTS user_role;
