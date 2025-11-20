# Admin TPS Management API

API untuk manajemen TPS (Tempat Pemungutan Suara) oleh admin/panitia. Mencakup CRUD TPS, manajemen operator, dan monitoring aktivitas voting per election.

## Authentication

Semua endpoint memerlukan:
- Header `Authorization: Bearer <access_token>`
- Role: **ADMIN** atau **PANITIA**

---

## Endpoints

### 1. TPS Management

#### 1.1. List All TPS

**GET** `/api/v1/admin/tps`

Mengambil daftar semua TPS.

**Response 200 OK**
```json
[
  {
    "id": 1,
    "code": "TPS01",
    "name": "TPS Aula Utama",
    "location": "Gedung A Lantai 1",
    "capacity": 500,
    "is_active": true,
    "open_time": "08:00",
    "close_time": "15:00",
    "pic_name": "John Doe",
    "pic_phone": "+62812345678",
    "has_active_qr": true,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  },
  {
    "id": 2,
    "code": "TPS02",
    "name": "TPS Fakultas Teknik",
    "location": "Gedung FT Lantai 2",
    "capacity": 300,
    "is_active": true,
    "open_time": "09:00",
    "close_time": "16:00",
    "pic_name": "Jane Smith",
    "pic_phone": "+62887654321",
    "has_active_qr": false,
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
]
```

---

#### 1.2. Get TPS Detail

**GET** `/api/v1/admin/tps/{tpsID}`

Mengambil detail TPS berdasarkan ID.

