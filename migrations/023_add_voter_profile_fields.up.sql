-- Migration: Add profile fields to voters and users tables
-- Date: 2025-11-25

-- Add profile fields to voters table
ALTER TABLE voters
ADD COLUMN IF NOT EXISTS email VARCHAR(255),
ADD COLUMN IF NOT EXISTS phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS photo_url TEXT,
ADD COLUMN IF NOT EXISTS bio TEXT,
ADD COLUMN IF NOT EXISTS voting_method_preference VARCHAR(20) DEFAULT 'ONLINE',
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add index for email lookups
CREATE INDEX IF NOT EXISTS idx_voters_email ON voters(email);
CREATE INDEX IF NOT EXISTS idx_voters_updated_at ON voters(updated_at);

-- Add login tracking fields to users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS login_count INTEGER DEFAULT 0;

-- Add index for last_login_at
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login_at);

-- Update existing records to have default voting_method_preference
UPDATE voters 
SET voting_method_preference = 'ONLINE' 
WHERE voting_method_preference IS NULL;

-- Add trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_voters_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_voters_updated_at ON voters;
CREATE TRIGGER trigger_voters_updated_at
    BEFORE UPDATE ON voters
    FOR EACH ROW
    EXECUTE FUNCTION update_voters_updated_at();

COMMENT ON COLUMN voters.email IS 'Voter email address (optional, can be updated)';
COMMENT ON COLUMN voters.phone IS 'Voter phone number (optional)';
COMMENT ON COLUMN voters.photo_url IS 'Profile photo URL from storage';
COMMENT ON COLUMN voters.bio IS 'Short biography or description';
COMMENT ON COLUMN voters.voting_method_preference IS 'Preferred voting method: ONLINE or TPS';
COMMENT ON COLUMN users.last_login_at IS 'Last successful login timestamp';
COMMENT ON COLUMN users.login_count IS 'Total number of successful logins';
