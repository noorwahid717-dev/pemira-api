-- +goose Down
DROP INDEX IF EXISTS idx_voters_semester;

ALTER TABLE voters DROP COLUMN IF EXISTS semester;