**Response 200 OK**
```json
{
  "id": 1,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "location": "Gedung A Lantai 1",
  "capacity": 500,
  "is_active": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Response 404 Not Found**
```json
{
  "code": "TPS_NOT_FOUND",
  "message": "TPS tidak ditemukan."
}
```

---

#### 1.3. Create TPS

**POST** `/api/v1/admin/tps`

Membuat TPS baru.

**Request Body**
```json
{
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "location": "Gedung A Lantai 1",
  "capacity": 500,
  "open_time": "08:00",
  "close_time": "15:00",
  "pic_name": "John Doe",
  "pic_phone": "+62812345678",
  "notes": "TPS utama, dekat kantin"
}
```

| Field      | Type   | Required | Description                            |
|------------|--------|----------|----------------------------------------|
| code       | string | Yes      | Kode TPS (unique, e.g. "TPS01")        |
| name       | string | Yes      | Nama TPS                               |
| location   | string | Yes      | Lokasi/alamat TPS                      |
| capacity   | int    | Yes      | Estimasi kapasitas pemilih             |
| open_time  | string | No       | Jam buka (format "HH:MM", default 08:00) |
| close_time | string | No       | Jam tutup (format "HH:MM", default 17:00) |
| pic_name   | string | No       | Nama penanggung jawab TPS              |
| pic_phone  | string | No       | Nomor kontak panitia TPS               |
| notes      | string | No       | Catatan internal                       |

**Response 201 Created**
```json
{
  "id": 1,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "location": "Gedung A Lantai 1",
  "capacity": 500,
  "is_active": true,
  "open_time": "08:00",
  "close_time": "15:00",
  "pic_name": "John Doe",
  "pic_phone": "+62812345678",
  "notes": "TPS utama, dekat kantin",
  "has_active_qr": false,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Response 422 Unprocessable Entity**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "code, name, dan location wajib diisi."
}
```

---

#### 1.4. Update TPS

**PUT** `/api/v1/admin/tps/{tpsID}`

Mengupdate informasi TPS. Semua field bersifat optional (partial update).

**Request Body**
```json
{
  "code": "TPS01A",
  "name": "TPS Aula Utama (Updated)",
  "location": "Gedung A Lantai 2",
  "capacity": 600,
  "is_active": false
}
```

| Field     | Type   | Required | Description                    |
|-----------|--------|----------|--------------------------------|
| code      | string | No       | Update kode TPS                |
| name      | string | No       | Update nama TPS                |
| location  | string | No       | Update lokasi TPS              |
| capacity  | int    | No       | Update kapasitas               |
| is_active | bool   | No       | Aktifkan/nonaktifkan TPS       |

**Response 200 OK**
```json
{
  "id": 1,
  "code": "TPS01A",
  "name": "TPS Aula Utama (Updated)",
  "location": "Gedung A Lantai 2",
  "capacity": 600,
  "is_active": false,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T12:30:00Z"
}
```

**Example: Deactivate TPS Only**
```json
{
  "is_active": false
}
```

---

#### 1.5. Delete TPS

**DELETE** `/api/v1/admin/tps/{tpsID}`

Menghapus TPS.

**Response 204 No Content**

**Response 404 Not Found**
```json
{
  "code": "TPS_NOT_FOUND",
  "message": "TPS tidak ditemukan."
}
```

---

### 2. QR Management

#### 2.1. Get QR Metadata

**GET** `/api/v1/admin/tps/{tpsID}/qr`

Mengambil metadata QR untuk TPS (status, token aktif, dll).

**Response 200 OK**
```json
{
  "tps_id": 3,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "active_qr": {
    "id": 12,
    "qr_token": "tps_qr_3_9Jsd8aKx...",
    "created_at": "2025-11-20T01:23:45Z"
  }
}
```

Jika belum ada QR aktif:
```json
{
  "tps_id": 3,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "active_qr": null
}
```

**Response 404 Not Found**
```json
{
  "code": "TPS_NOT_FOUND",
  "message": "TPS tidak ditemukan."
}
```

---

#### 2.2. Generate/Rotate QR

**POST** `/api/v1/admin/tps/{tpsID}/qr/rotate`

Generate QR baru atau rotate QR yang sudah ada. Digunakan untuk:
- Generate QR pertama kali untuk TPS baru
- Rotate QR jika bocor atau kompromi

**Request Body** (opsional, bisa kosong)
```json
{}
```

**Response 200 OK**
```json
{
  "tps_id": 3,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "active_qr": {
    "id": 13,
    "qr_token": "tps_qr_3_new_AbCdEf...",
    "created_at": "2025-11-20T02:00:00Z"
  }
}
```

**Response 404 Not Found**
```json
{
  "code": "TPS_NOT_FOUND",
  "message": "TPS tidak ditemukan."
}
```

**Notes:**
- QR lama akan di-set `is_active = false` dan `rotated_at = NOW()`
- Token baru di-generate secara random menggunakan crypto/rand
- Token format: `tps_qr_{tpsID}_{random32chars}`

---

#### 2.3. Get QR for Print

**GET** `/api/v1/admin/tps/{tpsID}/qr/print`

Mengambil payload QR dalam format siap cetak. Frontend dapat menggunakan library QR code untuk render.

**Response 200 OK**
```json
{
  "tps_id": 3,
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "qr_payload": "pemira://tps-checkin?t=tps_qr_3_new_AbCdEf..."
}
```

**Response 404 Not Found**
```json
{
  "code": "TPS_NOT_FOUND",
  "message": "TPS tidak ditemukan."
}
```

**Response 500 Internal Server Error** (jika tidak ada QR aktif)
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Gagal mengambil data cetak QR."
}
```

**Use Case:**
1. Admin buka halaman cetak QR
2. Frontend render QR code dari `qr_payload`
3. Admin print QR untuk ditempel di TPS
4. Format QR: `pemira://tps-checkin?t={token}`

---

### 3. Operator Management

#### 2.1. List Operators

**GET** `/api/v1/admin/tps/{tpsID}/operators`

Mengambil daftar operator untuk TPS tertentu.

**Response 200 OK**
```json
[
  {
    "user_id": 10,
    "username": "operator.tps01",
    "name": "John Doe",
    "email": "john@example.com"
  },
  {
    "user_id": 11,
    "username": "operator2.tps01",
    "name": "Jane Smith",
    "email": "jane@example.com"
  }
]
```

---

#### 2.2. Create Operator

**POST** `/api/v1/admin/tps/{tpsID}/operators`

Membuat akun operator baru untuk TPS.

**Request Body**
```json
{
  "username": "operator.tps01",
  "password": "SecurePassword123!",
  "name": "John Doe",
  "email": "john@example.com"
}
```

| Field    | Type   | Required | Description                     |
|----------|--------|----------|---------------------------------|
| username | string | Yes      | Username untuk login            |
| password | string | Yes      | Password (akan di-hash)         |
| name     | string | No       | Nama lengkap operator           |
| email    | string | No       | Email operator                  |

