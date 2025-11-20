-- +goose Up
-- Add support for LECTURER and STAFF roles

-- Update user_role enum to include LECTURER and STAFF
ALTER TYPE user_role ADD VALUE 'LECTURER';
ALTER TYPE user_role ADD VALUE 'STAFF';

-- =========================================================
-- TABEL: lecturers (Dosen)
-- =========================================================
CREATE TABLE lecturers (
    id                      BIGSERIAL PRIMARY KEY,
    
    nidn                    TEXT NOT NULL,      -- Nomor Induk Dosen Nasional
    name                    TEXT NOT NULL,
    email                   TEXT NULL,
    
    faculty_code            TEXT NULL,
    faculty_name            TEXT NULL,
    
    department_code         TEXT NULL,
    department_name         TEXT NULL,
    
    position                TEXT NULL,          -- Jabatan: Lektor, Asisten Ahli, dst
    employment_status       TEXT NULL,          -- Status: Tetap, Tidak Tetap
    
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_lecturers_nidn
    ON lecturers (nidn);

CREATE INDEX idx_lecturers_faculty
    ON lecturers (faculty_code);

CREATE INDEX idx_lecturers_department
    ON lecturers (department_code);

-- =========================================================
-- TABEL: staff_members (Staff/Tenaga Kependidikan)
-- =========================================================
CREATE TABLE staff_members (
    id                      BIGSERIAL PRIMARY KEY,
    
    nip                     TEXT NOT NULL,      -- Nomor Induk Pegawai
    name                    TEXT NOT NULL,
    email                   TEXT NULL,
    
    unit_code               TEXT NULL,          -- Unit kerja
    unit_name               TEXT NULL,
    
    position                TEXT NULL,          -- Jabatan
    employment_status       TEXT NULL,          -- Status kepegawaian
    
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_staff_members_nip
    ON staff_members (nip);

CREATE INDEX idx_staff_members_unit
    ON staff_members (unit_code);

-- =========================================================
-- Update user_accounts table
-- =========================================================
ALTER TABLE user_accounts
    ADD COLUMN lecturer_id BIGINT NULL REFERENCES lecturers(id) ON DELETE SET NULL,
    ADD COLUMN staff_id BIGINT NULL REFERENCES staff_members(id) ON DELETE SET NULL;

CREATE INDEX idx_user_accounts_lecturer_id
    ON user_accounts (lecturer_id)
    WHERE lecturer_id IS NOT NULL;

CREATE INDEX idx_user_accounts_staff_id
    ON user_accounts (staff_id)
    WHERE staff_id IS NOT NULL;

-- =========================================================
-- Update user_accounts to include full_name and email
-- =========================================================
ALTER TABLE user_accounts
    ADD COLUMN IF NOT EXISTS full_name TEXT NULL,
    ADD COLUMN IF NOT EXISTS email TEXT NULL;
