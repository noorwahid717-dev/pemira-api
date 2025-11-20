-- +goose Down
-- Rollback candidates table schema changes

-- Remove new columns
ALTER TABLE candidates
    DROP COLUMN IF EXISTS social_links,
    DROP COLUMN IF EXISTS media,
    DROP COLUMN IF EXISTS main_programs,
    DROP COLUMN IF EXISTS missions,
    DROP COLUMN IF EXISTS cohort_year,
    DROP COLUMN IF EXISTS study_program_name,
    DROP COLUMN IF EXISTS faculty_name,
    DROP COLUMN IF EXISTS tagline,
    DROP COLUMN IF EXISTS long_bio,
    DROP COLUMN IF EXISTS short_bio,
    DROP COLUMN IF EXISTS name;

-- Rename number back to candidate_number
ALTER TABLE candidates
    RENAME COLUMN number TO candidate_number;

-- Add back old columns
ALTER TABLE candidates
    ADD COLUMN chairman_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN vice_chairman_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN chairman_nim TEXT,
    ADD COLUMN vice_chairman_nim TEXT,
    ADD COLUMN mission TEXT;

-- Remove defaults
ALTER TABLE candidates 
    ALTER COLUMN chairman_name DROP DEFAULT,
    ALTER COLUMN vice_chairman_name DROP DEFAULT;

-- Recreate old index
DROP INDEX IF EXISTS ux_candidates_election_number;
CREATE UNIQUE INDEX ux_candidates_election_number
    ON candidates (election_id, candidate_number);