**Response 201 Created**
```json
{
  "user_id": 10,
  "username": "operator.tps01",
  "name": "John Doe",
  "email": "john@example.com"
}
```

**Response 422 Unprocessable Entity**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "username dan password wajib diisi."
}
```

**Notes:**
- Password dikirim dalam plaintext, pastikan menggunakan HTTPS
- Password akan di-hash dengan bcrypt sebelum disimpan
- Operator akan memiliki role `TPS_OPERATOR` dan `tps_id` sesuai TPS

---

#### 2.3. Remove Operator

**DELETE** `/api/v1/admin/tps/{tpsID}/operators/{userID}`

Menghapus operator dari TPS.

**Response 204 No Content**

---

### 4. Monitoring

#### 4.1. Monitor TPS per Election

**GET** `/api/v1/admin/elections/{electionID}/tps/monitor`

Mengambil data monitoring untuk semua TPS dalam suatu election.

**Response 200 OK**
```json
[
  {
    "tps_id": 1,
    "code": "TPS01",
    "name": "TPS Aula Utama",
    "location": "Gedung A Lantai 1",
    "total_checkins": 45,
    "approved_checkins": 42,
    "total_votes": 40,
    "last_activity_at": "2024-03-01T14:30:00Z"
  },
  {
    "tps_id": 2,
    "code": "TPS02",
    "name": "TPS Fakultas Teknik",
    "location": "Gedung FT Lantai 2",
    "total_checkins": 30,
    "approved_checkins": 28,
    "total_votes": 25,
    "last_activity_at": "2024-03-01T14:15:00Z"
  }
]
```

**Fields:**
- `total_checkins`: Total scan QR (semua status)
- `approved_checkins`: Total checkin yang di-approve
- `total_votes`: Total suara yang sudah masuk (channel TPS)
- `last_activity_at`: Waktu aktivitas terakhir (checkin/vote)

---

## Use Cases

### Scenario 1: Setup TPS untuk Election

```bash
# 1. Create TPS
POST /admin/tps
{
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "location": "Gedung A Lantai 1",
  "capacity": 500
}

# 2. Create Operators (bisa lebih dari 1)
POST /admin/tps/1/operators
{
  "username": "operator.tps01",
  "password": "SecurePass123!",
  "name": "John Doe",
  "email": "john@example.com"
}

POST /admin/tps/1/operators
{
  "username": "operator2.tps01",
  "password": "SecurePass456!",
  "name": "Jane Smith",
  "email": "jane@example.com"
}

# 3. List operators
GET /admin/tps/1/operators
```

### Scenario 2: Monitor Voting Activity

```bash
# Monitor all TPS for election
GET /admin/elections/1/tps/monitor

# Response shows:
# - Which TPS has most checkins
# - Checkin approval rate
# - Voting participation rate
# - Last activity timestamp
```

### Scenario 3: Deactivate TPS

```bash
# Temporarily deactivate TPS
PUT /admin/tps/1
{
  "is_active": false
}

# Reactivate later
PUT /admin/tps/1
{
  "is_active": true
}
```

### Scenario 4: Replace Operator

```bash
# Remove old operator
DELETE /admin/tps/1/operators/10

