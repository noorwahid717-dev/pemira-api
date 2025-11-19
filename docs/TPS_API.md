# TPS Module API Documentation

Modul untuk manajemen TPS (Tempat Pemungutan Suara), QR statis per TPS, dan flow check-in mahasiswa.

## 🔐 Authentication

Semua endpoint memerlukan JWT Bearer token:

```
Authorization: Bearer <JWT>
Content-Type: application/json
```

### Roles:
- **ADMIN / PANITIA_UNI**: Kelola TPS, panitia, QR
- **STUDENT**: Scan QR dan check-in
- **TPS_OPERATOR**: Panel TPS untuk approve/reject pemilih

## 📋 Response Format

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "TPS_NOT_FOUND",
    "message": "TPS tidak ditemukan.",
    "details": null
  }
}
```

## 🔤 Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| UNAUTHORIZED | 401 | Token tidak ada/invalid |
| FORBIDDEN | 403 | Role tidak punya akses |
| TPS_NOT_FOUND | 404 | TPS tidak ditemukan |
| TPS_INACTIVE | 400 | TPS belum/tidak aktif |
| TPS_CLOSED | 400 | TPS sudah ditutup |
| QR_INVALID | 400 | Payload QR tidak valid |
| QR_REVOKED | 400 | QR sudah tidak berlaku |
| ELECTION_NOT_OPEN | 400 | Pemilu bukan di fase voting |
| NOT_ELIGIBLE | 400 | Mahasiswa bukan DPT |
| ALREADY_VOTED | 409 | Mahasiswa sudah voting |
| CHECKIN_NOT_FOUND | 404 | Check-in tidak ada |
| CHECKIN_NOT_PENDING | 400 | Check-in bukan status PENDING |
| CHECKIN_EXPIRED | 400 | Check-in sudah kadaluarsa |
| TPS_ACCESS_DENIED | 403 | Panitia tidak di-assign ke TPS |
| VALIDATION_ERROR | 422 | Request body invalid |
| INTERNAL_ERROR | 500 | Kesalahan server |

---

## 🧱 ADMIN: Manajemen TPS

### 1. List TPS

```http
GET /admin/tps
```

**Query Parameters:**
- `status` (optional): ACTIVE | DRAFT | CLOSED
- `election_id` (optional): Filter by election
- `page` (optional, default: 1)
- `limit` (optional, default: 20)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "code": "TPS01",
        "name": "TPS 1 – Aula Utama",
        "location": "Aula Utama, Gedung A Lt.1",
        "status": "ACTIVE",
        "voting_date": "2025-06-13",
        "open_time": "08:00",
        "close_time": "16:00",
        "total_votes": 832,
        "total_checkins": 910
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_items": 5,
      "total_pages": 1
    }
  }
}
```

### 2. Get Detail TPS

```http
GET /admin/tps/:id
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "election_id": 1,
    "code": "TPS01",
    "name": "TPS 1 – Aula Utama",
    "location": "Aula Utama, Gedung A Lt.1",
    "status": "ACTIVE",
    "voting_date": "2025-06-13",
    "open_time": "08:00",
    "close_time": "16:00",
    "capacity_estimate": 1000,
    "area_faculty": {
      "id": null,
      "name": null
    },
    "qr": {
      "id": 10,
      "qr_secret_suffix": "8392AF",
      "is_active": true,
      "created_at": "2025-06-10T08:00:00Z"
    },
    "stats": {
      "total_votes": 832,
      "total_checkins": 910,
      "pending_checkins": 3,
      "approved_checkins": 850,
      "rejected_checkins": 60
    },
    "panitia": [
      {
        "user_id": 101,
        "name": "Budi Santosa",
        "role": "KETUA_TPS"
      }
    ]
  }
}
```

### 3. Create TPS

```http
POST /admin/tps
```

**Request Body:**
```json
{
  "election_id": 1,
  "code": "TPS01",
  "name": "TPS 1 – Aula Utama",
  "location": "Aula Utama, Gedung A Lt.1",
  "voting_date": "2025-06-13",
  "open_time": "08:00",
  "close_time": "16:00",
  "capacity_estimate": 1000,
  "status": "DRAFT"
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "TPS01",
    "status": "DRAFT"
  }
}
```

### 4. Update TPS

```http
PUT /admin/tps/:id
```

**Request Body:**
```json
{
  "name": "TPS 1 – Aula Utama (Updated)",
  "location": "Aula Utama, Gedung A Lt.1",
  "voting_date": "2025-06-13",
  "open_time": "08:00",
  "close_time": "16:00",
  "capacity_estimate": 1000,
  "status": "ACTIVE"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "ACTIVE"
  }
}
```

