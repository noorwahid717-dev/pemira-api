-- +goose Up
-- Adds announcement_at and finished_at columns to elections

ALTER TABLE elections
    ADD COLUMN IF NOT EXISTS announcement_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS finished_at     TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_elections_announcement_at ON elections (announcement_at);
CREATE INDEX IF NOT EXISTS idx_elections_finished_at ON elections (finished_at);
