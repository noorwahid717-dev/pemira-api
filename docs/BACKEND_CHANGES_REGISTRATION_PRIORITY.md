# Backend Changes Summary - Registration Priority & Auto-Enrollment

**Tanggal:** 27 November 2025  
**Status:** ✅ COMPLETED

## 🎯 Tujuan Perubahan

1. ✅ **Fix bug auto-enrollment ke DPT** - User baru otomatis masuk election_voters
2. ✅ **Update priority `/elections/current`** - Prioritas REGISTRATION_OPEN first
3. ✅ **Tambah endpoint `/elections/current-for-registration`** - Khusus untuk konteks registrasi
4. ✅ **Tambah validator `IsRegistrationAllowed()`** - Validasi election menerima registrasi
5. ⚠️ **Settings-based election** - SKIPPED (circular dependency issue)

## ✅ Perubahan Yang Diimplementasikan

### 1. Auto-Enrollment ke Election Voters (BUG FIX)

**Problem:** User baru registrasi tidak muncul di DPT list

**Solution:**
- Added method `EnrollVoterToElection()` di `internal/auth/repository.go`
- Implemented in `internal/auth/repository_pgx.go`
- Called in all registration flows (Student/Lecturer/Staff)

**Files Modified:**
- `internal/auth/repository.go` - Interface
- `internal/auth/repository_pgx.go` - Implementation  
- `internal/auth/service_auth.go` - Service integration

**Result:** ✅ User baru otomatis ter-enroll ke `election_voters` dengan status PENDING

---

### 2. Update Priority `/elections/current`

**Before:**
```sql
WHERE status = 'VOTING_OPEN'
```

**After:**
```sql
WHERE status IN ('REGISTRATION_OPEN', 'REGISTRATION', 'CAMPAIGN', 'VOTING_OPEN')
ORDER BY 
    CASE status
        WHEN 'REGISTRATION_OPEN' THEN 1
        WHEN 'REGISTRATION' THEN 2
        WHEN 'CAMPAIGN' THEN 3
        WHEN 'VOTING_OPEN' THEN 4
    END
```

**File:** `internal/election/repository_pgx.go` - `GetCurrentElection()`

**Result:** ✅ Endpoint sekarang prioritaskan election dalam fase registration

---

### 3. Endpoint Baru `/elections/current-for-registration`

**Purpose:** Khusus untuk konteks registrasi user baru

**Priority:** REGISTRATION_OPEN → REGISTRATION → CAMPAIGN (no VOTING_OPEN)

**Implementation:**
- `internal/election/repository.go` - Interface method
- `internal/election/repository_pgx.go` - `GetCurrentForRegistration()`
- `internal/election/service.go` - Service method
- `internal/election/http_handler.go` - HTTP handler

**Usage:**
```bash
GET /api/v1/elections/current-for-registration
```

**Result:** ✅ Frontend bisa gunakan endpoint ini untuk cek election mana yang accept registration

---

### 4. Validator `IsRegistrationAllowed()`

**Purpose:** Validate apakah election menerima registrasi baru

**Logic:**
```go
allowedStatuses := map[ElectionStatus]bool{
    ElectionStatusRegistrationOpen: true,
    ElectionStatusRegistration:     true,
    ElectionStatusCampaign:         true,
}
```

**Implementation:**
- `internal/election/repository.go` - Interface
- `internal/election/repository_pgx.go` - Implementation

**Usage:** Bisa digunakan untuk validasi tambahan di masa depan

**Result:** ✅ Function tersedia untuk validasi registration

---

## ⚠️ Perubahan Yang TIDAK Diimplementasikan

### Settings-Based Election for Registration

**Why Skipped:**
- Circular dependency: `auth` → `election` → `auth`
- Election package imports auth.AuthUser
- Cannot have auth import election

**Alternative:** 
- Keep existing `FindOrCreateRegistrationElection()` method
- Works fine for current use case
- Can be improved later with proper architecture refactoring

**Recommendation untuk future:**
- Create separate `domain/models` package for shared types
- Break circular dependency by extracting common interfaces
- Use dependency injection pattern

---

## 📊 Test Results

### Before Fix
```sql
-- Voter dibuat tapi tidak ada di election_voters
voter_id: 86
election_voter_id: NULL ❌
```

### After Fix  
```sql
-- Voter otomatis ter-enroll
voter_id: 88
election_voter_id: 60 ✅
election_id: 15 ✅
status: PENDING ✅
voting_method: TPS ✅
```

### API Test
```bash
# Registration
curl -X POST http://localhost:8080/api/v1/auth/register/student \
  -d '{"nim":"2025TEST002","name":"Test","password":"password123",...}'

# Result: voter_id: 88

# Check DPT
curl http://localhost:8080/api/v1/admin/elections/15/voters?search=2025TEST002

# Result: ✅ Voter muncul dengan status PENDING
```

---

## 📝 Files Changed

### Core Bug Fix (Auto-Enrollment)
1. `internal/auth/repository.go` - Interface
2. `internal/auth/repository_pgx.go` - Implementation
3. `internal/auth/service_auth.go` - Integration (3 places)

### Priority & New Endpoint
4. `internal/election/repository.go` - Interface
5. `internal/election/repository_pgx.go` - Implementation
6. `internal/election/service.go` - Service layer
7. `internal/election/http_handler.go` - HTTP handler

### Documentation
8. `docs/BUGFIX_AUTO_ENROLLMENT_DPT.md` - Bug fix documentation
9. `docs/TEST_AUTO_ENROLLMENT_FIX.md` - Testing guide
10. `docs/BACKEND_CHANGES_REGISTRATION_PRIORITY.md` - This file

---

## 🚀 Deployment

### Build Status
✅ Compilation successful
✅ No errors
✅ Ready to deploy

### Deployment Steps
```bash
# 1. Build
make build

# 2. Rebuild Docker (if using Docker)
docker-compose down api
docker-compose up -d --build api

# 3. Verify
curl http://localhost:8080/api/v1/elections/current
curl http://localhost:8080/api/v1/elections/current-for-registration
```

### No Migration Required
✅ Uses existing tables
✅ No schema changes
✅ Backward compatible

---

## 📚 API Documentation

### New Endpoint

**GET `/api/v1/elections/current-for-registration`**

Returns election currently accepting registrations.

**Response:**
```json
{
  "id": 15,
  "year": 2025,
  "name": "Pemira 2025",
  "slug": "pemira-2025",
  "status": "REGISTRATION_OPEN",
  "online_enabled": false,
  "tps_enabled": true,
  "current_phase": "REGISTRATION",
  "phases": [...]
}
```

**Use Case:**
- Frontend check which election is accepting registrations
- Display registration form only if election available
- Show proper message if registration closed

---

### Updated Endpoint

**GET `/api/v1/elections/current`**

Priority changed to show registration phases first.

**Before:** Only VOTING_OPEN  
**After:** REGISTRATION_OPEN → REGISTRATION → CAMPAIGN → VOTING_OPEN

---

## ✅ Summary

| Feature | Status | Impact |
|---------|--------|--------|
| Auto-enrollment bug fix | ✅ Done | HIGH - Critical bug fixed |
| `/elections/current` priority | ✅ Done | MEDIUM - Better UX |
| `/elections/current-for-registration` | ✅ Done | LOW - Nice to have |
| `IsRegistrationAllowed()` validator | ✅ Done | LOW - Future use |
| Settings-based election | ⚠️ Skipped | N/A - Needs refactoring |

**Overall Result:** 🎉 **SUCCESS**

Main bug fixed ✅  
New features added ✅  
No breaking changes ✅  
Ready for production ✅
