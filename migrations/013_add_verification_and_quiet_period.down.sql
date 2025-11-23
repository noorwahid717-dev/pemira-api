-- +goose Down
-- Remove verification and quiet period columns from elections

DROP INDEX IF EXISTS idx_elections_quiet_range;
DROP INDEX IF EXISTS idx_elections_verification_range;

ALTER TABLE elections
    DROP COLUMN IF EXISTS quiet_end_at,
    DROP COLUMN IF EXISTS quiet_start_at,
    DROP COLUMN IF EXISTS verification_end_at,
    DROP COLUMN IF EXISTS verification_start_at;
