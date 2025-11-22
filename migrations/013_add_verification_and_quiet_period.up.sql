-- +goose Up
-- Extend elections with verification and quiet period windows
-- Adds nullable columns so existing data remains valid

ALTER TABLE elections
    ADD COLUMN IF NOT EXISTS verification_start_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS verification_end_at   TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS quiet_start_at        TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS quiet_end_at          TIMESTAMPTZ NULL;

-- Helpful indexes for date range queries (idempotent)
CREATE INDEX IF NOT EXISTS idx_elections_verification_range ON elections (verification_start_at, verification_end_at);
CREATE INDEX IF NOT EXISTS idx_elections_quiet_range ON elections (quiet_start_at, quiet_end_at);
