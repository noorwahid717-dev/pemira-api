-- Seed data for lecturers (dosen) and staff (tenaga kependidikan)
-- Password untuk semua user: password123

-- =========================================================
-- LECTURERS (DOSEN)
-- =========================================================

-- Insert lecturers data
INSERT INTO lecturers (id, nidn, name, email, faculty_code, faculty_name, department_code, department_name, position, employment_status, created_at, updated_at) VALUES
(1, '0101018901', 'Dr. Ahmad Kusuma, S.Kom., M.T.', 'ahmad.kusuma@universitas.ac.id', 'FT', 'Fakultas Teknik', 'TI', 'Teknik Informatika', 'Lektor Kepala', 'Tetap', NOW(), NOW()),
(2, '0102019002', 'Dra. Siti Nurjanah, M.Pd.', 'siti.nurjanah@universitas.ac.id', 'FIP', 'Fakultas Ilmu Pendidikan', 'PGSD', 'Pendidikan Guru Sekolah Dasar', 'Lektor', 'Tetap', NOW(), NOW()),
(3, '0103019103', 'Prof. Dr. Budi Santoso, S.E., M.M.', 'budi.santoso@universitas.ac.id', 'FEB', 'Fakultas Ekonomi dan Bisnis', 'MANAJEMEN', 'Manajemen', 'Guru Besar', 'Tetap', NOW(), NOW()),
(4, '0104019204', 'Dr. Retno Wulandari, S.Si., M.Sc.', 'retno.wulandari@universitas.ac.id', 'FMIPA', 'Fakultas Matematika dan Ilmu Pengetahuan Alam', 'MATEMATIKA', 'Matematika', 'Lektor', 'Tetap', NOW(), NOW()),
(5, '0105019305', 'Ir. Joko Widodo, M.T.', 'joko.widodo@universitas.ac.id', 'FT', 'Fakultas Teknik', 'TS', 'Teknik Sipil', 'Asisten Ahli', 'Tidak Tetap', NOW(), NOW());

SELECT setval('lecturers_id_seq', (SELECT MAX(id) FROM lecturers));

-- =========================================================
-- STAFF MEMBERS (TENAGA KEPENDIDIKAN)
-- =========================================================

-- Insert staff members data
INSERT INTO staff_members (id, nip, name, email, unit_code, unit_name, position, employment_status, created_at, updated_at) VALUES
(1, '198901012015041001', 'Bambang Setiawan, S.Sos.', 'bambang.setiawan@universitas.ac.id', 'BAU', 'Biro Administrasi Umum', 'Kepala Sub Bagian Umum', 'PNS', NOW(), NOW()),
(2, '199002012016051002', 'Dewi Kusumawati, A.Md.', 'dewi.kusumawati@universitas.ac.id', 'BAAK', 'Biro Administrasi Akademik dan Kemahasiswaan', 'Staff Administrasi Akademik', 'PNS', NOW(), NOW()),
(3, '199103012017061003', 'Eko Prasetyo, S.Kom.', 'eko.prasetyo@universitas.ac.id', 'UPT-TIK', 'Unit Pelaksana Teknis Teknologi Informasi dan Komunikasi', 'Administrator Sistem', 'PNS', NOW(), NOW()),
(4, '199204012018071004', 'Fitri Handayani, S.E.', 'fitri.handayani@universitas.ac.id', 'BAK', 'Biro Administrasi Keuangan', 'Bendahara', 'Kontrak', NOW(), NOW()),
(5, '199305012019081005', 'Gunawan Wijaya', 'gunawan.wijaya@universitas.ac.id', 'PERPUSTAKAAN', 'Perpustakaan Universitas', 'Pustakawan', 'Kontrak', NOW(), NOW());

SELECT setval('staff_members_id_seq', (SELECT MAX(id) FROM staff_members));

-- =========================================================
-- USER ACCOUNTS FOR LECTURERS
-- =========================================================

-- Password hash untuk 'password123'
-- bcrypt hash generated with cost 10
INSERT INTO user_accounts (username, email, password_hash, full_name, role, lecturer_id, is_active, created_at, updated_at) VALUES
('0101018901', 'ahmad.kusuma@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Dr. Ahmad Kusuma, S.Kom., M.T.', 'LECTURER', 1, true, NOW(), NOW()),
('0102019002', 'siti.nurjanah@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Dra. Siti Nurjanah, M.Pd.', 'LECTURER', 2, true, NOW(), NOW()),
('0103019103', 'budi.santoso@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Prof. Dr. Budi Santoso, S.E., M.M.', 'LECTURER', 3, true, NOW(), NOW()),
('0104019204', 'retno.wulandari@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Dr. Retno Wulandari, S.Si., M.Sc.', 'LECTURER', 4, true, NOW(), NOW()),
('0105019305', 'joko.widodo@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Ir. Joko Widodo, M.T.', 'LECTURER', 5, true, NOW(), NOW());

-- =========================================================
-- USER ACCOUNTS FOR STAFF
-- =========================================================

INSERT INTO user_accounts (username, email, password_hash, full_name, role, staff_id, is_active, created_at, updated_at) VALUES
('198901012015041001', 'bambang.setiawan@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Bambang Setiawan, S.Sos.', 'STAFF', 1, true, NOW(), NOW()),
('199002012016051002', 'dewi.kusumawati@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Dewi Kusumawati, A.Md.', 'STAFF', 2, true, NOW(), NOW()),
('199103012017061003', 'eko.prasetyo@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Eko Prasetyo, S.Kom.', 'STAFF', 3, true, NOW(), NOW()),
('199204012018071004', 'fitri.handayani@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Fitri Handayani, S.E.', 'STAFF', 4, true, NOW(), NOW()),
('199305012019081005', 'gunawan.wijaya@universitas.ac.id', '$2a$10$rYc0RaXF5c.F3pHmJVvzwOE5h1VV1VLK5Cl5aGjL/WT5s0TqKqYWG', 'Gunawan Wijaya', 'STAFF', 5, true, NOW(), NOW());

-- =========================================================
-- SUMMARY
-- =========================================================

-- Lecturers created:
-- 1. 0101018901 - Dr. Ahmad Kusuma (Teknik Informatika)
-- 2. 0102019002 - Dra. Siti Nurjanah (PGSD)
-- 3. 0103019103 - Prof. Dr. Budi Santoso (Manajemen)
-- 4. 0104019204 - Dr. Retno Wulandari (Matematika)
-- 5. 0105019305 - Ir. Joko Widodo (Teknik Sipil)

-- Staff members created:
-- 1. 198901012015041001 - Bambang Setiawan (BAU)
-- 2. 199002012016051002 - Dewi Kusumawati (BAAK)
-- 3. 199103012017061003 - Eko Prasetyo (UPT-TIK)
-- 4. 199204012018071004 - Fitri Handayani (BAK)
-- 5. 199305012019081005 - Gunawan Wijaya (Perpustakaan)

-- Login credentials:
-- Username: NIDN untuk dosen (contoh: 0101018901)
-- Username: NIP untuk staff (contoh: 198901012015041001)
-- Password: password123 (untuk semua user)
