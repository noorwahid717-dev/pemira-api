-- Add voter_type column to voters table
ALTER TABLE voters 
ADD COLUMN IF NOT EXISTS voter_type TEXT DEFAULT 'STUDENT';

-- Add check constraint for valid voter types
ALTER TABLE voters 
ADD CONSTRAINT voters_voter_type_check 
CHECK (voter_type IN ('STUDENT', 'LECTURER', 'STAFF'));

-- Update existing voters: set STUDENT as default
UPDATE voters 
SET voter_type = 'STUDENT' 
WHERE voter_type IS NULL;

-- Sync voter_type from user_accounts for voters with accounts
UPDATE voters v
SET voter_type = ua.role::TEXT
FROM user_accounts ua
WHERE ua.voter_id = v.id
  AND ua.role IN ('STUDENT', 'LECTURER', 'STAFF');
