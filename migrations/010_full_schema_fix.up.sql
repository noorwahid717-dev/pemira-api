-- +goose Up
-- Consolidated idempotent migration to ensure full schema exists

-- Enum helpers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'election_status') THEN
        CREATE TYPE election_status AS ENUM ('DRAFT','REGISTRATION','CAMPAIGN','VOTING_OPEN','CLOSED','ARCHIVED');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'academic_status') THEN
        CREATE TYPE academic_status AS ENUM ('ACTIVE','GRADUATED','ON_LEAVE','DROPPED','INACTIVE');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tps_status') THEN
        CREATE TYPE tps_status AS ENUM ('DRAFT','ACTIVE','CLOSED');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tps_checkin_status') THEN
        CREATE TYPE tps_checkin_status AS ENUM ('PENDING','APPROVED','REJECTED','USED','EXPIRED');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'voting_method') THEN
        CREATE TYPE voting_method AS ENUM ('ONLINE','TPS');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vote_channel') THEN
        CREATE TYPE vote_channel AS ENUM ('ONLINE','TPS');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('ADMIN','PANITIA','KETUA_TPS','OPERATOR_PANEL','VIEWER');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'user_role' AND e.enumlabel = 'STUDENT') THEN
        ALTER TYPE user_role ADD VALUE 'STUDENT';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'user_role' AND e.enumlabel = 'TPS_OPERATOR') THEN
        ALTER TYPE user_role ADD VALUE 'TPS_OPERATOR';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'user_role' AND e.enumlabel = 'SUPER_ADMIN') THEN
        ALTER TYPE user_role ADD VALUE 'SUPER_ADMIN';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'user_role' AND e.enumlabel = 'LECTURER') THEN
        ALTER TYPE user_role ADD VALUE 'LECTURER';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'user_role' AND e.enumlabel = 'STAFF') THEN
        ALTER TYPE user_role ADD VALUE 'STAFF';
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'candidate_status') THEN
        CREATE TYPE candidate_status AS ENUM ('PENDING','APPROVED','REJECTED','WITHDRAWN');
    END IF;
