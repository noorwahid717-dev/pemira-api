# TPS Check-in Implementation

Dokumentasi lengkap implementasi sistem check-in TPS untuk PEMIRA API.

## 📁 File Structure

```
internal/
├── tps/
│   ├── service_checkin.go              # ✅ Check-in service (CheckinScan, ApproveCheckin, RejectCheckin)
│   ├── http_handler_checkin.go         # ✅ HTTP handlers untuk check-in API
│   ├── service_checkin_test.go         # ✅ Unit tests
│   ├── setup_checkin_example.go        # ✅ Setup & integration examples
│   ├── README_CHECKIN.md               # ✅ Dokumentasi check-in flow
│   ├── entity.go                       # Existing: TPS, TPSQR, TPSCheckin entities
│   ├── dto.go                          # Existing: Request/Response DTOs
│   └── errors.go                       # Existing: Error definitions
└── voting/
    └── service_tps.go                  # ✅ TPS voting service (CastTPSVote)
```

## 🎯 Features Implemented

### 1. CheckinScan - Mahasiswa scan QR di TPS

**File:** `internal/tps/service_checkin.go`

**Pattern:**
- ✅ `pgxpool.Pool` untuk connection pooling
- ✅ `WithTx(ctx, func(tx pgx.Tx) error)` untuk transaksi
- ✅ Repository pattern dengan `tx pgx.Tx` parameter
- ✅ Row-level integrity (no lock di check-in, lock ada di voting)

**Flow:**
1. Parse QR payload: `PEMIRA|TPS01|c9423e5f97d4`
2. Validasi QR active & TPS active (dalam transaksi)
3. Cek election dalam fase VOTING_OPEN
4. Cek voter eligible & belum voting
5. Buat row `tps_checkins` dengan status PENDING
6. Log audit
7. Return CheckinID & status

**Usage:**
```go
service := tps.NewCheckinService(db)
result, err := service.CheckinScan(ctx, voterID, qrPayload)
// result.Status == "PENDING"
```

### 2. ApproveCheckin - Panitia TPS menyetujui check-in

**File:** `internal/tps/service_checkin.go`

**Flow:**
1. Validasi operator punya akses ke TPS (dari `tps_panitia`)
2. Ambil check-in, pastikan status PENDING
3. Cek election masih VOTING_OPEN
4. Cek voter belum voting
5. Update check-in ke APPROVED
   - Set `approved_by_id`
   - Set `approved_at`
   - Set `expires_at` (15 menit dari sekarang)
6. Log audit
7. Return approval info

**Usage:**
```go
service := tps.NewCheckinService(db)
result, err := service.ApproveCheckin(ctx, operatorUserID, tpsID, checkinID)
// result.Status == "APPROVED"
// result.ApprovedAt
```

### 3. RejectCheckin - Panitia TPS menolak check-in

**File:** `internal/tps/service_checkin.go`

**Flow:**
1. Validasi operator punya akses ke TPS
2. Ambil check-in, pastikan status PENDING
3. Update check-in ke REJECTED
   - Set `rejection_reason`
4. Log audit

**Usage:**
```go
service := tps.NewCheckinService(db)
result, err := service.RejectCheckin(ctx, operatorUserID, tpsID, checkinID, "Identitas tidak sesuai")
// result.Status == "REJECTED"
```

### 4. CastTPSVote - Voting setelah check-in approved

**File:** `internal/voting/service_tps.go`

**Pattern:**
- ✅ `pgxpool.Pool` + `WithTx`
- ✅ **Row-level lock** dengan `FOR UPDATE` di `voter_status`
- ✅ Integrasi dengan check-in system

**Flow:**
1. Ambil latest `tps_checkins` dengan status APPROVED
2. Pastikan `expires_at > now()` (masih dalam window 15 menit)
3. **Lock `voter_status` dengan FOR UPDATE** ← locking berat ada di sini
4. Validasi belum voting & candidate valid
5. Generate receipt token hash
6. Insert ke `votes` dengan `voted_via = 'TPS'`
7. Update `voter_status`: `has_voted=true`, `tps_id`
8. Update `tps_checkins.status = USED`
9. Log audit
10. Return receipt

**Usage:**
```go
votingService := voting.NewTPSVotingService(db)
receipt, err := votingService.CastTPSVote(ctx, voterID, electionID, candidateID)
// receipt.Method == "TPS"
// receipt.TPS contains TPS info
```

### 5. GetTPSVotingEligibility - Cek eligibility untuk voting

**File:** `internal/voting/service_tps.go`

**Flow:**
1. Cek voter status
2. Ambil latest approved check-in
3. Cek expires
4. Return eligibility info

**Usage:**
```go
votingService := voting.NewTPSVotingService(db)
eligibility, err := votingService.GetTPSVotingEligibility(ctx, voterID, electionID)
if eligibility.Eligible {
    // Voter dapat voting
    // eligibility.ExpiresAt shows deadline
}
```

