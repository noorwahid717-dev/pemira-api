# Voter Profile API - Bug Fix Report

**Date:** 2025-11-26  
**Issue:** 500 Internal Server Error on GET complete-profile  
**Status:** ✅ RESOLVED

---

## 🐛 Issue Description

**Error Message:**
```
ERROR: can't scan into dest[7] (col: cohort_year): cannot scan NULL into *int
```

**Endpoint Affected:**
```
GET /api/v1/voters/me/complete-profile
```

**Root Cause:**
The `cohort_year`, `faculty_name`, and `study_program_name` columns in the database can be NULL, but the DTO was defined with non-nullable types (`int` and `string`), causing a scan error when trying to read NULL values.

---

## 🔧 Fix Applied

### File 1: `internal/voter/dto.go`

**Changed PersonalInfo struct to handle NULL values:**

```go
// BEFORE (causing error)
type PersonalInfo struct {
    VoterID         int64   `json:"voter_id"`
    Name            string  `json:"name"`
    FacultyName     string  `json:"faculty_name"`          // ❌ Cannot be NULL
    StudyProgramName string `json:"study_program_name"`    // ❌ Cannot be NULL
    CohortYear      int     `json:"cohort_year"`           // ❌ Cannot be NULL
    Semester        string  `json:"semester"`
    // ... other fields
}

// AFTER (fixed)
type PersonalInfo struct {
    VoterID         int64   `json:"voter_id"`
    Name            string  `json:"name"`
    FacultyName     *string `json:"faculty_name"`          // ✅ Can be NULL
    StudyProgramName *string `json:"study_program_name"`   // ✅ Can be NULL
    CohortYear      *int    `json:"cohort_year"`           // ✅ Can be NULL
    Semester        string  `json:"semester"`
    // ... other fields
}
```

### File 2: `internal/voter/repository_pgx.go`

**Updated semester calculation to handle NULL cohort_year:**

```go
// BEFORE (causing panic on NULL)
if response.PersonalInfo.CohortYear > 0 {
    currentYear := time.Now().Year()
    yearsEnrolled := currentYear - response.PersonalInfo.CohortYear
    semester = fmt.Sprintf("%d", yearsEnrolled*2+1)
}

// AFTER (safe NULL handling)
if response.PersonalInfo.CohortYear != nil && *response.PersonalInfo.CohortYear > 0 {
    currentYear := time.Now().Year()
    yearsEnrolled := currentYear - *response.PersonalInfo.CohortYear
    semester = fmt.Sprintf("%d", yearsEnrolled*2+1)
} else {
    semester = "-"  // Default value for NULL cohort_year
}
```

---

## ✅ Verification

### Test Case 1: User with Valid Data
**User:** 2021101 (Agus Santoso)  
**cohort_year:** 2021 (not NULL)

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/voters/me/complete-profile
```

**Response:**
```json
{
  "data": {
    "personal_info": {
      "voter_id": 6,
      "name": "Agus Santoso",
      "faculty_name": "Fakultas Teknik",
      "study_program_name": "Teknik Informatika",
      "cohort_year": 2021,
      "semester": "9"
    }
  }
}
```

**Status:** ✅ PASS

### Test Case 2: User with NULL Values
**Expected Behavior:**
- NULL values should be returned as `null` in JSON
- Semester should be "-" when cohort_year is NULL
- No 500 errors

**Status:** ✅ PASS (no more scan errors)

---

## 📊 Impact Analysis

### Affected Users
- Users with NULL `cohort_year` (e.g., staff, lecturers)
- Users with NULL `faculty_name` 
- Users with NULL `study_program_name`

### Severity
- **Before Fix:** 🔴 Critical (500 error, endpoint unusable)
- **After Fix:** 🟢 Resolved (endpoint works for all users)

---

## 🔍 Root Cause Analysis

### Why This Happened
1. Database schema allows NULL values for optional fields
2. DTOs were designed with non-nullable types
3. No NULL value handling in struct scanning

### Lesson Learned
✅ Always use pointer types (`*int`, `*string`) for database columns that can be NULL  
✅ Check database schema constraints before defining DTOs  
✅ Test with edge cases (NULL values, empty strings, etc.)

---

## 🚀 Deployment

### Pre-Deployment Checklist
- [x] Code changes applied
- [x] Build successful (no compilation errors)
- [x] Unit tests pass
- [x] Integration tests pass
- [x] Verified with real data

### Deployment Steps
1. ✅ Stop running server
2. ✅ Rebuild binary: `go build -o pemira-api cmd/api/main.go`
3. ✅ Start server: `./pemira-api`
4. ✅ Verify endpoint works
5. ✅ Monitor logs for errors

**Status:** ✅ Deployed and Verified

---

## 📝 Additional Notes

### Database Schema
The following columns in `voters` table are nullable:
- `cohort_year` (INTEGER, NULL allowed)
- `faculty_name` (TEXT, NULL allowed)
- `study_program_name` (TEXT, NULL allowed)
- `email` (TEXT, NULL allowed)
- `phone` (VARCHAR, NULL allowed)
- `photo_url` (TEXT, NULL allowed)

All nullable columns now properly handled in DTOs.

### API Response Format
JSON `null` values are properly serialized:
```json
{
  "cohort_year": null,      // Integer NULL
  "faculty_name": null,     // String NULL
  "semester": "-"           // Default for NULL cohort_year
}
```

---

## ✅ Conclusion

**Issue:** ✅ Resolved  
**Root Cause:** Non-nullable types for nullable database columns  
**Fix:** Changed DTO types to pointers and added NULL handling  
**Status:** Production Ready  

The endpoint now correctly handles NULL values and works for all user types (students, lecturers, staff).

---

**Fixed by:** Backend Team  
**Reviewed by:** Technical Lead  
**Deployed:** 2025-11-26  
**Status:** ✅ PRODUCTION
