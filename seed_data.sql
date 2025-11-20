-- Seed Data for Pemira API
-- This script creates sample data for testing

-- Clear existing data (in reverse FK order)
TRUNCATE TABLE user_sessions CASCADE;
TRUNCATE TABLE user_accounts CASCADE;
TRUNCATE TABLE candidates CASCADE;
TRUNCATE TABLE voters CASCADE;
TRUNCATE TABLE elections CASCADE;
TRUNCATE TABLE tps CASCADE;

-- Insert Elections
INSERT INTO elections (id, code, name, year, status, online_enabled, tps_enabled, description,
                      voting_start_at, voting_end_at, created_at, updated_at)
VALUES 
    (1, 'PEMIRA-2024', 'Pemilihan Raya BEM 2024', 2024, 'VOTING_OPEN', true, true,
     'Pemilihan Raya Badan Eksekutif Mahasiswa periode 2024-2025',
     NOW() - INTERVAL '1 day', NOW() + INTERVAL '7 days', NOW(), NOW()),
    (2, 'PEMIRA-2025', 'Pemilihan Raya BEM 2025', 2025, 'DRAFT', true, true,
     'Pemilihan Raya Badan Eksekutif Mahasiswa periode 2025-2026',
     NOW() + INTERVAL '30 days', NOW() + INTERVAL '37 days', NOW(), NOW());

-- Reset sequence
SELECT setval('elections_id_seq', (SELECT MAX(id) FROM elections));

-- Insert TPS (Tempat Pemungutan Suara)
INSERT INTO tps (id, election_id, code, name, location, status, voting_date, open_time, close_time, 
                capacity_estimate, created_at, updated_at)
VALUES 
    (1, 1, 'TPS-FT-001', 'TPS Fakultas Teknik', 'Gedung A Lt.1 Fakultas Teknik', 
     'ACTIVE', CURRENT_DATE, '08:00:00', '16:00:00', 100, NOW(), NOW()),
    (2, 1, 'TPS-FE-001', 'TPS Fakultas Ekonomi', 'Gedung B Lt.2 Fakultas Ekonomi', 
     'ACTIVE', CURRENT_DATE, '08:00:00', '16:00:00', 80, NOW(), NOW()),
    (3, 1, 'TPS-MIPA-001', 'TPS Fakultas MIPA', 'Lab Komputer MIPA', 
     'ACTIVE', CURRENT_DATE, '08:00:00', '16:00:00', 60, NOW(), NOW());

SELECT setval('tps_id_seq', (SELECT MAX(id) FROM tps));

-- Insert Voters (DPT - Daftar Pemilih Tetap)
INSERT INTO voters (id, nim, name, email, faculty_code, faculty_name, study_program_code, study_program_name, 
                   cohort_year, academic_status, created_at, updated_at)
VALUES 
    (1, '2021001', 'Ahmad Rizki', 'ahmad.rizki@university.ac.id', 'FT', 'Fakultas Teknik', 
     'TI', 'Teknik Informatika', 2021, 'ACTIVE', NOW(), NOW()),
    (2, '2021002', 'Siti Nurhaliza', 'siti.nur@university.ac.id', 'FE', 'Fakultas Ekonomi', 
     'MJ', 'Manajemen', 2021, 'ACTIVE', NOW(), NOW()),
    (3, '2021003', 'Budi Santoso', 'budi.santoso@university.ac.id', 'FMIPA', 'Fakultas MIPA', 
     'MAT', 'Matematika', 2021, 'ACTIVE', NOW(), NOW()),
    (4, '2022001', 'Dewi Lestari', 'dewi.lestari@university.ac.id', 'FT', 'Fakultas Teknik', 
     'TE', 'Teknik Elektro', 2022, 'ACTIVE', NOW(), NOW()),
    (5, '2022002', 'Eko Prasetyo', 'eko.prasetyo@university.ac.id', 'FE', 'Fakultas Ekonomi', 
     'AK', 'Akuntansi', 2022, 'ACTIVE', NOW(), NOW());

SELECT setval('voters_id_seq', (SELECT MAX(id) FROM voters));

-- Insert Candidates
INSERT INTO candidates (id, election_id, candidate_number, chairman_name, vice_chairman_name, 
                       chairman_nim, vice_chairman_nim, vision, mission, photo_url, status, created_at, updated_at)
VALUES 
    (1, 1, 1, 'Ardi Pratama', 'Putri Wulandari', '2020001', '2020002',
     'Mewujudkan kampus yang inklusif, inovatif, dan berdaya saing tinggi',
     'Meningkatkan kesejahteraan mahasiswa, Mendorong prestasi akademik dan non-akademik, Membangun komunikasi yang baik dengan pihak kampus',
     'https://example.com/photo1.jpg',
     'APPROVED', NOW(), NOW()),
    
    (2, 1, 2, 'Bima Saputra', 'Sarah Amelia', '2020003', '2020004',
     'Menciptakan lingkungan kampus yang demokratis dan responsif terhadap aspirasi mahasiswa',
     'Transparansi pengelolaan organisasi, Program kewirausahaan mahasiswa, Peningkatan fasilitas kampus',
     'https://example.com/photo2.jpg',
     'APPROVED', NOW(), NOW()),
    
    (3, 1, 3, 'Citra Dewi', 'Doni Hermawan', '2020005', '2020006',
     'Menjadikan kampus sebagai role model kampus hijau dan berkelanjutan',
     'Program kampus hijau, Kegiatan sosial kemasyarakatan, Digitalisasi layanan mahasiswa',
     'https://example.com/photo3.jpg',
     'APPROVED', NOW(), NOW());

SELECT setval('candidates_id_seq', (SELECT MAX(id) FROM candidates));-- Insert User Accounts with hashed passwords
-- Password for all users: "password123" (hashed with bcrypt)
-- Hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

INSERT INTO user_accounts (id, username, email, password_hash, full_name, role, is_active, created_at, updated_at)
VALUES 
    -- Admin
    (1, 'admin', 'admin@pemira.ac.id', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'Administrator', 'ADMIN', true, NOW(), NOW()),
    
    -- Panitia
    (2, 'panitia', 'panitia@pemira.ac.id', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'Panitia Pemira', 'PANITIA', true, NOW(), NOW()),
    
    -- Ketua TPS
    (3, 'ketua_tps1', 'ketua_tps1@pemira.ac.id', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'Ketua TPS 1', 'KETUA_TPS', true, NOW(), NOW()),
    
    -- Operator Panel
    (4, 'operator', 'operator@pemira.ac.id', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'Operator Panel', 'OPERATOR_PANEL', true, NOW(), NOW()),
    
    -- Viewer
    (5, 'viewer', 'viewer@pemira.ac.id', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'Public Viewer', 'VIEWER', true, NOW(), NOW());

SELECT setval('user_accounts_id_seq', (SELECT MAX(id) FROM user_accounts));

-- Display summary
SELECT 'Seed data inserted successfully!' as status;
SELECT COUNT(*) as election_count FROM elections;
SELECT COUNT(*) as tps_count FROM tps;
SELECT COUNT(*) as voter_count FROM voters;
SELECT COUNT(*) as candidate_count FROM candidates;
SELECT COUNT(*) as user_count FROM user_accounts;
