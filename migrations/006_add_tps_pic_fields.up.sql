-- +goose Up
-- =========================================================
-- Add PIC fields to TPS table
-- =========================================================

ALTER TABLE tps
    ADD COLUMN pic_name    TEXT NULL,
    ADD COLUMN pic_phone   TEXT NULL,
    ADD COLUMN notes       TEXT NULL;

-- =========================================================
-- Update tps_qr table structure
-- =========================================================

-- Rename qr_secret to qr_token for consistency
ALTER TABLE tps_qr RENAME COLUMN qr_secret TO qr_token;

-- Rename revoked_at to rotated_at for clarity
ALTER TABLE tps_qr RENAME COLUMN revoked_at TO rotated_at;

-- Update index name to match new column name
DROP INDEX IF EXISTS ux_tps_qr_secret;
CREATE UNIQUE INDEX ux_tps_qr_token ON tps_qr (qr_token);

-- Add constraint to ensure only one active QR per TPS
CREATE UNIQUE INDEX ux_tps_active_qr 
    ON tps_qr (tps_id) 
    WHERE is_active = TRUE;

COMMENT ON COLUMN tps.pic_name IS 'Penanggung jawab TPS';
COMMENT ON COLUMN tps.pic_phone IS 'Kontak panitia TPS';
COMMENT ON COLUMN tps.notes IS 'Catatan internal (opsional)';
COMMENT ON COLUMN tps_qr.qr_token IS 'Random token untuk di-embed ke QR';
COMMENT ON COLUMN tps_qr.rotated_at IS 'Timestamp ketika QR di-rotate';
