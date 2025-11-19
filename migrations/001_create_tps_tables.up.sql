-- TPS (Tempat Pemungutan Suara) Table
CREATE TABLE IF NOT EXISTS tps (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    location TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    voting_date DATE,
    open_time VARCHAR(5) NOT NULL,
    close_time VARCHAR(5) NOT NULL,
    capacity_estimate INTEGER DEFAULT 0,
    area_faculty_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT tps_status_check CHECK (status IN ('DRAFT', 'ACTIVE', 'CLOSED'))
);

-- TPS QR Codes Table
CREATE TABLE IF NOT EXISTS tps_qr (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    qr_secret_suffix VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT tps_qr_unique UNIQUE (tps_id, qr_secret_suffix)
);

-- TPS Panitia Assignment Table
CREATE TABLE IF NOT EXISTS tps_panitia (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT tps_panitia_role_check CHECK (role IN ('KETUA_TPS', 'OPERATOR_PANEL')),
    CONSTRAINT tps_panitia_unique UNIQUE (tps_id, user_id)
);

-- TPS Check-ins Table
CREATE TABLE IF NOT EXISTS tps_checkins (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id) ON DELETE CASCADE,
    voter_id BIGINT NOT NULL,
    election_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    scan_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by_id BIGINT,
    rejection_reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT tps_checkins_status_check CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'USED', 'EXPIRED'))
);

-- Indexes for performance
CREATE INDEX idx_tps_election_id ON tps(election_id);
CREATE INDEX idx_tps_status ON tps(status);
CREATE INDEX idx_tps_code ON tps(code);

CREATE INDEX idx_tps_qr_tps_id ON tps_qr(tps_id);
CREATE INDEX idx_tps_qr_active ON tps_qr(is_active);

CREATE INDEX idx_tps_panitia_tps_id ON tps_panitia(tps_id);
CREATE INDEX idx_tps_panitia_user_id ON tps_panitia(user_id);

CREATE INDEX idx_tps_checkins_tps_id ON tps_checkins(tps_id);
CREATE INDEX idx_tps_checkins_voter_id ON tps_checkins(voter_id);
CREATE INDEX idx_tps_checkins_election_id ON tps_checkins(election_id);
CREATE INDEX idx_tps_checkins_status ON tps_checkins(status);
CREATE INDEX idx_tps_checkins_scan_at ON tps_checkins(scan_at);

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
CREATE TRIGGER update_tps_updated_at BEFORE UPDATE ON tps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tps_checkins_updated_at BEFORE UPDATE ON tps_checkins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
