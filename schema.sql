-- PEMIRA Production Database Schema
-- Extracted from backup: pemira_production_backup_20251209_165647.sql
-- Schema: myschema

-- Create schema
CREATE SCHEMA IF NOT EXISTS myschema;

-- Set search path
SET search_path TO myschema, public;

-- ENUMS
CREATE TYPE myschema.academic_status AS ENUM (
    'ACTIVE',
    'GRADUATED',
    'ON_LEAVE',
    'DROPPED',
    'INACTIVE'
);

CREATE TYPE myschema.ballot_scan_status AS ENUM (
    'SCANNED',
    'APPLIED',
    'REJECTED',
    'DUPLICATE'
);

CREATE TYPE myschema.candidate_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'WITHDRAWN'
);

CREATE TYPE myschema.election_status AS ENUM (
    'DRAFT',
    'REGISTRATION',
    'CAMPAIGN',
    'VOTING_OPEN',
    'CLOSED',
    'ARCHIVED',
    'VERIFICATION',
    'QUIET_PERIOD',
    'VOTING_CLOSED',
    'RECAP'
);

CREATE TYPE myschema.election_voter_status AS ENUM (
    'PENDING',
    'VERIFIED',
    'REJECTED',
    'BLOCKED'
);

CREATE TYPE myschema.voting_channel AS ENUM (
    'ONLINE',
    'TPS'
);

CREATE TYPE myschema.voter_type AS ENUM (
    'STUDENT',
    'LECTURER',
    'STAFF'
);

CREATE TYPE myschema.voting_method AS ENUM (
    'ONLINE',
    'TPS',
    'HYBRID'
);

-- Reference: This schema is restored from production backup
-- All tables, indexes, and constraints are included in the backup file
-- Use restore_db.sh to restore the complete database structure and data
