-- =========================================================
-- ENUMS
-- =========================================================

-- Status TPS
CREATE TYPE tps_status AS ENUM ('DRAFT', 'ACTIVE', 'CLOSED');

-- Status check-in TPS
CREATE TYPE tps_checkin_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'USED', 'EXPIRED');

-- Metode voting
CREATE TYPE voting_method AS ENUM ('ONLINE', 'TPS');

-- Channel vote
CREATE TYPE vote_channel AS ENUM ('ONLINE', 'TPS');

-- =========================================================
-- TABEL: tps
-- =========================================================

CREATE TABLE tps (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    name                TEXT NOT NULL,
    location            TEXT NOT NULL,
    status              tps_status NOT NULL DEFAULT 'DRAFT',
    voting_date         DATE NOT NULL,
    open_time           TIME NOT NULL,
    close_time          TIME NOT NULL,
    capacity_estimate   INTEGER,
    area_faculty_id     BIGINT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_tps_election_code
    ON tps (election_id, code);

CREATE INDEX idx_tps_election
    ON tps (election_id);

CREATE INDEX idx_tps_status
    ON tps (status);

-- =========================================================
-- TABEL: tps_qr
-- =========================================================

CREATE TABLE tps_qr (
    id              BIGSERIAL PRIMARY KEY,
    tps_id          BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    qr_secret       TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX ux_tps_qr_secret
    ON tps_qr (qr_secret);

CREATE INDEX idx_tps_qr_active_tps
    ON tps_qr (tps_id, is_active);

-- =========================================================
-- TABEL: tps_checkins
-- =========================================================

CREATE TABLE tps_checkins (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    tps_id              BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    qr_id               BIGINT NOT NULL REFERENCES tps_qr(id) ON DELETE RESTRICT,
    voter_id            BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,

    status              tps_checkin_status NOT NULL DEFAULT 'PENDING',

    scan_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at         TIMESTAMPTZ NULL,
    approved_by_id      BIGINT NULL REFERENCES user_accounts(id) ON DELETE SET NULL,
    rejection_reason    TEXT NULL,
    expires_at          TIMESTAMPTZ NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tps_checkins_tps_status_scan
    ON tps_checkins (tps_id, status, scan_at);

CREATE INDEX idx_tps_checkins_election_voter_status
    ON tps_checkins (election_id, voter_id, status, scan_at DESC);

-- =========================================================
-- TABEL: voter_status
-- =========================================================

CREATE TABLE voter_status (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    voter_id            BIGINT NOT NULL REFERENCES voters(id) ON DELETE CASCADE,

    is_eligible         BOOLEAN NOT NULL DEFAULT TRUE,
    has_voted           BOOLEAN NOT NULL DEFAULT FALSE,

    voting_method       voting_method NULL,
    tps_id              BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,
    voted_at            TIMESTAMPTZ NULL,

    vote_token_hash     TEXT NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_voter_status_method_has_voted
        CHECK (
            (has_voted = FALSE AND voting_method IS NULL AND tps_id IS NULL AND voted_at IS NULL)
         OR (has_voted = TRUE AND voting_method IS NOT NULL AND voted_at IS NOT NULL)
        )
);

CREATE UNIQUE INDEX ux_voter_status_election_voter
    ON voter_status (election_id, voter_id);

CREATE INDEX idx_voter_status_election_has_voted
    ON voter_status (election_id, has_voted);

CREATE UNIQUE INDEX ux_voter_status_token_hash
    ON voter_status (vote_token_hash)
    WHERE vote_token_hash IS NOT NULL;

-- =========================================================
-- TABEL: votes
-- =========================================================

CREATE TABLE votes (
    id              BIGSERIAL PRIMARY KEY,
    election_id     BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    candidate_id    BIGINT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,

    token_hash      TEXT NOT NULL,
    channel         vote_channel NOT NULL,
    tps_id          BIGINT NULL REFERENCES tps(id) ON DELETE SET NULL,

    cast_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_votes_token_hash
    ON votes (token_hash);

CREATE INDEX idx_votes_election_candidate
    ON votes (election_id, candidate_id);

CREATE INDEX idx_votes_election_channel
    ON votes (election_id, channel);

CREATE INDEX idx_votes_election_tps
    ON votes (election_id, tps_id);

-- =========================================================
-- TRIGGER: auto-update updated_at
-- =========================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tps_updated_at 
    BEFORE UPDATE ON tps
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_voter_status_updated_at 
    BEFORE UPDATE ON voter_status
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