## 🗄️ Database Schema

### tps_checkins

```sql
CREATE TABLE tps_checkins (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id),
    voter_id BIGINT NOT NULL REFERENCES voters(id),
    election_id BIGINT NOT NULL REFERENCES elections(id),
    status VARCHAR(20) NOT NULL, -- PENDING, APPROVED, REJECTED, USED, EXPIRED
    scan_at TIMESTAMP NOT NULL,
    approved_at TIMESTAMP,
    approved_by_id BIGINT REFERENCES users(id),
    rejection_reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tps_checkins_voter ON tps_checkins(voter_id, election_id);
CREATE INDEX idx_tps_checkins_tps_status ON tps_checkins(tps_id, status);
```

### votes (updated)

```sql
CREATE TABLE votes (
    id BIGSERIAL PRIMARY KEY,
    election_id BIGINT NOT NULL REFERENCES elections(id),
    candidate_id BIGINT NOT NULL REFERENCES candidates(id),
    token_hash VARCHAR(255) NOT NULL,
    voted_via VARCHAR(20) NOT NULL, -- 'ONLINE' or 'TPS'
    voted_at TIMESTAMP NOT NULL
);
```

### voter_status (updated)

```sql
CREATE TABLE voter_status (
    election_id BIGINT NOT NULL,
    voter_id BIGINT NOT NULL,
    has_voted BOOLEAN NOT NULL DEFAULT false,
    voted_at TIMESTAMP,
    tps_id BIGINT REFERENCES tps(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (election_id, voter_id)
);
```

## 🔒 Transaction & Locking Strategy

### Check-in Operations (No Lock)

```go
// CheckinScan & ApproveCheckin: No FOR UPDATE needed
// - Insert/update tps_checkins only
// - Read voter_status tanpa lock
// - Validasi has_voted cukup untuk reject early
```

**Why no lock?**
- Check-in hanya membuat "permission" untuk voting
- Race condition di sini tidak critical
- Locking berat hanya di voting

### Voting Operations (With Lock)

```go
// CastTPSVote: Lock voter_status
SELECT has_voted, voted_at
FROM voter_status
WHERE election_id = $1 AND voter_id = $2
FOR UPDATE  // ← Lock here!
```

**Why lock?**
- Prevent double voting (critical operation)
- Ensure atomicity across votes, voter_status, tps_checkins
- Row-level lock: only block same voter, not whole table

## 🌐 API Endpoints

### Student Side

```
POST /api/v1/voter/tps/scan
Body: { "qr_payload": "PEMIRA|TPS01|c9423e..." }
Response: {
  "checkin_id": 123,
  "tps": { "id": 1, "code": "TPS01", "name": "Gedung A" },
  "status": "PENDING",
  "message": "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
  "scan_at": "2024-01-15T10:30:00Z"
}

GET /api/v1/voter/tps/status?election_id=1
Response: {
  "has_active_checkin": true,
  "status": "APPROVED",
  "tps": { ... },
  "expires_at": "2024-01-15T10:45:00Z"
}
```

### TPS Panel

```
GET /api/v1/tps/:tpsID/checkins?status=PENDING
Response: {
  "items": [
    {
      "id": 123,
      "voter": { "id": 456, "nim": "1234567890", "name": "John Doe" },
      "status": "PENDING",
      "scan_at": "2024-01-15T10:30:00Z",
      "has_voted": false
    }
  ]
}

POST /api/v1/tps/:tpsID/checkins/:checkinID/approve
Response: {
  "checkin_id": 123,
  "status": "APPROVED",
  "voter": { ... },
  "tps": { ... },
  "approved_at": "2024-01-15T10:31:00Z"
}

POST /api/v1/tps/:tpsID/checkins/:checkinID/reject
Body: { "reason": "Identitas tidak sesuai" }
Response: {
  "checkin_id": 123,
  "status": "REJECTED",
  "reason": "Identitas tidak sesuai"
}
```

### Voting

```
POST /api/v1/voter/vote/tps
Body: { "candidate_id": 789 }
Response: {
  "election_id": 1,
  "voter_id": 456,
  "method": "TPS",
  "voted_at": "2024-01-15T10:32:00Z",
  "receipt": {
    "token_hash": "HASH-20240115103200",
    "note": "Voting berhasil dilakukan di TPS"
  },
  "tps": { "id": 1, "code": "TPS01", "name": "Gedung A" }
}

GET /api/v1/voter/vote/tps/eligibility?election_id=1
Response: {
  "eligible": true,
  "reason": "Anda dapat melakukan voting",
  "tps_id": 1,
  "tps_code": "TPS01",
  "tps_name": "Gedung A",
  "expires_at": "2024-01-15T10:45:00Z"
}
```

## 🧪 Testing

### Unit Tests

**File:** `internal/tps/service_checkin_test.go`

