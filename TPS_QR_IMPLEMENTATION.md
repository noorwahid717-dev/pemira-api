# TPS QR dan PIC Implementation Summary

Implementasi fitur TPS QR management dan informasi panitia (PIC) untuk PEMIRA API.

## ✅ Completed Changes

### 1. Database Schema
- [x] Migrasi 006: Tambah kolom `pic_name`, `pic_phone`, `notes` di tabel `tps`
- [x] Rename `qr_secret` → `qr_token` di tabel `tps_qr`
- [x] Rename `revoked_at` → `rotated_at` di tabel `tps_qr`
- [x] Tambah unique constraint untuk one active QR per TPS

**File:** `migrations/006_add_tps_pic_fields.up.sql` & `.down.sql`

### 2. Entity & DTOs

#### TPS Entity
- [x] Tambah fields: `PICName`, `PICPhone`, `Notes`
- [x] Update TPSQR entity: `QRToken`, `RotatedAt`

**File:** `internal/tps/entity.go`

#### Admin DTOs
- [x] Update `TPSDTO` dengan PIC fields dan `HasActiveQR`
- [x] Update `TPSCreateRequest` dengan PIC fields
- [x] Update `TPSUpdateRequest` dengan PIC fields
- [x] Update `TPSMonitorDTO` dengan PIC fields dan jam operasional
- [x] Tambah `TPSQRMetadataResponse`
- [x] Tambah `TPSQRRotateResponse`
- [x] Tambah `TPSQRPrintResponse`
- [x] Tambah `ActiveQRDTO`

**Files:** 
- `internal/tps/admin_model.go`
- `internal/tps/dto.go`

### 3. Repository Layer

#### Admin Repository Interface
- [x] Tambah method: `GetTPSQRMetadata()`
- [x] Tambah method: `RotateTPSQR()`
- [x] Tambah method: `GetTPSQRForPrint()`

**File:** `internal/tps/admin_repository.go`

#### PgAdmin Repository Implementation
- [x] Update query `List()` untuk include PIC fields dan `has_active_qr`
- [x] Update query `GetByID()` untuk include PIC fields
- [x] Update query `Create()` untuk handle PIC fields
- [x] Update query `Update()` untuk handle PIC fields
- [x] Implementasi `GetTPSQRMetadata()`
- [x] Implementasi `RotateTPSQR()` dengan transaction
- [x] Implementasi `GetTPSQRForPrint()`
- [x] Helper function `generateQRToken()` dengan crypto/rand

**File:** `internal/tps/admin_repository_pgx.go`

### 4. Service Layer
- [x] Tambah method: `GetQRMetadata()`
- [x] Tambah method: `RotateQR()`
- [x] Tambah method: `GetQRForPrint()`

**File:** `internal/tps/admin_service.go`

### 5. HTTP Handlers
- [x] Handler: `GetQRMetadata()` - GET `/admin/tps/{tpsID}/qr`
- [x] Handler: `RotateQR()` - POST `/admin/tps/{tpsID}/qr/rotate`
- [x] Handler: `GetQRForPrint()` - GET `/admin/tps/{tpsID}/qr/print`

**File:** `internal/tps/admin_http_handler.go`

### 6. Routes Registration
- [x] Register QR management routes di main.go
  - GET `/admin/tps/{tpsID}/qr`
  - POST `/admin/tps/{tpsID}/qr/rotate`
  - GET `/admin/tps/{tpsID}/qr/print`

**File:** `cmd/api/main.go`

### 7. Documentation
- [x] Update ADMIN_TPS_API.md
  - Tambah section "QR Management"
  - Update request/response examples dengan PIC fields
  - Update list response dengan `has_active_qr`
  - Dokumentasi endpoint QR lengkap dengan use cases

**File:** `ADMIN_TPS_API.md`

---

## 📋 API Endpoints Summary

### TPS CRUD (Updated)
```
GET    /api/v1/admin/tps              # List all TPS (with PIC & QR status)
GET    /api/v1/admin/tps/{tpsID}      # Get TPS detail
POST   /api/v1/admin/tps              # Create TPS (with PIC fields)
PUT    /api/v1/admin/tps/{tpsID}      # Update TPS (with PIC fields)
DELETE /api/v1/admin/tps/{tpsID}      # Delete TPS
```

### QR Management (New)
```
GET    /api/v1/admin/tps/{tpsID}/qr          # Get QR metadata
POST   /api/v1/admin/tps/{tpsID}/qr/rotate   # Generate/rotate QR
GET    /api/v1/admin/tps/{tpsID}/qr/print    # Get QR for printing
```

---

## 🔑 Key Features

### 1. Static QR per TPS dengan Rotation
- QR code statis yang bisa dicetak
- Token tersimpan di database
- Bisa di-rotate kalau bocor
- One active QR per TPS (constraint di DB)

### 2. QR Token Format
```
tps_qr_{tpsID}_{random32chars}

Example:
tps_qr_3_9Jsd8aKxL2m4nP6qR8tU0vW2yZ4b
```

### 3. QR Payload untuk Checkin
```
pemira://tps-checkin?t={qr_token}

Example:
pemira://tps-checkin?t=tps_qr_3_9Jsd8aKxL2m4nP6qR8tU0vW2yZ4b
```

