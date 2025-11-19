-- Rollback auth system migration

DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_accounts;
DROP TYPE IF EXISTS user_role;