### 5. Assign Panitia TPS

```http
PUT /admin/tps/:id/panitia
```

**Request Body:**
```json
{
  "members": [
    {
      "user_id": 101,
      "role": "KETUA_TPS"
    },
    {
      "user_id": 102,
      "role": "OPERATOR_PANEL"
    }
  ]
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "tps_id": 1,
    "total_members": 2
  }
}
```

### 6. Regenerate QR TPS

```http
POST /admin/tps/:id/qr/regenerate
```

Untuk emergency (QR bocor). Mark QR lama sebagai revoked dan generate baru.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "tps_id": 1,
    "qr": {
      "id": 11,
      "payload": "PEMIRA|TPS01|c9423e5f97d4",
      "created_at": "2025-06-12T20:10:00Z"
    }
  }
}
```

---

## 📱 STUDENT: Scan QR & Check-in

### 7. Scan QR di TPS

```http
POST /tps/checkin/scan
```

Dipanggil saat mahasiswa scan QR via kamera HP.

**Request Body:**
```json
{
  "qr_payload": "PEMIRA|TPS01|c9423e5f97d4"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "checkin_id": 555,
    "tps": {
      "id": 1,
      "code": "TPS01",
      "name": "TPS 1 – Aula Utama"
    },
    "status": "PENDING",
    "message": "Check-in berhasil. Silakan menunggu verifikasi panitia TPS."
  }
}
```

**Possible Errors:**
- 400 QR_INVALID
- 400 QR_REVOKED
- 400 TPS_INACTIVE / TPS_CLOSED
- 400 NOT_ELIGIBLE
- 409 ALREADY_VOTED

### 8. Check Status Check-in

```http
GET /tps/checkin/status?election_id=1
```

Polling endpoint untuk cek status check-in mahasiswa.

**Response - No Active Check-in:**
```json
{
  "success": true,
  "data": {
    "has_active_checkin": false
  }
}
```

**Response - Pending:**
```json
{
  "success": true,
  "data": {
    "has_active_checkin": true,
    "status": "PENDING",
    "tps": {
      "id": 1,
      "code": "TPS01",
      "name": "TPS 1 – Aula Utama"
    },
    "scan_at": "2025-06-13T09:20:00Z"
  }
}
```

**Response - Approved:**
```json
{
  "success": true,
  "data": {
    "has_active_checkin": true,
    "status": "APPROVED",
    "tps": {
      "id": 1,
      "code": "TPS01",
      "name": "TPS 1 – Aula Utama"
    },
    "approved_at": "2025-06-13T09:22:30Z",
    "expires_at": "2025-06-13T09:37:30Z"
  }
}
```

---

## 🖥️ TPS PANEL: Panitia TPS

### 9. Get TPS Summary

```http
GET /tps/:tps_id/summary
```

Header panel TPS dengan statistik.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "TPS01",
    "name": "TPS 1 – Aula Utama",
    "location": "Aula Utama, Gedung A Lt.1",
    "status": "ACTIVE",
    "voting_date": "2025-06-13",
    "open_time": "08:00",
    "close_time": "16:00",
    "stats": {
      "total_checkins": 910,
      "pending_checkins": 3,
      "approved_checkins": 850,
      "rejected_checkins": 60,
      "total_votes": 832
    }
  }
}
```

### 10. List Check-in Queue

```http
GET /tps/:tps_id/checkins?status=PENDING&page=1&limit=50
```

**Query Parameters:**
- `status` (default: PENDING): PENDING | APPROVED | REJECTED | USED
- `page` (default: 1)
- `limit` (default: 50)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 555,
        "voter": {
          "id": 123,
          "nim": "2110510023",
          "name": "Noah Febriyansyah",
          "faculty": "Teknik",
          "study_program": "Informatika",
          "cohort_year": 2021,
          "academic_status": "ACTIVE"
        },
        "status": "PENDING",
        "scan_at": "2025-06-13T09:20:00Z",
        "has_voted": false
      }
    ]
  }
}
```

### 11. Approve Check-in

```http
POST /tps/:tps_id/checkins/:checkin_id/approve
```

Setelah panitia cocokkan identitas fisik.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "checkin_id": 555,
    "status": "APPROVED",
    "voter": {
      "id": 123,
      "nim": "2110510023",
      "name": "Noah Febriyansyah"
    },
    "tps": {
      "id": 1,
      "code": "TPS01"
    },
    "approved_at": "2025-06-13T09:22:30Z"
  }
}
```

