-- +goose Up
-- Allow editing voting_method anytime and store preference on voters

-- Ensure enum exists (safety)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'voting_method') THEN
        CREATE TYPE voting_method AS ENUM ('ONLINE','TPS');
    END IF;
END$$;

ALTER TABLE voters
    ADD COLUMN IF NOT EXISTS voting_method voting_method NULL DEFAULT 'ONLINE';

-- Relax constraint to allow voting_method to be set before voting
ALTER TABLE voter_status
    DROP CONSTRAINT IF EXISTS chk_voter_status_method_has_voted;

ALTER TABLE voter_status
    ADD CONSTRAINT chk_voter_status_method_has_voted
    CHECK (
        (has_voted = FALSE AND voted_at IS NULL)
     OR (has_voted = TRUE AND voting_method IS NOT NULL AND voted_at IS NOT NULL)
    );
