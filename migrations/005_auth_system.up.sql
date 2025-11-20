-- +goose Up
-- Auth System Migration
-- Adds user_accounts and user_sessions tables for JWT-based authentication

-- User role enum
CREATE TYPE user_role AS ENUM ('STUDENT', 'ADMIN', 'TPS_OPERATOR', 'SUPER_ADMIN');

-- User accounts table
CREATE TABLE user_accounts (
    id              BIGSERIAL PRIMARY KEY,
    username        TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    role            user_role NOT NULL,
    
    -- Optional linkage to other entities
    voter_id        BIGINT NULL REFERENCES voters(id) ON DELETE SET NULL,
    tps_id          BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,
    
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Username must be unique
CREATE UNIQUE INDEX ux_user_accounts_username
    ON user_accounts (username);

-- Index for lookups
CREATE INDEX idx_user_accounts_role
    ON user_accounts (role);

CREATE INDEX idx_user_accounts_voter_id
    ON user_accounts (voter_id)
    WHERE voter_id IS NOT NULL;

CREATE INDEX idx_user_accounts_tps_id
    ON user_accounts (tps_id)
    WHERE tps_id IS NOT NULL;

-- User sessions table for refresh tokens
CREATE TABLE user_sessions (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    
    refresh_token_hash   TEXT NOT NULL,
    user_agent           TEXT NULL,
    ip_address           INET NULL,
    
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at           TIMESTAMPTZ NOT NULL,
    revoked_at           TIMESTAMPTZ NULL
);

-- Index for user's sessions
CREATE INDEX idx_user_sessions_user
    ON user_sessions (user_id);

-- Unique constraint on refresh token hash
CREATE UNIQUE INDEX ux_user_sessions_token_hash
    ON user_sessions (refresh_token_hash)
    WHERE revoked_at IS NULL;

-- Index for cleanup of expired sessions
CREATE INDEX idx_user_sessions_expires_at
    ON user_sessions (expires_at)
    WHERE revoked_at IS NULL;

-- Comments
COMMENT ON TABLE user_accounts IS 'User accounts for all roles (STUDENT, ADMIN, TPS_OPERATOR)';
COMMENT ON TABLE user_sessions IS 'User sessions for refresh token management';
COMMENT ON COLUMN user_accounts.username IS 'Username (NIM for students, email/username for others)';
COMMENT ON COLUMN user_accounts.voter_id IS 'Link to voters table for STUDENT role';
COMMENT ON COLUMN user_accounts.tps_id IS 'Link to TPS for TPS_OPERATOR role';
COMMENT ON COLUMN user_sessions.refresh_token_hash IS 'Hashed refresh token (DO NOT store plain token)';
