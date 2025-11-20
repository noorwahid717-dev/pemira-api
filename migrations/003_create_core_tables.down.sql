-- +goose Down
-- Drop triggers
DROP TRIGGER IF EXISTS update_voters_updated_at ON voters;
DROP TRIGGER IF EXISTS update_elections_updated_at ON elections;

-- Drop tables (reverse order)
DROP TABLE IF EXISTS vote_tokens;
DROP TABLE IF EXISTS voters;
DROP TABLE IF EXISTS elections;

-- Drop enums
DROP TYPE IF EXISTS academic_status;
DROP TYPE IF EXISTS election_status;
