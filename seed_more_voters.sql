-- Seed more voters for testing
INSERT INTO voters (nim, name, email, faculty_code, faculty_name, study_program_code, study_program_name, cohort_year, academic_status, created_at, updated_at)
VALUES 
    -- Fakultas Teknik
    ('2021101', 'Agus Santoso', 'agus.santoso@university.ac.id', 'FT', 'Fakultas Teknik', 'TI', 'Teknik Informatika', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021102', 'Budi Hermawan', 'budi.hermawan@university.ac.id', 'FT', 'Fakultas Teknik', 'TI', 'Teknik Informatika', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021103', 'Citra Lestari', 'citra.lestari@university.ac.id', 'FT', 'Fakultas Teknik', 'TE', 'Teknik Elektro', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021104', 'Dian Pratama', 'dian.pratama@university.ac.id', 'FT', 'Fakultas Teknik', 'TE', 'Teknik Elektro', 2021, 'ACTIVE', NOW(), NOW()),
    ('2022101', 'Eka Putri', 'eka.putri@university.ac.id', 'FT', 'Fakultas Teknik', 'TI', 'Teknik Informatika', 2022, 'ACTIVE', NOW(), NOW()),
    ('2022102', 'Fajar Ramadan', 'fajar.ramadan@university.ac.id', 'FT', 'Fakultas Teknik', 'TS', 'Teknik Sipil', 2022, 'ACTIVE', NOW(), NOW()),
    ('2022103', 'Gita Sari', 'gita.sari@university.ac.id', 'FT', 'Fakultas Teknik', 'TM', 'Teknik Mesin', 2022, 'ACTIVE', NOW(), NOW()),
    ('2023101', 'Hendra Wijaya', 'hendra.wijaya@university.ac.id', 'FT', 'Fakultas Teknik', 'TI', 'Teknik Informatika', 2023, 'ACTIVE', NOW(), NOW()),
    ('2023102', 'Indah Permata', 'indah.permata@university.ac.id', 'FT', 'Fakultas Teknik', 'TE', 'Teknik Elektro', 2023, 'ACTIVE', NOW(), NOW()),
    ('2023103', 'Joko Susilo', 'joko.susilo@university.ac.id', 'FT', 'Fakultas Teknik', 'TS', 'Teknik Sipil', 2023, 'ACTIVE', NOW(), NOW()),
    
    -- Fakultas Ekonomi
    ('2021201', 'Kartika Dewi', 'kartika.dewi@university.ac.id', 'FE', 'Fakultas Ekonomi', 'MJ', 'Manajemen', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021202', 'Linda Puspita', 'linda.puspita@university.ac.id', 'FE', 'Fakultas Ekonomi', 'AK', 'Akuntansi', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021203', 'Made Wira', 'made.wira@university.ac.id', 'FE', 'Fakultas Ekonomi', 'EP', 'Ekonomi Pembangunan', 2021, 'ACTIVE', NOW(), NOW()),
    ('2022201', 'Nina Safira', 'nina.safira@university.ac.id', 'FE', 'Fakultas Ekonomi', 'MJ', 'Manajemen', 2022, 'ACTIVE', NOW(), NOW()),
    ('2022202', 'Omar Fauzi', 'omar.fauzi@university.ac.id', 'FE', 'Fakultas Ekonomi', 'AK', 'Akuntansi', 2022, 'ACTIVE', NOW(), NOW()),
    ('2023201', 'Putri Ayu', 'putri.ayu@university.ac.id', 'FE', 'Fakultas Ekonomi', 'MJ', 'Manajemen', 2023, 'ACTIVE', NOW(), NOW()),
    ('2023202', 'Qori Rahman', 'qori.rahman@university.ac.id', 'FE', 'Fakultas Ekonomi', 'EP', 'Ekonomi Pembangunan', 2023, 'ACTIVE', NOW(), NOW()),
    
    -- Fakultas MIPA
    ('2021301', 'Rina Wati', 'rina.wati@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'MAT', 'Matematika', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021302', 'Sandi Kurnia', 'sandi.kurnia@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'FIS', 'Fisika', 2021, 'ACTIVE', NOW(), NOW()),
    ('2021303', 'Tari Anggraini', 'tari.anggraini@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'KIM', 'Kimia', 2021, 'ACTIVE', NOW(), NOW()),
    ('2022301', 'Umar Hadi', 'umar.hadi@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'BIO', 'Biologi', 2022, 'ACTIVE', NOW(), NOW()),
    ('2022302', 'Vina Melati', 'vina.melati@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'MAT', 'Matematika', 2022, 'ACTIVE', NOW(), NOW()),
    ('2023301', 'Wawan Setiawan', 'wawan.setiawan@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'FIS', 'Fisika', 2023, 'ACTIVE', NOW(), NOW()),
    ('2023302', 'Yanti Kusuma', 'yanti.kusuma@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'KIM', 'Kimia', 2023, 'ACTIVE', NOW(), NOW()),
    ('2023303', 'Zaki Mubarok', 'zaki.mubarok@university.ac.id', 'FMIPA', 'Fakultas MIPA', 'BIO', 'Biologi', 2023, 'ACTIVE', NOW(), NOW());

-- Create user accounts for new voters
INSERT INTO user_accounts (username, email, password_hash, full_name, role, voter_id, is_active, created_at, updated_at)
SELECT 
    v.nim,
    v.email,
    '$2a$10$VuGd0ekbW2lxZejZSZjKE.C548Fi9zIjx3XgfBKdKjZK53SW/C6OO', -- password123
    v.name,
    'STUDENT',
    v.id,
    true,
    NOW(),
    NOW()
FROM voters v
WHERE v.nim IN (
    '2021101', '2021102', '2021103', '2021104', '2022101', '2022102', '2022103', '2023101', '2023102', '2023103',
    '2021201', '2021202', '2021203', '2022201', '2022202', '2023201', '2023202',
    '2021301', '2021302', '2021303', '2022301', '2022302', '2023301', '2023302', '2023303'
);

-- Initialize voter_status for election 1 for all voters
INSERT INTO voter_status (election_id, voter_id, is_eligible, has_voted, online_allowed, tps_allowed, created_at, updated_at)
SELECT 
    1, -- election_id
    v.id,
    true, -- is_eligible
    false, -- has_voted
    true, -- online_allowed
    true, -- tps_allowed
    NOW(),
    NOW()
FROM voters v
ON CONFLICT (election_id, voter_id) DO NOTHING;

SELECT COUNT(*) as total_voters FROM voters;
SELECT COUNT(*) as total_student_accounts FROM user_accounts WHERE role = 'STUDENT';
SELECT COUNT(*) as total_voter_status FROM voter_status WHERE election_id = 1;