### 4. PIC Information
Setiap TPS bisa punya info panitia:
- `pic_name`: Nama penanggung jawab
- `pic_phone`: Nomor kontak
- `notes`: Catatan internal (opsional)

### 5. Operating Hours
- `open_time`: Jam buka TPS (TIME)
- `close_time`: Jam tutup TPS (TIME)
- Format: "HH:MM" (e.g., "08:00", "15:00")

---

## 🎯 Use Cases

### Scenario 1: Setup TPS Baru dengan QR
```bash
# 1. Create TPS
POST /api/v1/admin/tps
{
  "code": "TPS01",
  "name": "TPS Aula Utama",
  "location": "Gedung A Lantai 1",
  "capacity": 500,
  "open_time": "08:00",
  "close_time": "15:00",
  "pic_name": "John Doe",
  "pic_phone": "+62812345678"
}

# 2. Generate QR pertama kali
POST /api/v1/admin/tps/1/qr/rotate

# 3. Ambil QR untuk print
GET /api/v1/admin/tps/1/qr/print
→ Dapatkan qr_payload untuk render QR code

# 4. Print QR dan tempel di TPS
```

### Scenario 2: QR Bocor - Perlu Rotate
```bash
# 1. Rotate QR (deactivate lama, generate baru)
POST /api/v1/admin/tps/1/qr/rotate

# 2. Print QR baru
GET /api/v1/admin/tps/1/qr/print

# 3. Tempel QR baru, cabut QR lama
```

### Scenario 3: Monitoring TPS dengan Info PIC
```bash
# List semua TPS dengan info lengkap
GET /api/v1/admin/tps

Response:
[
  {
    "id": 1,
    "code": "TPS01",
    "name": "TPS Aula Utama",
    "pic_name": "John Doe",
    "pic_phone": "+62812345678",
    "open_time": "08:00",
    "close_time": "15:00",
    "has_active_qr": true,
    ...
  }
]

# Admin bisa langsung lihat:
# - Siapa PIC TPS
# - Kontak PIC jika ada masalah
# - Jam operasional
# - Status QR (sudah generate atau belum)
```

---

## 🔐 Security Notes

1. **QR Token Generation**
   - Menggunakan `crypto/rand` untuk random generation
   - Length: 32 karakter (base64 encoded dari 32 bytes)
   - Tidak predictable

2. **QR Rotation**
   - QR lama di-set `is_active = false`
   - Timestamp `rotated_at` dicatat
   - QR lama tetap di database untuk audit trail

3. **Database Constraint**
   - Unique constraint: hanya 1 active QR per TPS
   - Mencegah multiple active QR secara accident

---

## 📊 Database Changes

### Before
```sql
-- tps table
id, election_id, code, name, location, status, 
voting_date, open_time, close_time, capacity_estimate, 
area_faculty_id, created_at, updated_at

-- tps_qr table  
id, tps_id, qr_secret, is_active, created_at, revoked_at
```

### After
```sql
-- tps table (added 3 columns)
id, election_id, code, name, location, status,
voting_date, open_time, close_time, capacity_estimate,
area_faculty_id, 
pic_name, pic_phone, notes,  -- NEW
created_at, updated_at

-- tps_qr table (renamed 2 columns)
id, tps_id, qr_token, is_active, created_at, rotated_at
-- + unique constraint on (tps_id) WHERE is_active = TRUE
```

---

## 🚀 Testing

### Manual Test dengan cURL

```bash
# 1. Create TPS dengan PIC info
curl -X POST http://localhost:8080/api/v1/admin/tps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TPS01",
    "name": "TPS Aula Utama",
    "location": "Gedung A",
    "capacity": 500,
    "open_time": "08:00",
    "close_time": "15:00",
    "pic_name": "John",
    "pic_phone": "+628123"
  }'

# 2. Generate QR
curl -X POST http://localhost:8080/api/v1/admin/tps/1/qr/rotate \
  -H "Authorization: Bearer <token>"

# 3. Get QR metadata
curl http://localhost:8080/api/v1/admin/tps/1/qr \
  -H "Authorization: Bearer <token>"

# 4. Get QR for print
curl http://localhost:8080/api/v1/admin/tps/1/qr/print \
  -H "Authorization: Bearer <token>"
```

---

## 📦 Git Commits

1. **feat: add TPS PIC fields and update QR schema**
   - Migration 006
   - Schema updates
   - Goose annotations

2. **feat: implement TPS PIC fields and QR management**
   - Entity & DTOs updates
   - Repository layer
   - Service layer
   - HTTP handlers
   - Routes registration

3. **docs: update TPS API documentation with QR and PIC fields**
   - Complete API documentation
   - Request/response examples
   - Use cases

---

## ✨ Next Steps (Optional Enhancements)

- [ ] QR Code SVG generation di backend (pakai library Go)
- [ ] Bulk QR generation untuk multiple TPS
- [ ] QR rotation history/audit log
- [ ] Notifikasi otomatis ke PIC via WhatsApp/SMS
- [ ] Dashboard monitoring dengan info PIC
- [ ] Auto-generate QR saat create TPS
- [ ] QR expiration policy

---

## 📝 Notes

- Migration 006 sudah applied ke database
- Semua changes sudah di-commit (3 commits)
- Documentation lengkap dan up-to-date
- Ready untuk integration testing dengan frontend
