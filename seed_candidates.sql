-- Seed data for candidates
-- Candidates untuk PEMIRA-2024 (election_id = 1)

INSERT INTO candidates (
    election_id, number, name, photo_url, short_bio, long_bio, tagline,
    faculty_name, study_program_name, cohort_year, vision, missions, 
    main_programs, media, social_links, status
) VALUES
(
    1,  -- election_id
    1,  -- number
    'Ahmad Budi - Siti Rahma',  -- name
    'https://example.com/photos/candidate-1.jpg',  -- photo_url
    'Mahasiswa aktif dan berpengalaman dalam organisasi kampus',  -- short_bio
    'Ahmad Budi adalah mahasiswa Teknik Informatika angkatan 2021 yang aktif di BEM dan berbagai organisasi. Siti Rahma adalah mahasiswa Manajemen yang berpengalaman dalam event organizing.',  -- long_bio
    'Bersama Membangun Kampus Digital',  -- tagline
    'Fakultas Teknik',  -- faculty_name
    'Teknik Informatika',  -- study_program_name
    2021,  -- cohort_year
    'Mewujudkan kampus yang inklusif, digital, dan berprestasi',  -- vision
    '["Meningkatkan fasilitas teknologi kampus", "Memperluas program beasiswa", "Mengembangkan soft skill mahasiswa"]'::jsonb,  -- missions
    '[{"title": "Program Digitalisasi Kampus", "description": "Implementasi aplikasi akademik terintegrasi"}, {"title": "Beasiswa Prestasi", "description": "Menambah kuota beasiswa untuk mahasiswa berprestasi"}]'::jsonb,  -- main_programs
    '{"video_url": "https://youtube.com/watch?v=xxx", "instagram": "@paslon1_pemira"}'::jsonb,  -- media
    '[{"platform": "Instagram", "url": "https://instagram.com/paslon1"}, {"platform": "Twitter", "url": "https://twitter.com/paslon1"}]'::jsonb,  -- social_links
    'APPROVED'  -- status
),
(
    1,  -- election_id
    2,  -- number
    'Devi Kusuma - Eko Prasetyo',  -- name
    'https://example.com/photos/candidate-2.jpg',  -- photo_url
    'Pengalaman di bidang kemahasiswaan dan sosial',  -- short_bio
    'Devi Kusuma adalah mahasiswa Ekonomi yang aktif dalam kegiatan sosial kemasyarakatan. Eko Prasetyo adalah mahasiswa Teknik Sipil dengan pengalaman organisasi yang luas.',  -- long_bio
    'Kampus Sejahtera untuk Semua',  -- tagline
    'Fakultas Ekonomi',  -- faculty_name
    'Manajemen',  -- study_program_name
    2021,  -- cohort_year
    'Menciptakan kampus yang peduli dan berdaya saing',  -- vision
    '["Meningkatkan kesejahteraan mahasiswa", "Memperkuat kerjasama dengan industri", "Mengoptimalkan program pengembangan diri"]'::jsonb,  -- missions
    '[{"title": "Program Kesejahteraan Mahasiswa", "description": "Subsidi transportasi dan makan untuk mahasiswa"}, {"title": "Kerjasama Industri", "description": "Membangun partnership dengan perusahaan untuk magang dan kerja"}]'::jsonb,  -- main_programs
    '{"video_url": "https://youtube.com/watch?v=yyy", "instagram": "@paslon2_pemira"}'::jsonb,  -- media
    '[{"platform": "Instagram", "url": "https://instagram.com/paslon2"}, {"platform": "TikTok", "url": "https://tiktok.com/@paslon2"}]'::jsonb,  -- social_links
    'APPROVED'  -- status
),
(
    1,  -- election_id
    3,  -- number
    'Farhan Rizki - Intan Permata',  -- name
    'https://example.com/photos/candidate-3.jpg',  -- photo_url
    'Inovatif dan visioner untuk kemajuan kampus',  -- short_bio
    'Farhan Rizki adalah mahasiswa MIPA dengan berbagai prestasi akademik. Intan Permata adalah mahasiswa Teknik dengan pengalaman kepemimpinan yang kuat.',  -- long_bio
    'Inovasi untuk Kemajuan Bersama',  -- tagline
    'Fakultas MIPA',  -- faculty_name
    'Matematika',  -- study_program_name
    2022,  -- cohort_year
    'Membangun kampus yang inovatif dan kompetitif di tingkat nasional',  -- vision
    '["Mendorong riset dan inovasi mahasiswa", "Meningkatkan prestasi akademik dan non-akademik", "Membangun ekosistem startup kampus"]'::jsonb,  -- missions
    '[{"title": "Inkubator Startup Mahasiswa", "description": "Mendukung mahasiswa yang ingin membangun startup"}, {"title": "Program Riset Unggulan", "description": "Pendanaan dan mentoring untuk riset mahasiswa"}]'::jsonb,  -- main_programs
    '{"video_url": "https://youtube.com/watch?v=zzz", "instagram": "@paslon3_pemira"}'::jsonb,  -- media
    '[{"platform": "Instagram", "url": "https://instagram.com/paslon3"}, {"platform": "LinkedIn", "url": "https://linkedin.com/in/paslon3"}]'::jsonb,  -- social_links
    'APPROVED'  -- status
);

-- Summary
-- 3 candidates created for PEMIRA-2024
-- All candidates are APPROVED and ready for voting
