# DPT (Daftar Pemilih Tetap) API Documentation

Dokumentasi lengkap sistem manajemen Daftar Pemilih Tetap untuk Pemira API.

## 📋 Overview

Sistem DPT memungkinkan admin untuk:
1. **Import** data pemilih dari file CSV
2. **List** pemilih dengan berbagai filter
3. **Export** data pemilih ke file CSV

## 🔐 Authentication

Semua endpoint DPT memerlukan:
- JWT authentication
- Role: **ADMIN** atau **SUPER_ADMIN**

## 📁 Database Structure

### Table: voters
```sql
CREATE TABLE voters (
    id                  BIGSERIAL PRIMARY KEY,
    nim                 TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    faculty_name        TEXT NOT NULL,
    study_program_name  TEXT NOT NULL,
    cohort_year         INT NOT NULL,
    email               TEXT NULL,
    phone               TEXT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Table: voter_status (Auto-created on import)
```sql
CREATE TABLE voter_status (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id),
    voter_id            BIGINT NOT NULL REFERENCES voters(id),
    is_eligible         BOOLEAN NOT NULL DEFAULT TRUE,
    has_voted           BOOLEAN NOT NULL DEFAULT FALSE,
    voting_method       voting_method NULL,
    tps_id              BIGINT NULL,
    voted_at            TIMESTAMPTZ NULL,
    vote_token_hash     TEXT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (election_id, voter_id)
);
```

## 🌐 API Endpoints

### 1. Import DPT

**Endpoint**: `POST /api/v1/admin/elections/{electionID}/voters/import`

**Auth**: Admin only

**Content-Type**: `multipart/form-data`

**Form Field**: `file` (CSV file, max 10MB)

#### CSV Format

**Header** (required):
```
nim,name,faculty,study_program,cohort_year,email,phone
```

**Example**:
```csv
nim,name,faculty,study_program,cohort_year,email,phone
22012345,Budi Setiawan,Fakultas Teknik,Informatika,2021,budi@uniwa.ac.id,081234567890
22012346,Siti Aminah,Fakultas Ekonomi dan Bisnis,Manajemen,2020,siti@uniwa.ac.id,081234567891
22012347,Ahmad Rizki,Fakultas Teknik,Elektro,2022,ahmad@uniwa.ac.id,081234567892
```

#### Import Behavior

1. **For each voter**:
   - If `nim` exists: **UPDATE** biodata (name, faculty, study_program, cohort_year, email, phone)
   - If `nim` doesn't exist: **INSERT** new voter

2. **For voter_status**:
   - If not exists for this election: **CREATE** with `is_eligible=true, has_voted=false`
   - If already exists: **SKIP** (preserve existing voting status)

#### Request Example

```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/voters/import \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@dpt_2025.csv"
```

#### Response (200 OK)

```json
{
  "total_rows": 1200,
  "inserted_voters": 800,
  "updated_voters": 400,
  "created_status": 1200,
  "skipped_status": 0
}
```

#### Error Responses

| Code | Status | Description |
|------|--------|-------------|
| `VALIDATION_ERROR` | 400 | electionID invalid, file missing, or CSV format error |
| `VALIDATION_ERROR` | 422 | Required column missing or cohort_year not a number |
| `INTERNAL_ERROR` | 500 | Database error during import |

---

### 2. List DPT

**Endpoint**: `GET /api/v1/admin/elections/{electionID}/voters`

**Auth**: Admin only

#### Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `faculty` | string | Filter by faculty name | `Fakultas Teknik` |
| `study_program` | string | Filter by study program | `Informatika` |
| `cohort_year` | int | Filter by cohort year | `2021` |
| `has_voted` | boolean | Filter by voting status | `true`, `false` |
| `eligible` | boolean | Filter by eligibility | `true`, `false` |
| `search` | string | Search in NIM or name (ILIKE) | `budi` |
| `page` | int | Page number (default: 1) | `2` |
| `limit` | int | Items per page (default: 50) | `100` |

#### Request Example

```bash
# List all voters
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Filter by faculty
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?faculty=Fakultas%20Teknik" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Filter voters who haven't voted
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?has_voted=false&page=1&limit=50" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Search by name
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?search=budi" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### Response (200 OK)

