-- Migration: Add slug column to elections table
-- Date: 2025-11-26
-- Description: Add optional slug field for URL-friendly election identifiers

-- Add slug column
ALTER TABLE elections 
ADD COLUMN IF NOT EXISTS slug TEXT;

-- Create unique index on slug (only for non-null values)
CREATE UNIQUE INDEX IF NOT EXISTS ux_elections_slug 
ON elections(slug) 
WHERE slug IS NOT NULL;

-- Add comment
COMMENT ON COLUMN elections.slug IS 'URL-friendly identifier for election (optional)';
