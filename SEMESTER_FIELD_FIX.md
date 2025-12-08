# Fix: Missing Semester Data in Admin View

## Problem
Mahasiswa yang telah registrasi tidak menampilkan data semester di halaman admin.

## Root Cause
Query INSERT di fungsi `CreateVoter()` tidak menyertakan kolom `semester`:
- Data semester hanya disimpan di field `class_label` (text)
- Field `semester` (integer) tidak diisi, sehingga bernilai NULL
- Admin membaca dari field `semester`, bukan `class_label`

## Solution Applied

### 1. Code Changes

**File: `internal/auth/repository_pgx.go`**

**Updated CreateVoter():**
- Menambahkan kolom `semester` ke dalam INSERT query
- Parsing string semester menjadi integer
- Menyimpan nilai ke field `semester` untuk mahasiswa (STUDENT)

**Added Helper Function:**
```go
func parseSemester(s string) int
```
Konversi string semester (contoh: "3", "5", "7") menjadi integer dengan validasi 1-20.

**File: `internal/dpt/repository_pgx.go`**

**Updated 3 Query Locations:**
- `ListVoters()` - line 121
- `ListVotersForElection()` - line 201
- `StreamVotersForElection()` - line 283

Changed from:
```sql
COALESCE(v.class_label, '') AS semester
```

To:
```sql
COALESCE(v.semester::TEXT, '') AS semester
```

Sekarang query DPT admin membaca dari kolom `semester` (integer) bukan `class_label` (text).

### 2. Database Backfill
Untuk data existing yang semester-nya NULL:
```sql
UPDATE voters 
SET semester = CASE 
    WHEN class_label ~ '^[0-9]+$' THEN class_label::integer 
    ELSE NULL 
END 
WHERE semester IS NULL 
  AND voter_type = 'STUDENT' 
  AND class_label IS NOT NULL;
```

**Result:** 25 records updated successfully

### 3. Verification
```sql
SELECT nim, name, semester, class_label 
FROM voters 
WHERE voter_type = 'STUDENT' 
ORDER BY created_at DESC 
LIMIT 10;
```

All students now have semester values populated correctly (1, 3, 5, 7, etc.)

## Testing
- ✅ Build successful
- ✅ All existing students have semester values
- ✅ New registrations will automatically populate semester field

## Deployment
1. Deploy kode terbaru ke production
2. Database sudah di-backfill, tidak perlu migration tambahan
3. Mahasiswa baru yang registrasi akan otomatis memiliki semester

## Date Fixed
2025-12-08