```bash
# Run all TPS tests
go test ./internal/tps/... -v

# Run specific test
go test ./internal/tps/... -run TestCheckinScan_Success -v

# Run with race detector
go test ./internal/tps/... -race -v
```

**Test Coverage:**
- ✅ `TestCheckinScan_Success` - Happy path
- ✅ `TestCheckinScan_InvalidQR` - Invalid QR format
- ✅ `TestCheckinScan_AlreadyVoted` - Voter sudah voting
- ✅ `TestCheckinScan_ExistingPending` - Return existing pending checkin
- ✅ `TestApproveCheckin_Success` - Happy path
- ✅ `TestApproveCheckin_AccessDenied` - Unauthorized operator
- ✅ `TestApproveCheckin_NotPending` - Check-in bukan pending
- ✅ `TestRejectCheckin_Success` - Happy path
- ✅ `TestFullCheckinFlow_Integration` - Full flow test
- ✅ `TestConcurrentCheckins_RaceCondition` - Race condition test

### Integration Test

```go
// See: internal/tps/setup_checkin_example.go
func ExampleIntegrationTest() {
    // 1. Setup test database
    // 2. Create service
    // 3. Test CheckinScan
    // 4. Test ApproveCheckin
    // 5. Verify results
}
```

## 🚀 Setup & Integration

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "pemira-api/internal/tps"
)

func main() {
    ctx := context.Background()
    
    // Create connection pool
    db, err := pgxpool.New(ctx, "postgres://user:pass@localhost/pemira")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create service
    checkinService := tps.NewCheckinService(db)
    
    // Create handler
    handler := tps.NewCheckinHandler(checkinService)
    
    // Register routes
    handler.RegisterRoutes(router)
}
```

### Full Application Setup

See: `internal/tps/setup_checkin_example.go`

```go
// Setup with middlewares, CORS, etc.
checkinHandler, err := tps.SetupCheckinSystem(ctx, dbURL)
checkinHandler.RegisterRoutes(r)
```

## 🔄 Background Jobs

### Expire Check-in Job

```go
// Run every minute to expire check-ins
func ExpireCheckinJob(db *pgxpool.Pool) {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        db.Exec(ctx, `
            UPDATE tps_checkins
            SET status = 'EXPIRED', updated_at = NOW()
            WHERE status = 'APPROVED' AND expires_at < NOW()
        `)
    }
}
```

## 📊 Monitoring & Metrics

Key metrics to monitor:
- Check-in creation rate
- Approval rate (approved vs rejected)
- Average approval time
- Check-in expiration rate
- Voting completion rate (approved → voted)

## 🔐 Security Considerations

1. **QR Secret**: Harus unique per TPS dan dapat di-revoke
2. **Access Control**: Operator hanya bisa approve di TPS yang di-assign
3. **TTL**: Check-in approved expires dalam 15 menit
4. **Double Voting**: Prevented by FOR UPDATE lock di voting
5. **Audit**: Semua operasi di-log untuk traceability

## 🎓 Key Patterns & Best Practices

### 1. Transaction Pattern

```go
func (s *Service) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
    tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback(ctx)
            panic(p)
        }
    }()
    if err := fn(tx); err != nil {
        _ = tx.Rollback(ctx)
        return err
    }
    return tx.Commit(ctx)
}
```

### 2. Repository with Transaction

```go
// Pass tx as parameter, not store in struct
func (s *Service) getTPSByID(ctx context.Context, tx pgx.Tx, id int64) (*TPS, error) {
    query := `SELECT ... FROM tps WHERE id = $1`
    var tps TPS
    err := tx.QueryRow(ctx, query, id).Scan(...)
    return &tps, err
}
```

### 3. Error Handling

```go
// Domain errors
var (
    ErrTPSInactive = errors.New("TPS belum/tidak aktif")
    ErrAlreadyVoted = errors.New("Mahasiswa sudah pernah voting")
)

// HTTP error mapping
func GetErrorCode(err error) (string, int) {
    if ec, ok := errorCodeMap[err]; ok {
        return ec.Code, ec.HTTPStatus
    }
    return "INTERNAL_ERROR", http.StatusInternalServerError
}
```

## 📝 Next Steps

1. **WebSocket Integration**: Real-time updates ke panel TPS dan mahasiswa
2. **Caching**: Cache active check-ins di Redis untuk fast lookup
3. **Analytics**: Dashboard untuk monitoring TPS activity
4. **Mobile App**: QR scanner di mobile app
5. **Offline Mode**: Support offline voting di TPS dengan sync later

## 🤝 Contributing

When extending this system:
1. Follow the existing transaction pattern
2. Add appropriate error handling
3. Write unit tests for new features
4. Update this documentation
5. Consider security implications

## 📚 References

- PostgreSQL pgx driver: https://github.com/jackc/pgx
- Transaction patterns: See `service_checkin.go`
- Testing patterns: See `service_checkin_test.go`
- Setup examples: See `setup_checkin_example.go`
