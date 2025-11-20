-- +goose Up
-- Update candidates table schema to support comprehensive candidate profiles

-- Drop old columns
ALTER TABLE candidates
    DROP COLUMN IF EXISTS chairman_name,
    DROP COLUMN IF EXISTS vice_chairman_name,
    DROP COLUMN IF EXISTS chairman_nim,
    DROP COLUMN IF EXISTS vice_chairman_nim,
    DROP COLUMN IF EXISTS mission;

-- Rename candidate_number to number
ALTER TABLE candidates
    RENAME COLUMN candidate_number TO number;

-- Add new columns for comprehensive candidate profile
ALTER TABLE candidates
    ADD COLUMN name TEXT NOT NULL DEFAULT '',
    ADD COLUMN short_bio TEXT DEFAULT '',
    ADD COLUMN long_bio TEXT DEFAULT '',
    ADD COLUMN tagline TEXT DEFAULT '',
    ADD COLUMN faculty_name TEXT DEFAULT '',
    ADD COLUMN study_program_name TEXT DEFAULT '',
    ADD COLUMN cohort_year INTEGER,
    ADD COLUMN missions JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN main_programs JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN media JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN social_links JSONB DEFAULT '[]'::jsonb;

-- Remove default from name after adding it
ALTER TABLE candidates ALTER COLUMN name DROP DEFAULT;

-- Update index name to match new column
DROP INDEX IF EXISTS ux_candidates_election_number;
CREATE UNIQUE INDEX ux_candidates_election_number
    ON candidates (election_id, number);

COMMENT ON COLUMN candidates.name IS 'Candidate name or ticket name (e.g., "John Doe - Jane Smith")';
COMMENT ON COLUMN candidates.number IS 'Candidate number for the election (unique per election)';
COMMENT ON COLUMN candidates.missions IS 'Array of mission statements in JSON format';
COMMENT ON COLUMN candidates.main_programs IS 'Array of main program objects with title and description';
COMMENT ON COLUMN candidates.media IS 'Media links object with video_url, instagram, etc.';
COMMENT ON COLUMN candidates.social_links IS 'Array of social media link objects';
