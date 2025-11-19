# Voting API - Kontrak Lengkap

## 📋 Daftar Isi
1. [Overview](#overview)
2. [Error Codes](#error-codes)
3. [Endpoints](#endpoints)
4. [Security & Validation](#security--validation)
5. [Implementation Notes](#implementation-notes)

---

## Overview

Modul Voting mengatur:
- ✅ **Voting Online** langsung dari dashboard pemilih
- ✅ **Voting via TPS** setelah check-in disetujui panitia
- ✅ **Validasi ketat** mencegah double voting
- ✅ **Transaction-based** dengan row-level locks
- ✅ **Anonymous receipts** tanpa mengungkap pilihan

### Asumsi
- Auth menggunakan JWT: `Authorization: Bearer <token>`
- JWT Claims minimal: `sub` (voter_id), `role` (STUDENT)
- Status voter dasar dari `/me/voter-status` (modul voter)
- TPS check-in sudah di-handle oleh modul TPS

---

## Error Codes

| Code | HTTP | Deskripsi | Contoh Kasus |
|------|------|-----------|--------------|
| `UNAUTHORIZED` | 401 | Token tidak valid/expired | Token hilang atau salah |
| `FORBIDDEN` | 403 | Role tidak boleh voting | User bukan STUDENT |
| `ELECTION_NOT_FOUND` | 404 | Pemilu aktif tidak ada | Belum ada pemilu dibuka |
| `ELECTION_NOT_OPEN` | 400 | Fase voting belum/sudah tutup | Masih fase kampanye |
| `NOT_ELIGIBLE` | 400 | Bukan DPT / tidak eligible | Mahasiswa bukan DPT |
| `ALREADY_VOTED` | 409 | Sudah voting sebelumnya | Double voting attempt |
| `CANDIDATE_NOT_FOUND` | 404 | Kandidat tidak ada | ID kandidat salah |
| `CANDIDATE_INACTIVE` | 400 | Kandidat tidak aktif | Kandidat di-disable |
| `TPS_REQUIRED` | 400 | Voting TPS belum check-in | Belum scan QR |
| `TPS_CHECKIN_NOT_FOUND` | 404 | Check-in tidak ditemukan | Belum ada check-in |
| `TPS_CHECKIN_NOT_APPROVED` | 400 | Check-in belum disetujui | Status masih PENDING |
| `TPS_CHECKIN_EXPIRED` | 400 | Check-in sudah kadaluarsa | Lewat 15 menit |
| `TPS_MISMATCH` | 400 | TPS tidak cocok | TPS berbeda |
| `METHOD_NOT_ALLOWED` | 400 | Mode voting tidak diizinkan | Online disabled |
| `VALIDATION_ERROR` | 422 | Request body tidak valid | candidate_id kosong |
| `INTERNAL_ERROR` | 500 | Server error | Database down |

### Format Error Response

```json
{
  "success": false,
  "error": {
    "code": "ALREADY_VOTED",
    "message": "Anda sudah menggunakan hak suara untuk pemilu ini.",
    "details": null
  }
}
```

---

## Endpoints

### 1. Get Voting Config

**Endpoint:** `GET /voting/config`

**Auth:** Required (STUDENT)

**Deskripsi:** Cek apakah user boleh voting dan mode apa yang tersedia.

**Response:**
```json
{
  "success": true,
  "data": {
    "election": {
      "id": 1,
      "code": "PEMIRA_UNIWA_2025",
      "name": "Pemira UNIWA 2025",
      "status": "VOTING_OPEN",
      "voting_start_at": "2025-06-12T08:00:00Z",
      "voting_end_at": "2025-06-15T16:00:00Z"
    },
    "voter": {
      "id": 123,
      "nim": "2110510023",
      "name": "Noah Febriyansyah",
      "is_eligible": true,
      "has_voted": false,
      "voting_method": null,
      "voted_at": null
    },
    "mode": {
      "online_enabled": true,
      "tps_enabled": true
    }
  }
}
```

---

### 2. Cast Vote Online

**Endpoint:** `POST /voting/online/cast`

**Auth:** Required (STUDENT)

**Deskripsi:** Submit vote secara online dengan validasi ketat.

**Request Body:**
```json
{
  "candidate_id": 2
}
```

**Validasi:**
1. ✅ Voter must be eligible (in DPT)
2. ✅ Voter has not voted yet
3. ✅ Election in VOTING phase
4. ✅ Candidate is active
5. ✅ Online voting enabled

**Transaction Flow:**
```sql
BEGIN TRANSACTION;
  SELECT * FROM voter_status WHERE voter_id = ? FOR UPDATE;
  -- Validate: not voted, eligible, valid phase
  INSERT INTO vote_tokens ...;
  INSERT INTO votes ...;
  UPDATE voter_status SET has_voted = true ...;
  UPDATE vote_stats ...;
  INSERT INTO audit_logs ...;
COMMIT;
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "election_id": 1,
    "voter_id": 123,
    "method": "ONLINE",
    "voted_at": "2025-06-13T10:23:45Z",
    "receipt": {
      "token_hash": "vt_9f2a8d3c3b5e",
      "note": "Simpan token ini sebagai bukti bahwa sistem telah mencatat suara Anda."
    }
  }
}
```

**Errors:**
- `400` VALIDATION_ERROR - candidate_id kosong
- `400` ELECTION_NOT_OPEN - voting belum/sudah tutup
- `400` NOT_ELIGIBLE - bukan DPT
- `404` CANDIDATE_NOT_FOUND - kandidat tidak ada
- `409` ALREADY_VOTED - sudah voting

---

### 3. Cast Vote via TPS

**Endpoint:** `POST /voting/tps/cast`

**Auth:** Required (STUDENT)

**Deskripsi:** Submit vote via TPS setelah check-in disetujui.

**Request Body:**
```json
{
  "candidate_id": 2
}
```

**Prerequisites:**
- ✅ Check-in TPS status = APPROVED
- ✅ Check-in belum expire (< 15 menit)
- ✅ TPS voting mode enabled

**Response (200):**
```json
{
  "success": true,
  "data": {
    "election_id": 1,
    "voter_id": 123,
    "method": "TPS",
    "tps": {
      "id": 5,
      "code": "TPS02",
      "name": "TPS 2 – FEB"
    },
    "voted_at": "2025-06-13T11:15:02Z",
    "receipt": {
      "token_hash": "vt_7d1fbd1029ac",
      "note": "Suara Anda melalui TPS sudah dicatat."
    }
  }
}
```

**Errors:**
- `400` METHOD_NOT_ALLOWED - TPS mode disabled
- `400` TPS_CHECKIN_NOT_FOUND - belum check-in
- `400` TPS_CHECKIN_NOT_APPROVED - belum disetujui
- `400` TPS_CHECKIN_EXPIRED - sudah kadaluarsa
- `409` ALREADY_VOTED - sudah voting

---

### 4. Check TPS Voting Status

**Endpoint:** `GET /voting/tps/status`

**Auth:** Required (STUDENT)

**Deskripsi:** Cek eligibility untuk TPS voting sebelum menampilkan form.

**Response - Not Eligible:**
```json
{
  "success": true,
  "data": {
    "eligible": false,
    "reason": "TPS_REQUIRED"
  }
}
```

**Response - Eligible:**
```json
{
  "success": true,
  "data": {
    "eligible": true,
    "tps": {
      "id": 5,
      "code": "TPS02",
      "name": "TPS 2 – FEB"
    },
    "expires_at": "2025-06-13T11:25:00Z"
  }
}
```

---

### 5. Get Voting Receipt

**Endpoint:** `GET /voting/receipt`

**Auth:** Required (STUDENT)

**Deskripsi:** Dapatkan bukti voting tanpa mengungkap kandidat.

**Response - Not Voted:**
```json
{
  "success": true,
  "data": {
    "has_voted": false
  }
}
```

**Response - Already Voted:**
```json
{
  "success": true,
  "data": {
    "has_voted": true,
    "election_id": 1,
    "method": "TPS",
    "tps": {
      "id": 5,
      "code": "TPS02",
      "name": "TPS 2 – FEB"
    },
    "voted_at": "2025-06-13T11:15:02Z",
    "receipt": {
      "token_hash": "vt_7d1fbd1029ac"
    }
  }
}
```

**⚠️ PENTING:** Endpoint ini **TIDAK BOLEH** mengungkap kandidat yang dipilih!

---

## Security & Validation

### 1. Transaction Isolation
```go
tx, err := pool.BeginTx(ctx, pgx.TxOptions{
    IsoLevel: pgx.ReadCommitted,
})
```

### 2. Row-Level Locking
```sql
SELECT * FROM voter_election_status 
WHERE voter_id = $1 AND election_id = $2 
FOR UPDATE;
```

### 3. Anonymous Voting
- Vote table **TIDAK** punya kolom `voter_id`
- Hubungan voter → vote hanya via `token_hash`
- Token di-hash sebelum disimpan

### 4. Idempotency
Jika client double-submit:
- Option 1: Return `409 ALREADY_VOTED`
- Option 2: Return `200 OK` dengan receipt yang sama (lebih kompleks)

### 5. Rate Limiting
```go
// Apply to voting endpoints
rateLimiter := middleware.NewRateLimiter(3, 5) // 3 req/min, burst 5
r.With(rateLimiter.Limit).Post("/voting/online/cast", handler.CastOnlineVote)
```

### 6. Audit Logging
Setiap cast vote log:
```go
auditSvc.Log(ctx, 
    voterID, 
    audit.ActionVoteCast, 
    "VOTE", 
    candidateID, 
    map[string]interface{}{
        "election_id": electionID,
        "method": "ONLINE",
        // TIDAK log kandidat di metadata
    },
    ipAddress,
    userAgent,
)
```

---

## Implementation Notes

### ✅ Hal yang Wajib

1. **Selalu pakai transaction + FOR UPDATE**
   ```go
   tx.QueryRow(ctx, `SELECT ... FOR UPDATE`)
   ```

2. **Validasi sequence ketat**
   - Check election phase
   - Check voter eligibility
   - Check not voted yet
   - Check candidate active

3. **Generate anonymous token**
   ```go
   token := generateRandomToken(32)
   hash := sha256(token)
   // Store hash, return token
   ```

4. **Update voter_status atomically**
   ```go
   UPDATE voter_election_status 
   SET has_voted = true, 
       voted_at = NOW(), 
       voting_method = $1
   WHERE id = $2
   ```

5. **Audit trail**
   - Log setiap voting attempt (success/fail)
   - Jangan log candidate_id di audit

### ⚠️ Hal yang Dilarang

1. ❌ **Jangan expose candidate_id di receipt/status**
2. ❌ **Jangan skip transaction**
3. ❌ **Jangan skip row lock**
4. ❌ **Jangan simpan voter_id di votes table**
5. ❌ **Jangan trust election_id dari client**

### 🔧 Optimasi

1. **Materialized view untuk live count**
   ```sql
   CREATE TABLE vote_stats (
       election_id BIGINT,
       candidate_id BIGINT,
       total_votes BIGINT,
       updated_at TIMESTAMP,
       PRIMARY KEY (election_id, candidate_id)
   );
   ```

2. **Index untuk performa**
   ```sql
   CREATE INDEX idx_voter_status_lookup 
   ON voter_election_status(voter_id, election_id, has_voted);
   ```

3. **Cache config di memory**
   ```go
   // Cache election config per 30 detik
   cachedConfig := cache.Get("election:current")
   ```

---

## Testing Checklist

### Unit Tests
- [ ] `CastOnlineVote` - success case
- [ ] `CastOnlineVote` - already voted
- [ ] `CastOnlineVote` - not eligible
- [ ] `CastTPSVote` - success case
- [ ] `CastTPSVote` - no check-in
- [ ] `CastTPSVote` - check-in not approved
- [ ] Token generation uniqueness

### Integration Tests
- [ ] Concurrent voting attempts (race condition)
- [ ] Double-submit idempotency
- [ ] Transaction rollback on error
- [ ] TPS check-in expiry

### Load Tests
- [ ] 100 concurrent votes
- [ ] 1000 votes per minute
- [ ] Database connection pool exhaustion

---

## OpenAPI Spec

Lihat file lengkap: [`docs/openapi-voting.yaml`](./openapi-voting.yaml)

Preview di Swagger UI:
```bash
# Install swagger-ui
npm install -g swagger-ui-dist

# Serve OpenAPI spec
swagger-ui-serve docs/openapi-voting.yaml
```

---

## Reference Implementation

Lihat kode lengkap:
- [`internal/voting/transaction.go`](../internal/voting/transaction.go) - Transaction-based voting
- [`internal/voting/service.go`](../internal/voting/service.go) - Business logic
- [`internal/voting/http_handler.go`](../internal/voting/http_handler.go) - HTTP handlers
- [`internal/voting/dto.go`](../internal/voting/dto.go) - Request/Response DTOs

---

**Status**: ✅ Complete API Contract  
**Last Updated**: 2025-11-19  
**Version**: 1.0.0
