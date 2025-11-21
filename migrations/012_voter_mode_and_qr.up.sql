-- +goose Up
-- Add preferred voting method flags to voter_status and QR for TPS voter registration

-- Ensure voting_method enum exists (should be created earlier)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'voting_method') THEN
        CREATE TYPE voting_method AS ENUM ('ONLINE','TPS');
    END IF;
END$$;

-- Add columns to voter_status
ALTER TABLE voter_status
    ADD COLUMN IF NOT EXISTS preferred_method voting_method NULL,
    ADD COLUMN IF NOT EXISTS online_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS tps_allowed BOOLEAN NOT NULL DEFAULT TRUE;

-- Create QR table for TPS voter registration
CREATE TABLE IF NOT EXISTS voter_tps_qr (
    id              BIGSERIAL PRIMARY KEY,
    voter_id        BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    election_id     BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    qr_token        TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    rotated_at      TIMESTAMPTZ NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_voter_tps_qr_active'
    ) THEN
        ALTER TABLE voter_tps_qr
            ADD CONSTRAINT uq_voter_tps_qr_active UNIQUE (voter_id, election_id, is_active) DEFERRABLE INITIALLY IMMEDIATE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_voter_tps_qr_token'
    ) THEN
        ALTER TABLE voter_tps_qr
            ADD CONSTRAINT uq_voter_tps_qr_token UNIQUE (qr_token);
    END IF;
END$$;
