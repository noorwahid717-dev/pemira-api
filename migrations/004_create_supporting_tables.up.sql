-- =========================================================
-- ENUM TAMBAHAN
-- =========================================================

-- Role user di sistem
CREATE TYPE user_role AS ENUM (
    'ADMIN',
    'PANITIA',
    'KETUA_TPS',
    'OPERATOR_PANEL',
    'VIEWER'
);

-- Status kandidat
CREATE TYPE candidate_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'WITHDRAWN'
);

-- =========================================================
-- TABEL: user_accounts
-- =========================================================

CREATE TABLE user_accounts (
    id              BIGSERIAL PRIMARY KEY,
    
    username        TEXT NOT NULL,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    
    full_name       TEXT NOT NULL,
    role            user_role NOT NULL DEFAULT 'VIEWER',
    
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_user_accounts_username
    ON user_accounts (username);

CREATE UNIQUE INDEX ux_user_accounts_email
    ON user_accounts (email);

CREATE INDEX idx_user_accounts_role
    ON user_accounts (role);

-- =========================================================
-- TABEL: candidates
-- =========================================================

CREATE TABLE candidates (
    id                  BIGSERIAL PRIMARY KEY,
    
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    
    candidate_number    INTEGER NOT NULL,
    chairman_name       TEXT NOT NULL,
    vice_chairman_name  TEXT NOT NULL,
    
    chairman_nim        TEXT NULL,
    vice_chairman_nim   TEXT NULL,
    
    vision              TEXT NULL,
    mission             TEXT NULL,
    photo_url           TEXT NULL,
    
    status              candidate_status NOT NULL DEFAULT 'PENDING',
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_candidates_election_number
    ON candidates (election_id, candidate_number);

CREATE INDEX idx_candidates_election
    ON candidates (election_id);

CREATE INDEX idx_candidates_status
    ON candidates (status);

-- =========================================================
-- TRIGGER: auto-update updated_at
-- =========================================================

CREATE TRIGGER update_user_accounts_updated_at 
    BEFORE UPDATE ON user_accounts
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_candidates_updated_at 
    BEFORE UPDATE ON candidates
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