# Add new operator
POST /admin/tps/1/operators
{
  "username": "operator.new.tps01",
  "password": "NewPassword123!",
  "name": "New Operator",
  "email": "new@example.com"
}
```

---

## Integration dengan TPS Workflow

### Complete TPS Flow:

1. **Admin Setup**
   ```bash
   POST /admin/tps → Create TPS
   POST /admin/tps/{id}/operators → Create operator accounts
   ```

2. **Generate QR Code** (Future feature or external)
   - Static QR code untuk TPS
   - Format: `TPS_CODE:SECRET_SUFFIX`

3. **Operator Login**
   ```bash
   POST /auth/login → Operator login dengan credentials
   ```

4. **Student Checkin** (existing flow)
   ```bash
   # Student scans QR at TPS
   POST /tps/checkin/scan → Mahasiswa scan QR
   
   # Operator approves
   POST /tps/checkin/{id}/approve → Operator approve checkin
   ```

5. **Student Votes** (existing flow)
   ```bash
   POST /voting/tps/cast → Mahasiswa voting di TPS
   ```

6. **Admin Monitor**
   ```bash
   GET /admin/elections/1/tps/monitor → Real-time monitoring
   ```

---

## Database Schema

### TPS Table
```sql
CREATE TABLE tps (
    id                BIGSERIAL PRIMARY KEY,
    election_id       BIGINT NOT NULL,
    code              VARCHAR(20) NOT NULL UNIQUE,
    name              VARCHAR(255) NOT NULL,
    location          TEXT NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    capacity_estimate INTEGER DEFAULT 0,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### User Accounts (Operators)
```sql
CREATE TABLE user_accounts (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          VARCHAR(20) NOT NULL, -- 'TPS_OPERATOR'
    tps_id        BIGINT REFERENCES tps(id),
    voter_id      BIGINT REFERENCES voters(id),
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### TPS Checkins
```sql
CREATE TABLE tps_checkins (
    id                BIGSERIAL PRIMARY KEY,
    tps_id            BIGINT NOT NULL REFERENCES tps(id),
    voter_id          BIGINT NOT NULL,
    election_id       BIGINT NOT NULL,
    status            VARCHAR(20) NOT NULL, -- PENDING, APPROVED, REJECTED
    scan_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at       TIMESTAMP,
    approved_by_id    BIGINT,
    rejection_reason  TEXT,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Votes (TPS Channel)
```sql
CREATE TABLE votes (
    id          BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL,
    voter_id    BIGINT NOT NULL,
    channel     VARCHAR(20) NOT NULL, -- 'TPS' or 'ONLINE'
    tps_id      BIGINT REFERENCES tps(id),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

---

## Testing dengan cURL

### 1. Create TPS
```bash
curl -X POST http://localhost:8080/api/v1/admin/tps \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TPS01",
    "name": "TPS Aula Utama",
    "location": "Gedung A Lantai 1",
    "capacity": 500
  }'
```

### 2. List TPS
```bash
curl http://localhost:8080/api/v1/admin/tps \
  -H "Authorization: Bearer <admin_token>"
```

### 3. Create Operator
```bash
curl -X POST http://localhost:8080/api/v1/admin/tps/1/operators \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "operator.tps01",
    "password": "SecurePassword123!",
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### 4. List Operators
```bash
curl http://localhost:8080/api/v1/admin/tps/1/operators \
  -H "Authorization: Bearer <admin_token>"
```

### 5. Monitor TPS
```bash
curl http://localhost:8080/api/v1/admin/elections/1/tps/monitor \
  -H "Authorization: Bearer <admin_token>"
```

### 6. Deactivate TPS
```bash
curl -X PUT http://localhost:8080/api/v1/admin/tps/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}'
```

### 7. Delete Operator
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/tps/1/operators/10 \
  -H "Authorization: Bearer <admin_token>"
```

### 8. Delete TPS
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/tps/1 \
  -H "Authorization: Bearer <admin_token>"
```

---

## Security Notes

1. **Password Handling**
   - Password dikirim plaintext via API (use HTTPS!)
   - Password di-hash dengan bcrypt sebelum disimpan
   - Cost factor: bcrypt.DefaultCost (10)

2. **Access Control**
   - Hanya ADMIN yang bisa akses semua endpoints
   - Operator TPS hanya bisa akses TPS yang assigned

3. **Best Practices**
   - Generate strong password untuk operator
   - Rotate password secara berkala
   - Deactivate operator setelah election selesai
   - Audit log untuk tracking perubahan

---

## Monitoring Metrics

Admin dapat track:
- **TPS Performance**: Checkin vs vote ratio
- **Operator Activity**: Last activity timestamp
- **Queue Status**: Pending checkins
- **Approval Rate**: Approved vs total checkins
- **Voting Participation**: Total votes per TPS

---

## Next Steps

Setelah implementasi ini, frontend admin dapat:
1. ✅ Setup TPS untuk election
2. ✅ Buat akun operator TPS
3. ✅ Monitor aktivitas real-time per TPS
4. ✅ Deactivate TPS jika ada masalah
5. ✅ Manage operator (add/remove)
6. ✅ Track voting progress per location

Fitur lanjutan yang bisa ditambahkan:
- [ ] QR code generation API
- [ ] TPS assignment per election
- [ ] Operator shift management
- [ ] Real-time dashboard WebSocket
- [ ] Export monitoring report
- [ ] Bulk TPS import (CSV/Excel)