END$$;

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- elections
CREATE TABLE IF NOT EXISTS elections (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    year INTEGER NOT NULL,
    status election_status NOT NULL DEFAULT 'DRAFT',
    online_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    tps_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    registration_start_at TIMESTAMPTZ NULL,
    registration_end_at TIMESTAMPTZ NULL,
    campaign_start_at TIMESTAMPTZ NULL,
    campaign_end_at TIMESTAMPTZ NULL,
    voting_start_at TIMESTAMPTZ NULL,
    voting_end_at TIMESTAMPTZ NULL,
    recap_start_at TIMESTAMPTZ NULL,
    recap_end_at TIMESTAMPTZ NULL,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_election_mode_enabled CHECK (online_enabled = TRUE OR tps_enabled = TRUE)
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_elections_code ON elections (code);
CREATE INDEX IF NOT EXISTS idx_elections_year ON elections (year);
CREATE INDEX IF NOT EXISTS idx_elections_status ON elections (status);
DROP TRIGGER IF EXISTS update_elections_updated_at ON elections;
CREATE TRIGGER update_elections_updated_at BEFORE UPDATE ON elections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- voters
CREATE TABLE IF NOT EXISTS voters (
    id BIGSERIAL PRIMARY KEY,
    nim TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT NULL,
    faculty_code TEXT NULL,
    faculty_name TEXT NULL,
    study_program_code TEXT NULL,
    study_program_name TEXT NULL,
    cohort_year INTEGER NULL,
    class_label TEXT NULL,
    academic_status academic_status NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_voters_nim ON voters (nim);
CREATE INDEX IF NOT EXISTS idx_voters_faculty ON voters (faculty_code, study_program_code);
CREATE INDEX IF NOT EXISTS idx_voters_cohort ON voters (cohort_year);
CREATE INDEX IF NOT EXISTS idx_voters_academic_status ON voters (academic_status);
DROP TRIGGER IF EXISTS update_voters_updated_at ON voters;
CREATE TRIGGER update_voters_updated_at BEFORE UPDATE ON voters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- tps
CREATE TABLE IF NOT EXISTS tps (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    status tps_status NOT NULL DEFAULT 'DRAFT',
    voting_date DATE NOT NULL,
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    capacity_estimate INTEGER,
    area_faculty_id BIGINT NULL,
    pic_name TEXT NULL,
    pic_phone TEXT NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_tps_election_code ON tps (election_id, code);
CREATE INDEX IF NOT EXISTS idx_tps_election ON tps (election_id);
CREATE INDEX IF NOT EXISTS idx_tps_status ON tps (status);
DROP TRIGGER IF EXISTS update_tps_updated_at ON tps;
CREATE TRIGGER update_tps_updated_at BEFORE UPDATE ON tps FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- tps_qr
CREATE TABLE IF NOT EXISTS tps_qr (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    qr_token TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_tps_qr_token ON tps_qr (qr_token);
CREATE UNIQUE INDEX IF NOT EXISTS ux_tps_active_qr ON tps_qr (tps_id) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_tps_qr_active_tps ON tps_qr (tps_id, is_active);
COMMENT ON COLUMN tps_qr.qr_token IS 'Random token for QR';

-- vote_tokens
CREATE TABLE IF NOT EXISTS vote_tokens (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    voter_id BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    method voting_method NOT NULL,
    tps_id BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_vote_tokens_token_hash ON vote_tokens (token_hash);
CREATE UNIQUE INDEX IF NOT EXISTS ux_vote_tokens_election_voter ON vote_tokens (election_id, voter_id);
CREATE INDEX IF NOT EXISTS idx_vote_tokens_election_used ON vote_tokens (election_id, used_at);
CREATE INDEX IF NOT EXISTS idx_vote_tokens_tps ON vote_tokens (tps_id);

-- user_accounts
CREATE TABLE IF NOT EXISTS user_accounts (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'VIEWER',
    voter_id BIGINT NULL,
    tps_id BIGINT NULL,
    lecturer_id BIGINT NULL,
    staff_id BIGINT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Ensure missing columns exist
ALTER TABLE user_accounts
    ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS full_name TEXT NOT NULL DEFAULT '';
ALTER TABLE user_accounts
    ALTER COLUMN email DROP DEFAULT,
    ALTER COLUMN full_name DROP DEFAULT;

ALTER TABLE user_accounts
    ADD COLUMN IF NOT EXISTS voter_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS tps_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS lecturer_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS staff_id BIGINT NULL;
-- Foreign keys (add if missing)
ALTER TABLE user_accounts
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN full_name SET NOT NULL,
    ALTER COLUMN role SET NOT NULL,
    ALTER COLUMN is_active SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_accounts_voter') THEN
        ALTER TABLE user_accounts
            ADD CONSTRAINT fk_user_accounts_voter FOREIGN KEY (voter_id) REFERENCES voters(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_accounts_tps') THEN
        ALTER TABLE user_accounts
            ADD CONSTRAINT fk_user_accounts_tps FOREIGN KEY (tps_id) REFERENCES tps(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_accounts_lecturer') THEN
        ALTER TABLE user_accounts
            ADD CONSTRAINT fk_user_accounts_lecturer FOREIGN KEY (lecturer_id) REFERENCES lecturers(id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_accounts_staff') THEN
        ALTER TABLE user_accounts
            ADD CONSTRAINT fk_user_accounts_staff FOREIGN KEY (staff_id) REFERENCES staff_members(id) ON DELETE SET NULL;
    END IF;
END$$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_user_accounts_username ON user_accounts (username);
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_accounts_email ON user_accounts (email);
CREATE INDEX IF NOT EXISTS idx_user_accounts_role ON user_accounts (role);
CREATE INDEX IF NOT EXISTS idx_user_accounts_voter_id ON user_accounts (voter_id) WHERE voter_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_accounts_tps_id ON user_accounts (tps_id) WHERE tps_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_accounts_lecturer_id ON user_accounts (lecturer_id) WHERE lecturer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_accounts_staff_id ON user_accounts (staff_id) WHERE staff_id IS NOT NULL;
DROP TRIGGER IF EXISTS update_user_accounts_updated_at ON user_accounts;
CREATE TRIGGER update_user_accounts_updated_at BEFORE UPDATE ON user_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- lecturers
CREATE TABLE IF NOT EXISTS lecturers (
    id BIGSERIAL PRIMARY KEY,
    nidn TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT NULL,
    faculty_code TEXT NULL,
    faculty_name TEXT NULL,
    department_code TEXT NULL,
    department_name TEXT NULL,
    position TEXT NULL,
    employment_status TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_lecturers_nidn ON lecturers (nidn);
CREATE INDEX IF NOT EXISTS idx_lecturers_faculty ON lecturers (faculty_code);
CREATE INDEX IF NOT EXISTS idx_lecturers_department ON lecturers (department_code);

-- staff_members
CREATE TABLE IF NOT EXISTS staff_members (
    id BIGSERIAL PRIMARY KEY,
    nip TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT NULL,
    unit_code TEXT NULL,
    unit_name TEXT NULL,
    position TEXT NULL,
    employment_status TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_staff_members_nip ON staff_members (nip);
CREATE INDEX IF NOT EXISTS idx_staff_members_unit ON staff_members (unit_code);

-- user_sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT NULL,
    ip_address INET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_sessions_token_hash ON user_sessions (refresh_token_hash) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions (expires_at) WHERE revoked_at IS NULL;

-- candidates (final schema)
CREATE TABLE IF NOT EXISTS candidates (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    name TEXT NOT NULL,
    photo_url TEXT,
    short_bio TEXT DEFAULT '',
    long_bio TEXT DEFAULT '',
    tagline TEXT DEFAULT '',
    faculty_name TEXT DEFAULT '',
    study_program_name TEXT DEFAULT '',
    cohort_year INTEGER,
    vision TEXT,
    missions JSONB DEFAULT '[]'::jsonb,
    main_programs JSONB DEFAULT '[]'::jsonb,
    media JSONB DEFAULT '{}'::jsonb,
    social_links JSONB DEFAULT '[]'::jsonb,
    status candidate_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_candidates_election_number ON candidates (election_id, number);
CREATE INDEX IF NOT EXISTS idx_candidates_election ON candidates (election_id);
CREATE INDEX IF NOT EXISTS idx_candidates_status ON candidates (status);
DROP TRIGGER IF EXISTS update_candidates_updated_at ON candidates;
CREATE TRIGGER update_candidates_updated_at BEFORE UPDATE ON candidates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- tps_checkins
CREATE TABLE IF NOT EXISTS tps_checkins (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    tps_id BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    qr_id BIGINT NOT NULL REFERENCES tps_qr(id) ON DELETE RESTRICT,
    voter_id BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    status tps_checkin_status NOT NULL DEFAULT 'PENDING',
    scan_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ NULL,
    approved_by_id BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL,
    rejection_reason TEXT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tps_checkins_tps_status_scan ON tps_checkins (tps_id, status, scan_at);
CREATE INDEX IF NOT EXISTS idx_tps_checkins_election_voter_status ON tps_checkins (election_id, voter_id, status, scan_at DESC);

-- voter_status
CREATE TABLE IF NOT EXISTS voter_status (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    voter_id BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    is_eligible BOOLEAN NOT NULL DEFAULT TRUE,
    has_voted BOOLEAN NOT NULL DEFAULT FALSE,
    voting_method voting_method NULL,
    tps_id BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,
    voted_at TIMESTAMPTZ NULL,
    vote_token_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_voter_status_method_has_voted CHECK (
        (has_voted = FALSE AND voting_method IS NULL AND tps_id IS NULL AND voted_at IS NULL)
        OR (has_voted = TRUE AND voting_method IS NOT NULL AND voted_at IS NOT NULL)
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_voter_status_election_voter ON voter_status (election_id, voter_id);
CREATE INDEX IF NOT EXISTS idx_voter_status_election_has_voted ON voter_status (election_id, has_voted);
CREATE UNIQUE INDEX IF NOT EXISTS ux_voter_status_token_hash ON voter_status (vote_token_hash) WHERE vote_token_hash IS NOT NULL;
DROP TRIGGER IF EXISTS update_voter_status_updated_at ON voter_status;
CREATE TRIGGER update_voter_status_updated_at BEFORE UPDATE ON voter_status FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- votes
CREATE TABLE IF NOT EXISTS votes (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    candidate_id BIGINT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    channel vote_channel NOT NULL,
    tps_id BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,
    cast_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_votes_token_hash ON votes (token_hash);
CREATE INDEX IF NOT EXISTS idx_votes_election_candidate ON votes (election_id, candidate_id);
CREATE INDEX IF NOT EXISTS idx_votes_election_channel ON votes (election_id, channel);
CREATE INDEX IF NOT EXISTS idx_votes_election_tps ON votes (election_id, tps_id);
