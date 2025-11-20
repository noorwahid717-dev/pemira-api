# LECTURER & STAFF IMPLEMENTATION

## Overview
Sistem PEMIRA kini mendukung **dosen (lecturer)** dan **staff universitas** sebagai pemilih selain mahasiswa.

## Database Schema

### Tabel Lecturers (Dosen)
```sql
CREATE TABLE lecturers (
    id                      BIGSERIAL PRIMARY KEY,
    nidn                    TEXT NOT NULL,      -- Nomor Induk Dosen Nasional
    name                    TEXT NOT NULL,
    email                   TEXT NULL,
    faculty_code            TEXT NULL,
    faculty_name            TEXT NULL,
    department_code         TEXT NULL,
    department_name         TEXT NULL,
    position                TEXT NULL,          -- Jabatan: Lektor, Asisten Ahli, dst
    employment_status       TEXT NULL,          -- Status: Tetap, Tidak Tetap
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Tabel Staff Members (Staff/Tenaga Kependidikan)
```sql
CREATE TABLE staff_members (
    id                      BIGSERIAL PRIMARY KEY,
    nip                     TEXT NOT NULL,      -- Nomor Induk Pegawai
    name                    TEXT NOT NULL,
    email                   TEXT NULL,
    unit_code               TEXT NULL,          -- Unit kerja
    unit_name               TEXT NULL,
    position                TEXT NULL,          -- Jabatan
    employment_status       TEXT NULL,          -- Status kepegawaian
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### User Accounts Update
Tabel `user_accounts` ditambahkan kolom:
- `lecturer_id` BIGINT NULL REFERENCES lecturers(id)
- `staff_id` BIGINT NULL REFERENCES staff_members(id)

### User Roles
Enum `user_role` ditambahkan:
- `LECTURER` - untuk dosen
- `STAFF` - untuk staff/tenaga kependidikan

## Authentication

### Login Credentials

**Dosen:**
- Username: NIDN (Nomor Induk Dosen Nasional)
- Password: sesuai yang di-set

**Staff:**
- Username: NIP (Nomor Induk Pegawai)
- Password: sesuai yang di-set

### Login Response

**Lecturer Login:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "LBstOYdU...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": 11,
    "username": "0101018901",
    "role": "LECTURER",
    "lecturer_id": 1,
    "profile": {
      "name": "Dr. Ahmad Kusuma, S.Kom., M.T.",
      "faculty_name": "Fakultas Teknik",
      "department_name": "Teknik Informatika",
      "position": "Lektor Kepala"
    }
  }
}
```

**Staff Login:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "OsfgCWYx...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": 16,
    "username": "198901012015041001",
    "role": "STAFF",
    "staff_id": 1,
    "profile": {
      "name": "Bambang Setiawan, S.Sos.",
      "position": "Kepala Sub Bagian Umum",
      "unit_name": "Biro Administrasi Umum"
    }
  }
}
```

## JWT Claims

Token JWT sekarang include:
- `lecturer_id` (jika user adalah dosen)
- `staff_id` (jika user adalah staff)

## Seed Data

File: `seed_lecturers_staff.sql`

**Lecturers (5 dosen):**
1. `0101018901` - Dr. Ahmad Kusuma (Teknik Informatika)
2. `0102019002` - Dra. Siti Nurjanah (PGSD)
3. `0103019103` - Prof. Dr. Budi Santoso (Manajemen)
4. `0104019204` - Dr. Retno Wulandari (Matematika)
5. `0105019305` - Ir. Joko Widodo (Teknik Sipil)

**Staff (5 staff):**
1. `198901012015041001` - Bambang Setiawan (BAU)
2. `199002012016051002` - Dewi Kusumawati (BAAK)
3. `199103012017061003` - Eko Prasetyo (UPT-TIK)
4. `199204012018071004` - Fitri Handayani (BAK)
5. `199305012019081005` - Gunawan Wijaya (Perpustakaan)

**Default Password:** `password123`

## API Endpoints

### POST /api/v1/auth/login
Login untuk semua role (STUDENT, LECTURER, STAFF, ADMIN, dll)

**Request:**
```json
{
  "username": "0101018901",  // NIDN untuk lecturer, NIP untuk staff
  "password": "password123"
}
```

### GET /api/v1/auth/me
Mendapatkan info user yang sedang login

**Response untuk Lecturer:**
```json
{
  "id": 11,
  "username": "0101018901",
  "role": "LECTURER",
  "profile": {
    "name": "Dr. Ahmad Kusuma, S.Kom., M.T.",
    "faculty_name": "Fakultas Teknik",
    "department_name": "Teknik Informatika",
    "position": "Lektor Kepala"
  }
}
```

## Testing

```bash
# Login sebagai dosen
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "0101018901", "password": "password123"}'

# Login sebagai staff
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "198901012015041001", "password": "password123"}'

# Get current user info
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

## Migration

File migration: `migrations/008_add_lecturer_staff_support.up.sql`

Jalankan dengan:
```bash
PGPASSWORD=pemira psql -h localhost -U pemira -d pemira \
  -f migrations/008_add_lecturer_staff_support.up.sql
```

Atau rollback dengan:
```bash
PGPASSWORD=pemira psql -h localhost -U pemira -d pemira \
  -f migrations/008_add_lecturer_staff_support.down.sql
```

## Implementation Details

**Files Modified:**
1. `internal/shared/constants/constants.go` - Added RoleLecturer & RoleStaff
2. `internal/auth/model.go` - Added lecturer_id & staff_id fields
3. `internal/auth/repository_pgx.go` - Updated all queries & GetUserProfile
4. `internal/auth/service_auth.go` - Updated AuthUser response
5. `internal/auth/jwt.go` - Added lecturer_id & staff_id to JWT claims

**Files Created:**
1. `migrations/008_add_lecturer_staff_support.up.sql` - Migration up
2. `migrations/008_add_lecturer_staff_support.down.sql` - Migration down
3. `seed_lecturers_staff.sql` - Seed data for testing

## Summary

✅ Dosen dapat login dengan NIDN
✅ Staff dapat login dengan NIP
✅ Profile data lengkap (fakultas/unit, jabatan, dll)
✅ JWT claims include lecturer_id/staff_id
✅ Seed data tersedia untuk testing
✅ All endpoints tested successfully
