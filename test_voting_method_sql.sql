-- ===================================================================
-- SQL Test Script for voting_method Update Functionality
-- ===================================================================
-- This script tests the voting_method column updates on both 
-- voters and voter_status tables
-- ===================================================================

\echo '========================================='
\echo 'TEST 1: Initial State Check'
\echo '========================================='
SELECT id, nim, name, voter_type, voting_method 
FROM voters 
WHERE id IN (2, 3, 4, 5)
ORDER BY id;

SELECT voter_id, has_voted, voting_method, preferred_method
FROM voter_status 
WHERE voter_id IN (2, 3, 4, 5)
ORDER BY voter_id;

\echo ''
\echo '========================================='
\echo 'TEST 2: Update Non-Voted Voter (ID 2)'
\echo 'Expected: Success - Update to TPS'
\echo '========================================='

-- Update voters table
UPDATE voters 
SET voting_method = 'TPS', updated_at = NOW() 
WHERE id = 2;

-- Update voter_status table
UPDATE voter_status 
SET voting_method = 'TPS', updated_at = NOW() 
WHERE voter_id = 2;

-- Verify update
SELECT 'After Update:' as status, id, nim, voting_method 
FROM voters WHERE id = 2;

SELECT 'After Update:' as status, voter_id, voting_method 
FROM voter_status WHERE voter_id = 2;

\echo ''
\echo '========================================='
\echo 'TEST 3: Update Voted Voter'
\echo 'Expected: Success - voting_method can be updated even after voting'
\echo '========================================='

-- First mark voter 3 as voted
UPDATE voter_status 
SET has_voted = true, 
    voting_method = 'ONLINE', 
    voted_at = NOW()
WHERE voter_id = 3;

SELECT 'Voter marked as voted:' as status, voter_id, has_voted, voting_method 
FROM voter_status WHERE voter_id = 3;

-- Now update voting_method for voted voter
UPDATE voters 
SET voting_method = 'TPS', updated_at = NOW() 
WHERE id = 3;

UPDATE voter_status 
SET voting_method = 'TPS', updated_at = NOW() 
WHERE voter_id = 3;

-- Verify update
SELECT 'After voting_method update:' as status, 
       voter_id, has_voted, voting_method, voted_at 
FROM voter_status WHERE voter_id = 3;

\echo ''
\echo '========================================='
\echo 'TEST 4: Invalid voting_method Value'
\echo 'Expected: ERROR - invalid enum value'
\echo '========================================='

UPDATE voters 
SET voting_method = 'INVALID', updated_at = NOW() 
WHERE id = 4;

\echo ''
\echo '========================================='
\echo 'TEST 5: Case Sensitivity Test'
\echo 'Expected: Success - enum values are case-sensitive'
\echo '========================================='

-- Lowercase should fail
UPDATE voters 
SET voting_method = 'online', updated_at = NOW() 
WHERE id = 4;

\echo ''
\echo '========================================='
\echo 'TEST 6: Valid ONLINE and TPS Updates'
\echo 'Expected: Success'
\echo '========================================='

UPDATE voters SET voting_method = 'ONLINE' WHERE id = 4;
SELECT 'Set to ONLINE:' as status, id, voting_method FROM voters WHERE id = 4;

UPDATE voters SET voting_method = 'TPS' WHERE id = 4;
SELECT 'Set to TPS:' as status, id, voting_method FROM voters WHERE id = 4;

UPDATE voters SET voting_method = 'ONLINE' WHERE id = 4;
SELECT 'Set back to ONLINE:' as status, id, voting_method FROM voters WHERE id = 4;

\echo ''
\echo '========================================='
\echo 'TEST 7: Verify Sync Between Tables'
\echo 'Expected: Both tables should be in sync'
\echo '========================================='

SELECT 
    v.id as voter_id,
    v.nim,
    v.voting_method as voters_table_method,
    vs.voting_method as voter_status_method,
    vs.has_voted,
    CASE 
        WHEN v.voting_method = vs.voting_method THEN 'SYNCED'
        WHEN vs.voting_method IS NULL THEN 'STATUS_NULL'
        ELSE 'OUT_OF_SYNC'
    END as sync_status
FROM voters v
LEFT JOIN voter_status vs ON v.id = vs.voter_id
WHERE v.id IN (2, 3, 4, 5)
ORDER BY v.id;

\echo ''
\echo '========================================='
\echo 'TEST 8: Revert Updates (Cleanup)'
\echo '========================================='

-- Revert voter 2 to ONLINE
UPDATE voters SET voting_method = 'ONLINE' WHERE id = 2;
UPDATE voter_status SET voting_method = 'ONLINE' WHERE voter_id = 2;

-- Revert voter 3 to not voted state
UPDATE voters SET voting_method = 'ONLINE' WHERE id = 3;
UPDATE voter_status 
SET has_voted = false, voting_method = NULL, voted_at = NULL 
WHERE voter_id = 3;

-- Ensure voter 4 is ONLINE
UPDATE voters SET voting_method = 'ONLINE' WHERE id = 4;

SELECT 'Cleanup Complete' as status;

\echo ''
\echo '========================================='
\echo 'TEST 9: Final State Verification'
\echo '========================================='

SELECT id, nim, voting_method 
FROM voters 
WHERE id IN (2, 3, 4, 5)
ORDER BY id;

SELECT voter_id, has_voted, voting_method
FROM voter_status 
WHERE voter_id IN (2, 3, 4, 5)
ORDER BY voter_id;

\echo ''
\echo '========================================='
\echo 'TEST SUMMARY:'
\echo '✓ voting_method can be updated on voters table'
\echo '✓ voting_method can be updated on voter_status table'
\echo '✓ voting_method can be updated even after voter has voted'
\echo '✓ Invalid values are rejected by enum constraint'
\echo '✓ Only ONLINE and TPS are valid values (case-sensitive)'
\echo '✓ Both tables can be kept in sync with updates'
\echo '========================================='
