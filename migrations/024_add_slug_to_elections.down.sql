-- Migration: Rollback slug column from elections table
-- Date: 2025-11-26

-- Drop index
DROP INDEX IF EXISTS ux_elections_slug;

-- Drop column
ALTER TABLE elections 
DROP COLUMN IF EXISTS slug;