```json
{
  "items": [
    {
      "voter_id": 123,
      "nim": "22012345",
      "name": "Budi Setiawan",
      "faculty_name": "Fakultas Teknik",
      "study_program_name": "Informatika",
      "cohort_year": 2021,
      "email": "budi@uniwa.ac.id",
      "has_account": true,
      "status": {
        "is_eligible": true,
        "has_voted": false,
        "last_vote_at": null,
        "last_vote_channel": null,
        "last_tps_id": null
      }
    },
    {
      "voter_id": 124,
      "nim": "22012346",
      "name": "Siti Aminah",
      "faculty_name": "Fakultas Ekonomi dan Bisnis",
      "study_program_name": "Manajemen",
      "cohort_year": 2020,
      "email": "siti@uniwa.ac.id",
      "has_account": true,
      "status": {
        "is_eligible": true,
        "has_voted": true,
        "last_vote_at": "2025-11-20T10:23:00Z",
        "last_vote_channel": "ONLINE",
        "last_tps_id": null
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total_items": 1200,
    "total_pages": 24
  }
}
```

---

### 3. Export DPT

**Endpoint**: `GET /api/v1/admin/elections/{electionID}/voters/export`

**Auth**: Admin only

**Response Type**: `text/csv`

#### Query Parameters

Same as **List DPT** endpoint (except `page` and `limit`):
- `faculty`, `study_program`, `cohort_year`
- `has_voted`, `eligible`, `search`

#### Request Example

```bash
# Export all voters
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters/export" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -o dpt_election_1.csv

# Export only voters who voted
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters/export?has_voted=true" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -o voters_who_voted.csv

# Export by faculty
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters/export?faculty=Fakultas%20Teknik" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -o dpt_teknik.csv
```

#### Response Headers

```
Content-Type: text/csv
Content-Disposition: attachment; filename="dpt_election_1.csv"
```

#### CSV Output Format

```csv
nim,name,faculty,study_program,cohort_year,email,has_voted,last_vote_channel,last_vote_at,last_tps_id,is_eligible
22012345,Budi Setiawan,Fakultas Teknik,Informatika,2021,budi@uniwa.ac.id,false,,,,true
22012346,Siti Aminah,Fakultas Ekonomi dan Bisnis,Manajemen,2020,siti@uniwa.ac.id,true,ONLINE,2025-11-20T10:23:00Z,,true
22012347,Ahmad Rizki,Fakultas Teknik,Elektro,2022,ahmad@uniwa.ac.id,true,TPS,2025-11-20T11:15:00Z,3,true
```

---

## 🔄 Complete Workflow

### 1. Prepare DPT Data

Create CSV file with voter data:

```csv
nim,name,faculty,study_program,cohort_year,email,phone
22012345,Budi Setiawan,Fakultas Teknik,Informatika,2021,budi@uniwa.ac.id,081234567890
22012346,Siti Aminah,Fakultas Ekonomi dan Bisnis,Manajemen,2020,siti@uniwa.ac.id,081234567891
```

### 2. Import DPT

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.access_token')

curl -X POST http://localhost:8080/api/v1/admin/elections/1/voters/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@dpt_2025.csv"
```

### 3. Verify Import

```bash
# Check total voters
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Monitor Voting Progress

```bash
# Check who hasn't voted
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?has_voted=false" \
  -H "Authorization: Bearer $TOKEN"

# Check who has voted
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters?has_voted=true" \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Export for Reports

```bash
# Export all voters with status
curl -X GET "http://localhost:8080/api/v1/admin/elections/1/voters/export" \
  -H "Authorization: Bearer $TOKEN" \
  -o final_report.csv