**Possible Errors:**
- 403 TPS_ACCESS_DENIED
- 404 CHECKIN_NOT_FOUND
- 400 CHECKIN_NOT_PENDING
- 409 ALREADY_VOTED

### 12. Reject Check-in

```http
POST /tps/:tps_id/checkins/:checkin_id/reject
```

**Request Body:**
```json
{
  "reason": "Data tidak cocok dengan identitas."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "checkin_id": 555,
    "status": "REJECTED",
    "reason": "Data tidak cocok dengan identitas."
  }
}
```

---

## 🛰️ WebSocket: Real-time Queue

### Connect to TPS Queue

```
GET /ws/tps/:tps_id/queue
```

**Auth**: JWT via query param `?token=<jwt>` atau header

**Server → Client Events:**

#### New Check-in
```json
{
  "type": "CHECKIN_NEW",
  "data": {
    "checkin_id": 555,
    "voter": {
      "nim": "2110510023",
      "name": "Noah Febriyansyah"
    },
    "scan_at": "2025-06-13T09:20:00Z"
  }
}
```

#### Check-in Updated
```json
{
  "type": "CHECKIN_UPDATED",
  "data": {
    "checkin_id": 555,
    "status": "APPROVED"
  }
}
```

---

## 🔗 Integration with Voting Module

### Flow Check-in → Voting:

1. **Mahasiswa scan QR** → `POST /tps/checkin/scan`
2. **Polling status** → `GET /tps/checkin/status` (atau via WS)
3. **Panitia approve** → `POST /tps/:tps_id/checkins/:id/approve`
4. **Mahasiswa bisa voting** → `POST /voting/tps/cast` (modul Voting)
5. **After voting** → Mark checkin as `USED`

### Validation di Voting Module:

```
- Check checkin.status == APPROVED
- Check checkin.expires_at > NOW()
- Check !voter.has_voted
```

---

## 📊 Database Schema

### tps
```sql
- id (PK)
- election_id (FK)
- code (unique)
- name
- location
- status (DRAFT|ACTIVE|CLOSED)
- voting_date
- open_time
- close_time
- capacity_estimate
- area_faculty_id (FK, nullable)
- created_at
- updated_at
```

### tps_qr
```sql
- id (PK)
- tps_id (FK)
- qr_secret_suffix
- is_active
- revoked_at
- created_at
```

### tps_panitia
```sql
- id (PK)
- tps_id (FK)
- user_id (FK)
- role (KETUA_TPS|OPERATOR_PANEL)
- created_at
```

### tps_checkins
```sql
- id (PK)
- tps_id (FK)
- voter_id (FK)
- election_id (FK)
- status (PENDING|APPROVED|REJECTED|USED|EXPIRED)
- scan_at
- approved_at
- approved_by_id (FK)
- rejection_reason
- expires_at
- created_at
- updated_at
```

---

## 🎯 Usage Examples

### Admin Flow:
1. Create TPS: `POST /admin/tps`
2. Assign panitia: `PUT /admin/tps/1/panitia`
3. Get QR for printing: Response includes `qr.payload`
4. Monitor: `GET /admin/tps/1`

### Student Flow:
1. Arrive at TPS, scan QR: `POST /tps/checkin/scan`
2. Wait for approval (poll): `GET /tps/checkin/status`
3. When APPROVED, vote: `POST /voting/tps/cast`

### Panitia TPS Flow:
1. Open panel: `GET /tps/1/summary`
2. See queue (real-time via WS): `GET /ws/tps/1/queue`
3. Verify identity, approve: `POST /tps/1/checkins/555/approve`
4. Or reject: `POST /tps/1/checkins/555/reject`

---

## 🔒 Security Notes

1. **QR Payload Format**: `PEMIRA|<TPS_CODE>|<SECRET>`
   - Secret should be cryptographically random
   - Rotate QR if compromised

2. **Check-in Expiry**: 
   - Approved check-ins expire in 15 minutes
   - Prevent replay attacks

3. **Access Control**:
   - TPS operations require role TPS_OPERATOR
   - Verify panitia assignment to TPS
   - Admin can access all TPS

4. **Rate Limiting**:
   - Limit scan attempts per voter
   - Prevent QR brute-force

5. **WebSocket Auth**:
   - Verify JWT on connection
   - Check TPS access permissions
   - Auto-disconnect on token expiry
