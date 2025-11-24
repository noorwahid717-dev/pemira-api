-- Remove voter_type column from voters table
ALTER TABLE voters 
DROP CONSTRAINT IF EXISTS voters_voter_type_check;

ALTER TABLE voters 
DROP COLUMN IF EXISTS voter_type;
