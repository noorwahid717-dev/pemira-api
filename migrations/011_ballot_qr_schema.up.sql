-- +goose Up
-- Schema for candidate QR codes and TPS ballot scan logging

-- Enum for ballot scan status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ballot_scan_status') THEN
        CREATE TYPE ballot_scan_status AS ENUM ('SCANNED','APPLIED','REJECTED','DUPLICATE');
    END IF;
END$$;

-- Extend tps_checkin_status with VOTED
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tps_checkin_status')
       AND NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'tps_checkin_status' AND e.enumlabel = 'VOTED') THEN
        ALTER TYPE tps_checkin_status ADD VALUE 'VOTED';
    END IF;
END$$;

-- Table of candidate QR definitions
CREATE TABLE IF NOT EXISTS candidate_qr_codes (
    id              BIGSERIAL PRIMARY KEY,
    election_id     BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    candidate_id    BIGINT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    version         INT NOT NULL DEFAULT 1,
    qr_token        TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at      TIMESTAMPTZ NULL
);

-- Unique constraints
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'uq_candidate_qr_active'
    ) THEN
        ALTER TABLE candidate_qr_codes
            ADD CONSTRAINT uq_candidate_qr_active UNIQUE (candidate_id, election_id, is_active) DEFERRABLE INITIALLY IMMEDIATE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'uq_candidate_qr_token'
    ) THEN
        ALTER TABLE candidate_qr_codes
            ADD CONSTRAINT uq_candidate_qr_token UNIQUE (qr_token);
    END IF;
END$$;

-- Ballot scan log table
CREATE TABLE IF NOT EXISTS tps_ballot_scans (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    tps_id              BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    checkin_id          BIGINT NOT NULL REFERENCES tps_checkins(id) ON DELETE CASCADE,
    voter_id            BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    candidate_id        BIGINT NULL REFERENCES candidates(id) ON DELETE SET NULL,
    candidate_qr_id     BIGINT NULL REFERENCES candidate_qr_codes(id) ON DELETE SET NULL,
    raw_payload         TEXT NOT NULL,
    payload_valid       BOOLEAN NOT NULL DEFAULT FALSE,
    status              ballot_scan_status NOT NULL DEFAULT 'SCANNED',
    rejected_reason     TEXT NULL,
    scanned_by_user_id  BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE RESTRICT,
    scanned_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for fast lookup/reporting
CREATE INDEX IF NOT EXISTS idx_ballot_scans_election ON tps_ballot_scans (election_id);
CREATE INDEX IF NOT EXISTS idx_ballot_scans_tps ON tps_ballot_scans (tps_id);
CREATE INDEX IF NOT EXISTS idx_ballot_scans_checkin ON tps_ballot_scans (checkin_id);
CREATE INDEX IF NOT EXISTS idx_ballot_scans_voter ON tps_ballot_scans (voter_id);
CREATE INDEX IF NOT EXISTS idx_ballot_scans_status ON tps_ballot_scans (status);

-- Add metadata columns to votes
ALTER TABLE votes
    ADD COLUMN IF NOT EXISTS candidate_qr_id BIGINT NULL REFERENCES candidate_qr_codes(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS ballot_scan_id BIGINT NULL REFERENCES tps_ballot_scans(id) ON DELETE SET NULL;

-- Add voted_at column on tps_checkins
ALTER TABLE tps_checkins
    ADD COLUMN IF NOT EXISTS voted_at TIMESTAMPTZ NULL;
