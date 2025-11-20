-- +goose Up
-- =========================================================
-- ENUM TAMBAHAN
-- =========================================================

-- Status pemilu
CREATE TYPE election_status AS ENUM (
    'DRAFT',
    'REGISTRATION',
    'CAMPAIGN',
    'VOTING_OPEN',
    'CLOSED',
    'ARCHIVED'
);

-- Status akademik mahasiswa
CREATE TYPE academic_status AS ENUM (
    'ACTIVE',
    'GRADUATED',
    'ON_LEAVE',
    'DROPPED',
    'INACTIVE'
);

-- =========================================================
-- TABEL: elections
-- =========================================================

CREATE TABLE elections (
    id                      BIGSERIAL PRIMARY KEY,
    code                    TEXT NOT NULL,
    name                    TEXT NOT NULL,
    year                    INTEGER NOT NULL,

    status                  election_status NOT NULL DEFAULT 'DRAFT',

    online_enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    tps_enabled             BOOLEAN NOT NULL DEFAULT TRUE,

    registration_start_at   TIMESTAMPTZ NULL,
    registration_end_at     TIMESTAMPTZ NULL,

    campaign_start_at       TIMESTAMPTZ NULL,
    campaign_end_at         TIMESTAMPTZ NULL,

    voting_start_at         TIMESTAMPTZ NULL,
    voting_end_at           TIMESTAMPTZ NULL,

    recap_start_at          TIMESTAMPTZ NULL,
    recap_end_at            TIMESTAMPTZ NULL,

    description             TEXT NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_election_mode_enabled
        CHECK (online_enabled = TRUE OR tps_enabled = TRUE)
);

CREATE UNIQUE INDEX ux_elections_code
    ON elections (code);

CREATE INDEX idx_elections_year
    ON elections (year);

CREATE INDEX idx_elections_status
    ON elections (status);

-- =========================================================
-- TABEL: voters
-- =========================================================

CREATE TABLE voters (
    id                      BIGSERIAL PRIMARY KEY,

    nim                     TEXT NOT NULL,
    name                    TEXT NOT NULL,
    email                   TEXT NULL,

    faculty_code            TEXT NULL,
    faculty_name            TEXT NULL,

    study_program_code      TEXT NULL,
    study_program_name      TEXT NULL,

    cohort_year             INTEGER NULL,
    class_label             TEXT NULL,

    academic_status         academic_status NOT NULL DEFAULT 'ACTIVE',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_voters_nim
    ON voters (nim);

CREATE INDEX idx_voters_faculty
    ON voters (faculty_code, study_program_code);

CREATE INDEX idx_voters_cohort
    ON voters (cohort_year);

CREATE INDEX idx_voters_academic_status
    ON voters (academic_status);

-- =========================================================
-- TABEL: vote_tokens
-- =========================================================

CREATE TABLE vote_tokens (
    id              BIGSERIAL PRIMARY KEY,

    election_id     BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    voter_id        BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,

    token_hash      TEXT NOT NULL,

    method          voting_method NOT NULL,
    tps_id          BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,

    issued_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at         TIMESTAMPTZ NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_vote_tokens_token_hash
    ON vote_tokens (token_hash);

CREATE UNIQUE INDEX ux_vote_tokens_election_voter
    ON vote_tokens (election_id, voter_id);

CREATE INDEX idx_vote_tokens_election_used
    ON vote_tokens (election_id, used_at);

CREATE INDEX idx_vote_tokens_tps
    ON vote_tokens (tps_id);

-- =========================================================
-- TRIGGER: auto-update updated_at
-- =========================================================

CREATE TRIGGER update_elections_updated_at 
    BEFORE UPDATE ON elections
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_voters_updated_at 
    BEFORE UPDATE ON voters
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
