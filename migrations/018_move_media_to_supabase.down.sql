-- +goose Down
-- Revert Supabase Storage changes

-- Remove storage_path from candidate_media
ALTER TABLE candidate_media 
DROP COLUMN IF EXISTS storage_path;

-- Make data NOT NULL again (requires data to exist)
-- Note: This will fail if there are records with NULL data
ALTER TABLE candidate_media 
ALTER COLUMN data SET NOT NULL;

-- Remove storage_path from branding_files
ALTER TABLE branding_files 
DROP COLUMN IF EXISTS storage_path;

-- Make data NOT NULL again for branding files
ALTER TABLE branding_files 
ALTER COLUMN data SET NOT NULL;