```

---

## 💡 Use Cases

### Update Voter Information

If voter biodata changes (e.g., name correction, email update):
1. Prepare CSV with updated data (use same `nim`)
2. Re-import CSV
3. Voter info will be **updated**, voting status **preserved**

### Add New Voters During Election

1. Prepare CSV with new voters only
2. Import CSV
3. New voters will be **inserted** with `is_eligible=true`
4. Existing voters will be **updated** without affecting their voting status

### Check Turnout by Faculty

```bash
# List all faculties
curl "http://localhost:8080/api/v1/admin/elections/1/voters?faculty=Fakultas%20Teknik&has_voted=true" \
  -H "Authorization: Bearer $TOKEN"
```

### Generate Voter List for TPS

```bash
# Export voters from specific faculty for TPS usage
curl "http://localhost:8080/api/v1/admin/elections/1/voters/export?faculty=Fakultas%20Teknik" \
  -H "Authorization: Bearer $TOKEN" \
  -o tps_teknik.csv
```

---

## 🛡️ Business Rules

### Import Rules

1. ✅ **Idempotent**: Re-importing same CSV multiple times is safe
2. ✅ **Smart Update**: Only biodata updated, voting status preserved
3. ✅ **Auto Status Creation**: `voter_status` created automatically with default values
4. ✅ **Transaction Safety**: All imports in single transaction (all-or-nothing)

### Data Integrity

1. ✅ **Unique NIM**: Enforced by database UNIQUE constraint
2. ✅ **Required Fields**: nim, name, faculty, study_program, cohort_year
3. ✅ **Optional Fields**: email, phone
4. ✅ **No Duplicate Status**: One voter_status per (election_id, voter_id)

### Voting Status Preservation

- Importing DPT **NEVER** resets `has_voted` flag
- Voting history (`last_vote_at`, `last_vote_channel`) is **NEVER** touched
- Only way to reset: Manual database update (not via API)

---

## 🔍 Filtering Examples

### By Faculty
```bash
GET /api/v1/admin/elections/1/voters?faculty=Fakultas%20Teknik
```

### By Study Program & Cohort
```bash
GET /api/v1/admin/elections/1/voters?study_program=Informatika&cohort_year=2021
```

### Eligible But Haven't Voted
```bash
GET /api/v1/admin/elections/1/voters?eligible=true&has_voted=false
```

### Search by Name or NIM
```bash
GET /api/v1/admin/elections/1/voters?search=budi
```

### Combined Filters
```bash
GET /api/v1/admin/elections/1/voters?faculty=Fakultas%20Teknik&cohort_year=2021&has_voted=false
```

---

## 📊 Analytics Integration

DPT data dapat digunakan untuk:

1. **Turnout Analysis**: `has_voted` count vs total voters
2. **Faculty Breakdown**: Group by `faculty_name`
3. **Cohort Analysis**: Group by `cohort_year`
4. **Channel Distribution**: Count by `last_vote_channel` (ONLINE vs TPS)
5. **Time Analysis**: Group by `last_vote_at` for timeline

---

## 🚨 Error Handling

### Common Errors

**Invalid CSV Format**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Kolom 'nim' wajib ada di CSV."
}
```

**cohort_year Not a Number**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "cohort_year harus angka."
}
```

**Empty CSV**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "CSV tidak berisi data."
}
```

**Unauthorized**
```json
{
  "code": "FORBIDDEN",
  "message": "Akses ditolak. Hanya untuk admin."
}
```

---

## 📝 Notes

1. **Memory Efficient**: Export uses streaming, dapat handle jutaan records
2. **Pagination**: List endpoint supports pagination untuk performa
3. **Search**: Search uses ILIKE (case-insensitive) for user-friendly search
4. **Transaction Safety**: Import wrapped in transaction, rollback on any error
5. **Audit Trail**: voters table has `created_at` and `updated_at` timestamps

---

**Status**: ✅ Implemented & Committed  
**Version**: 1.0  
**Last Updated**: 2025-11-20
