-- +goose Up
ALTER TABLE voters ADD COLUMN IF NOT EXISTS semester INTEGER NULL;

CREATE INDEX IF NOT EXISTS idx_voters_semester ON voters(semester) WHERE semester IS NOT NULL;
