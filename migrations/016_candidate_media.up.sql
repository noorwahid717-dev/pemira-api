-- +goose Up
-- Candidate media storage (Supabase Storage only) and profile photo reference

CREATE TABLE IF NOT EXISTS candidate_media (
    id UUID PRIMARY KEY,
    candidate_id BIGINT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    slot TEXT NOT NULL CHECK (slot IN ('profile', 'poster', 'photo_extra', 'pdf_program', 'pdf_visimisi')),
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_admin_id BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_candidate_media_candidate ON candidate_media (candidate_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_candidate_media_profile ON candidate_media (candidate_id) WHERE slot = 'profile';

COMMENT ON COLUMN candidate_media.storage_path IS 'Path to media file in Supabase Storage (e.g., candidates/{id}/profile/{uuid}.jpg)';

ALTER TABLE candidates
    ADD COLUMN IF NOT EXISTS photo_media_id UUID NULL REFERENCES candidate_media(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS updated_by_admin_id BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL;
