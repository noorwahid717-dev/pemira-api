# Quick Test Guide: Auto-Enrollment Bug Fix

## 🧪 Test Scenario

Test bahwa user baru yang registrasi otomatis muncul di DPT list.

## Prerequisites

1. Pastikan ada election aktif (status: VOTING_OPEN, REGISTRATION, CAMPAIGN, atau CLOSED)
2. Note current count di DPT: `GET /api/v1/admin/elections/{electionID}/voters`

## Test Case 1: Student Registration

### 1. Register Student Baru
```bash
curl -X POST http://localhost:8080/api/v1/auth/register/student \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "2025001",
    "name": "Test Student DPT",
    "password": "password123",
    "email": "test.student@pemira.ac.id",
    "faculty_name": "Fakultas Teknik",
    "study_program_name": "Teknik Informatika",
    "semester": "3",
    "voting_mode": "ONLINE"
  }'
```

**Expected Response:**
```json
{
  "user": {
    "id": 85,
    "username": "2025001",
    "role": "STUDENT",
    "voter_id": 85,
    "profile": {
      "name": "Test Student DPT",
      "faculty_name": "Fakultas Teknik",
      "study_program_name": "Teknik Informatika",
      "semester": "3"
    }
  },
  "message": "Registrasi mahasiswa berhasil.",
  "voting_mode": "ONLINE"
}
```

### 2. Check DPT List (Should Include New Voter)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?page=1&per_page=50" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

**Expected:**
- ✅ Total count bertambah 1
- ✅ Ada entry untuk NIM "2025001" dengan:
  - `status`: "PENDING"
  - `voting_method`: "ONLINE"
  - `name`: "Test Student DPT"

### 3. Check Database Directly
```sql
SELECT 
    v.id as voter_id,
    v.nim,
    v.name,
    ev.id as election_voter_id,
    ev.status,
    ev.voting_method,
    ev.election_id
FROM voters v
LEFT JOIN election_voters ev ON ev.voter_id = v.id
WHERE v.nim = '2025001';
```

**Expected:**
- `voter_id`: not null (e.g., 85)
- `election_voter_id`: not null (bukan NULL!)
- `status`: PENDING
- `voting_method`: ONLINE

## Test Case 2: Lecturer Registration

### 1. Register Lecturer Baru
```bash
curl -X POST http://localhost:8080/api/v1/auth/register/lecturer-staff \
  -H "Content-Type: application/json" \
  -d '{
    "type": "LECTURER",
    "nidn": "0123456789",
    "name": "Test Lecturer DPT",
    "password": "password123",
    "email": "test.lecturer@pemira.ac.id",
    "faculty_name": "Fakultas Teknik",
    "department_name": "Teknik Informatika",
    "position": "Lektor",
    "voting_mode": "TPS"
  }'
```

### 2. Verify di DPT List
```bash
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?page=1&per_page=50" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

**Expected:**
- ✅ Ada entry untuk NIDN "0123456789"
- ✅ `status`: "PENDING"
- ✅ `voting_method`: "TPS"

## Test Case 3: Staff Registration

### 1. Register Staff Baru
```bash
curl -X POST http://localhost:8080/api/v1/auth/register/lecturer-staff \
  -H "Content-Type: application/json" \
  -d '{
    "type": "STAFF",
    "nip": "1234567890",
    "name": "Test Staff DPT",
    "password": "password123",
    "email": "test.staff@pemira.ac.id",
    "unit_name": "Biro Akademik",
    "position": "Staff Admin",
    "voting_mode": "ONLINE"
  }'
```

### 2. Verify di DPT List
Same as above, check for NIP "1234567890"

## Test Case 4: Dashboard Consistency

### Before Registration
```bash
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/dashboard" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

Note: `total_eligible`

### After Registration
Call the same endpoint again.

**Expected:**
- ✅ `total_eligible` bertambah 1
- ✅ Jumlah di DPT list juga bertambah 1 (konsisten!)

## 🔍 SQL Verification Query

```sql
-- Compare voters vs election_voters count
SELECT 
    'Total Voters' as metric,
    COUNT(*) as count
FROM voters
WHERE created_at > NOW() - INTERVAL '1 hour'

UNION ALL

SELECT 
    'Enrolled in Election' as metric,
    COUNT(*) as count
FROM election_voters ev
JOIN voters v ON v.id = ev.voter_id
WHERE v.created_at > NOW() - INTERVAL '1 hour';
```

**Expected:** Kedua angka harus SAMA!

## ❌ Before Fix (Expected Failure)

**Symptom:**
```sql
-- Query result before fix
Total Voters          : 3  ✅
Enrolled in Election  : 0  ❌ (BUG!)
```

## ✅ After Fix (Expected Success)

**Symptom:**
```sql
-- Query result after fix
Total Voters          : 3  ✅
Enrolled in Election  : 3  ✅ (FIXED!)
```

## 🚨 Edge Cases to Test

### 1. Duplicate Registration (Should Not Error)
Try registering the same user twice → Should get "username/nim already exists" error, but if somehow voter exists, the enrollment should not crash.

### 2. Multiple Elections
If there are multiple elections, the voter should be enrolled to the current active one (determined by `FindOrCreateRegistrationElection`).

### 3. Different Voting Modes
- Register with `voting_mode: "ONLINE"` → Check `voting_method` = "ONLINE"
- Register with `voting_mode: "TPS"` → Check `voting_method` = "TPS"
- Register without `voting_mode` → Should default to "ONLINE"

## ✅ Success Criteria

- [ ] All 3 registration types (student/lecturer/staff) create election_voters entry
- [ ] New voters appear in DPT list immediately after registration
- [ ] Status is always "PENDING" for new registrations
- [ ] Voting method matches the choice during registration
- [ ] Dashboard count matches DPT list count
- [ ] No errors during registration flow
