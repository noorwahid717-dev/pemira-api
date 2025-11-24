-- +goose Up
-- Branding assets for election admin logos (Supabase Storage only)

CREATE TABLE IF NOT EXISTS branding_files (
    id UUID PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    slot TEXT NOT NULL CHECK (slot IN ('primary', 'secondary')),
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_admin_id BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_branding_files_election_slot
    ON branding_files (election_id, slot);

COMMENT ON COLUMN branding_files.storage_path IS 'Path to branding file in Supabase Storage (e.g., elections/{id}/branding/{slot}/{uuid}.png)';

CREATE TABLE IF NOT EXISTS branding_settings (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL UNIQUE REFERENCES elections(id) ON DELETE CASCADE,
    primary_logo_id UUID NULL REFERENCES branding_files(id) ON DELETE SET NULL,
    secondary_logo_id UUID NULL REFERENCES branding_files(id) ON DELETE SET NULL,
    updated_by_admin_id BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
