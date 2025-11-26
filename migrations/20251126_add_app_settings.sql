-- Migration: Add app settings table
-- Created: 2025-11-26

CREATE TABLE IF NOT EXISTS app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by INT REFERENCES user_accounts(id)
);

-- Insert default active election
INSERT INTO app_settings (key, value, description)
VALUES 
    ('active_election_id', '1', 'ID election yang aktif saat ini untuk admin dashboard'),
    ('default_election_id', '1', 'ID election default untuk voter')
ON CONFLICT (key) DO NOTHING;

-- Create index
CREATE INDEX IF NOT EXISTS idx_app_settings_key ON app_settings(key);
