-- +goose Up
-- Extend elections table for admin endpoints (phases, mode settings, academic year)

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'election_status') THEN
        BEGIN
            ALTER TYPE election_status ADD VALUE IF NOT EXISTS 'VERIFICATION';
        EXCEPTION WHEN duplicate_object THEN NULL;
        END;
        BEGIN
            ALTER TYPE election_status ADD VALUE IF NOT EXISTS 'QUIET_PERIOD';
        EXCEPTION WHEN duplicate_object THEN NULL;
        END;
        BEGIN
            ALTER TYPE election_status ADD VALUE IF NOT EXISTS 'VOTING_CLOSED';
        EXCEPTION WHEN duplicate_object THEN NULL;
        END;
        BEGIN
            ALTER TYPE election_status ADD VALUE IF NOT EXISTS 'RECAP';
        EXCEPTION WHEN duplicate_object THEN NULL;
        END;
    END IF;
END$$;

ALTER TABLE elections
    ADD COLUMN IF NOT EXISTS academic_year TEXT NULL,
    ADD COLUMN IF NOT EXISTS current_phase TEXT NULL,
    ADD COLUMN IF NOT EXISTS online_login_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS online_max_sessions_per_voter INT NULL,
    ADD COLUMN IF NOT EXISTS tps_require_checkin BOOLEAN NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS tps_require_ballot_qr BOOLEAN NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS tps_max INT NULL;
