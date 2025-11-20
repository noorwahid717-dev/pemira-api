-- Create student user accounts linked to voters
-- Password for all: "password123"

INSERT INTO user_accounts (username, email, password_hash, full_name, role, voter_id, is_active, created_at, updated_at)
VALUES 
    -- Student 1 - linked to voter 1 (Ahmad Rizki - NIM: 2021001)
    ('2021001', 'ahmad.rizki@university.ac.id', 
     '$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke', 
     'Ahmad Rizki', 'STUDENT', 1, true, NOW(), NOW()),
    
    -- Student 2 - linked to voter 2 (Siti Nurhaliza - NIM: 2021002)
    ('2021002', 'siti.nur@university.ac.id', 
     '$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke', 
     'Siti Nurhaliza', 'STUDENT', 2, true, NOW(), NOW()),
    
    -- Student 3 - linked to voter 3 (Budi Santoso - NIM: 2021003)
    ('2021003', 'budi.santoso@university.ac.id', 
     '$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke', 
     'Budi Santoso', 'STUDENT', 3, true, NOW(), NOW()),
    
    -- Student 4 - linked to voter 4 (Dewi Lestari - NIM: 2022001)
    ('2022001', 'dewi.lestari@university.ac.id', 
     '$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke', 
     'Dewi Lestari', 'STUDENT', 4, true, NOW(), NOW()),
    
    -- Student 5 - linked to voter 5 (Eko Prasetyo - NIM: 2022002)
    ('2022002', 'eko.prasetyo@university.ac.id', 
     '$2y$10$A6f.x0pip2mUFBw4qEumbOzfnJ2Yp/IjuTGZpewZeUzxRskHVhSke', 
     'Eko Prasetyo', 'STUDENT', 5, true, NOW(), NOW());

-- Update TPS operator to link with TPS
UPDATE user_accounts 
SET tps_id = 1 
WHERE username = 'ketua_tps1';

-- Display result
SELECT 
    ua.id, 
    ua.username, 
    ua.role, 
    ua.voter_id,
    v.nim,
    v.name as voter_name,
    ua.tps_id
FROM user_accounts ua
LEFT JOIN voters v ON ua.voter_id = v.id
ORDER BY ua.role, ua.id;
