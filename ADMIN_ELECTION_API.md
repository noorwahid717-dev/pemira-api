# Admin Election Management API

API untuk manajemen pemilu oleh admin/panitia. Mencakup CRUD election, kontrol voting status, dan toggle mode voting (online/TPS).

## Authentication

Semua endpoint memerlukan:
- Header `Authorization: Bearer <access_token>`
- Role: **ADMIN** atau **PANITIA**

---

## Endpoints

### 1. List Elections

**GET** `/api/v1/admin/elections`

Mengambil daftar pemilu dengan filter dan pagination.

#### Query Parameters
| Parameter | Type   | Required | Description                          |
|-----------|--------|----------|--------------------------------------|
| year      | int    | No       | Filter by year                       |
| status    | string | No       | Filter by status (DRAFT, VOTING_OPEN, dll) |
| search    | string | No       | Search by name or slug               |
| page      | int    | No       | Page number (default: 1)             |
| limit     | int    | No       | Items per page (default: 20)         |

#### Response 200 OK
```json
{
  "items": [
    {
      "id": 1,
      "year": 2024,
      "name": "Pemilu Raya 2024",
      "slug": "pemira-2024",
      "status": "VOTING_OPEN",
      "voting_start_at": "2024-03-01T00:00:00Z",
      "voting_end_at": null,
      "online_enabled": true,
      "tps_enabled": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-03-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_items": 5,
    "total_pages": 1
  }
}
```

---

### 2. Get Election Detail

**GET** `/api/v1/admin/elections/{electionID}`

Mengambil detail pemilu berdasarkan ID.

#### Response 200 OK
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "VOTING_OPEN",
  "voting_start_at": "2024-03-01T00:00:00Z",
  "voting_end_at": null,
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-03-01T00:00:00Z"
}
```

#### Response 404 Not Found
```json
{
  "error": {
    "code": "ELECTION_NOT_FOUND",
    "message": "Pemilu tidak ditemukan."
  }
}
```

---

### 3. Create Election

**POST** `/api/v1/admin/elections`

Membuat pemilu baru dengan status DRAFT.

#### Request Body
```json
{
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "online_enabled": true,
  "tps_enabled": true
}
```

| Field          | Type   | Required | Description                      |
|----------------|--------|----------|----------------------------------|
| year           | int    | Yes      | Tahun pemilu                     |
| name           | string | Yes      | Nama pemilu                      |
| slug           | string | Yes      | Slug/code pemilu (unique)        |
| online_enabled | bool   | Yes      | Aktifkan voting online           |
| tps_enabled    | bool   | Yes      | Aktifkan voting TPS              |

#### Response 201 Created
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "DRAFT",
  "voting_start_at": null,
  "voting_end_at": null,
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

#### Response 422 Unprocessable Entity
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "year, name, dan slug wajib diisi."
  }
}
```

---

### 4. Update Election

**PUT** `/api/v1/admin/elections/{electionID}`

Mengupdate informasi pemilu. Semua field bersifat optional (partial update).

#### Request Body
```json
{
  "year": 2024,
  "name": "Pemilu Raya 2024 (Updated)",
  "slug": "pemira-2024-v2",
  "online_enabled": false,
  "tps_enabled": true
}
```

| Field          | Type   | Required | Description                      |
|----------------|--------|----------|----------------------------------|
| year           | int    | No       | Update tahun pemilu              |
| name           | string | No       | Update nama pemilu               |
| slug           | string | No       | Update slug pemilu               |
| online_enabled | bool   | No       | Toggle voting online             |
| tps_enabled    | bool   | No       | Toggle voting TPS                |

**Note**: Untuk toggle `online_enabled` dan `tps_enabled`, kirim field yang ingin diubah saja.

#### Response 200 OK
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024 (Updated)",
  "slug": "pemira-2024-v2",
  "status": "DRAFT",
  "voting_start_at": null,
  "voting_end_at": null,
  "online_enabled": false,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T11:30:00Z"
}
```

#### Example: Toggle Online Only
```json
{
  "online_enabled": false
}
```

---

### 5. Open Voting

**POST** `/api/v1/admin/elections/{electionID}/open-voting`

Membuka voting untuk pemilu (set status = VOTING_OPEN).

#### Business Rules
- Status pemilu harus **bukan** VOTING_OPEN
- Status pemilu harus **bukan** ARCHIVED
- Jika `voting_start_at` masih null, akan di-set ke waktu sekarang

#### Response 200 OK
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "VOTING_OPEN",
  "voting_start_at": "2024-03-01T00:00:00Z",
  "voting_end_at": null,
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-03-01T00:00:00Z"
}
```

#### Response 400 Bad Request
```json
{
  "error": {
    "code": "ELECTION_ALREADY_OPEN",
    "message": "Pemilu sudah dalam status voting terbuka."
  }
}
```

```json
{
  "error": {
    "code": "INVALID_STATUS_CHANGE",
    "message": "Status pemilu tidak dapat dibuka untuk voting."
  }
}
```

