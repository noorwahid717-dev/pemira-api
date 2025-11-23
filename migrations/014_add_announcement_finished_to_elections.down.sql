-- +goose Down
-- Remove announcement_at and finished_at columns from elections

DROP INDEX IF EXISTS idx_elections_finished_at;
DROP INDEX IF EXISTS idx_elections_announcement_at;

ALTER TABLE elections
    DROP COLUMN IF EXISTS finished_at,
    DROP COLUMN IF EXISTS announcement_at;