---

### 6. Close Voting

**POST** `/api/v1/admin/elections/{electionID}/close-voting`

Menutup voting untuk pemilu (set status = VOTING_CLOSED).

#### Business Rules
- Status pemilu harus VOTING_OPEN
- `voting_end_at` akan di-set ke waktu sekarang

#### Response 200 OK
```json
{
  "id": 1,
  "year": 2024,
  "name": "Pemilu Raya 2024",
  "slug": "pemira-2024",
  "status": "VOTING_CLOSED",
  "voting_start_at": "2024-03-01T00:00:00Z",
  "voting_end_at": "2024-03-05T23:59:59Z",
  "online_enabled": true,
  "tps_enabled": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-03-05T23:59:59Z"
}
```

#### Response 400 Bad Request
```json
{
  "error": {
    "code": "ELECTION_NOT_OPEN",
    "message": "Pemilu tidak dalam status voting terbuka."
  }
}
```

---

## Election Status Flow

```
DRAFT
  ↓
REGISTRATION (optional)
  ↓
CAMPAIGN (optional)
  ↓
VOTING_OPEN  ← Open Voting
  ↓
VOTING_CLOSED ← Close Voting
  ↓
CLOSED (optional)
  ↓
ARCHIVED
```

---

## Toggle Voting Mode

Voting mode (online/TPS) dapat di-toggle kapan saja menggunakan endpoint **Update Election**.

### Contoh Skenario

#### 1. Matikan Online, Aktifkan TPS Only
```bash
PUT /api/v1/admin/elections/1
{
  "online_enabled": false,
  "tps_enabled": true
}
```

#### 2. Aktifkan Hybrid (Online + TPS)
```bash
PUT /api/v1/admin/elections/1
{
  "online_enabled": true,
  "tps_enabled": true
}
```

#### 3. Online Only
```bash
PUT /api/v1/admin/elections/1
{
  "online_enabled": true,
  "tps_enabled": false
}
```

**Note**: Perubahan mode voting akan langsung mempengaruhi:
- Endpoint `/elections/current` (field `online_enabled`, `tps_enabled`)
- Endpoint `/elections/{id}/me/status` (field `online_allowed`, `tps_allowed`)
- Logika voting di `/voting/online/cast` dan `/voting/tps/cast`

---

## Integration dengan Existing Features

### 1. Public Election API
- `GET /api/v1/elections/current` → Menampilkan pemilu dengan status VOTING_OPEN
- Response includes `online_enabled` dan `tps_enabled`

### 2. Voter Status API
- `GET /api/v1/elections/{id}/me/status` → Menampilkan eligibility dan channel yang diperbolehkan
- Response includes `online_allowed` dan `tps_allowed` berdasarkan election setting

### 3. Voting API
- `POST /api/v1/voting/online/cast` → Cek `online_enabled` sebelum accept vote
- `POST /api/v1/voting/tps/cast` → Cek `tps_enabled` sebelum accept vote

### 4. Analytics & Monitoring
- Dashboard admin dapat filter berdasarkan election
- Real-time voting stats per election

---

## Testing dengan cURL

### 1. Create Election
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "year": 2024,
    "name": "Pemilu Raya 2024",
    "slug": "pemira-2024",
    "online_enabled": true,
    "tps_enabled": true
  }'
```

### 2. List Elections
```bash
curl -X GET "http://localhost:8080/api/v1/admin/elections?year=2024&page=1&limit=20" \
  -H "Authorization: Bearer <admin_token>"
```

### 3. Open Voting
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/open-voting \
  -H "Authorization: Bearer <admin_token>"
```

### 4. Toggle Mode
```bash
curl -X PUT http://localhost:8080/api/v1/admin/elections/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "online_enabled": false,
    "tps_enabled": true
  }'
```

### 5. Close Voting
```bash
curl -X POST http://localhost:8080/api/v1/admin/elections/1/close-voting \
  -H "Authorization: Bearer <admin_token>"
```

---

## Database Schema

Election table sudah mendukung fitur ini dengan kolom:
- `status`: Election status (DRAFT, VOTING_OPEN, VOTING_CLOSED, dll)
- `voting_start_at`: Timestamp voting dibuka
- `voting_end_at`: Timestamp voting ditutup
- `online_enabled`: Boolean flag untuk voting online
- `tps_enabled`: Boolean flag untuk voting TPS

---

## Next Steps

Setelah implementasi ini, frontend admin dapat:
1. ✅ Membuat pemilu baru
2. ✅ Mengelola info pemilu (nama, tahun, slug)
3. ✅ Toggle mode voting (online/TPS/hybrid) kapan saja
4. ✅ Membuka voting dengan satu klik
5. ✅ Menutup voting dengan satu klik
6. ✅ Melihat status pemilu real-time

Fitur lanjutan yang bisa ditambahkan:
- Schedule voting (set `voting_start_at` dan `voting_end_at` di masa depan)
- Archive election
- Clone election dari tahun sebelumnya
- Validation rule untuk status transition
