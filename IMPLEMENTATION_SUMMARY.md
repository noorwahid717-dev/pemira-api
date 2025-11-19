# TPS Check-in Implementation Summary

## ✅ Implementation Complete

Implementasi lengkap sistem check-in TPS untuk PEMIRA API dengan pola yang diminta:
- ✅ `pgxpool.Pool` untuk connection pooling
- ✅ `WithTx(ctx, func(tx pgx.Tx) error)` untuk transaksi
- ✅ Repository per modul dengan parameter `tx pgx.Tx`
- ✅ Row-level integrity dengan `FOR UPDATE` hanya di voting

## 📁 Files Created

### Core Implementation (1,700 lines)

1. **internal/tps/service_checkin.go** (605 lines)
   - `CheckinService` struct dengan `pgxpool.Pool`
   - `CheckinScan()` - Mahasiswa scan QR di TPS
   - `ApproveCheckin()` - Panitia TPS approve check-in
   - `RejectCheckin()` - Panitia TPS reject check-in
   - Helper methods dengan parameter `tx pgx.Tx`
   - `withTx()` helper untuk transaction management

2. **internal/tps/http_handler_checkin.go** (235 lines)
   - `CheckinHandler` untuk REST API
   - Route registration
   - Request/response handling
   - Error handling & HTTP status mapping

3. **internal/voting/service_tps.go** (393 lines)
   - `TPSVotingService` untuk voting di TPS
   - `CastTPSVote()` - Voting dengan check-in approved
   - `GetTPSVotingEligibility()` - Cek eligibility
   - Row-level lock dengan `FOR UPDATE` di `voter_status`
   - Integration dengan check-in system

4. **internal/tps/service_checkin_test.go** (467 lines)
   - 15+ unit tests
   - Integration tests
   - Race condition tests
   - Test helpers & fixtures

### Documentation (29,818 characters)

5. **internal/tps/README_CHECKIN.md** (8,577 chars)
   - Detailed flow documentation
   - Database schema
   - API endpoints
   - Error handling
   - WebSocket integration guide

6. **internal/tps/setup_checkin_example.go** (6,285 chars)
   - Setup examples
   - Integration examples
   - Background job examples
   - Main.go integration guide

7. **docs/TPS_CHECKIN_IMPLEMENTATION.md** (12,996 chars)
   - Complete technical documentation
   - Architecture overview
   - Pattern explanations
   - Security considerations
   - Production considerations

8. **docs/TPS_CHECKIN_QUICK_REFERENCE.md** (8,245 chars)
   - Quick start guide
   - Common operations
   - Debugging tips
   - Status flow diagram

## 🎯 Key Features

### 1. CheckinScan - Mahasiswa Scan QR
```go
service := tps.NewCheckinService(db)
result, err := service.CheckinScan(ctx, voterID, "PEMIRA|TPS01|abc123")
// Creates check-in with status PENDING
```

**Flow:**
1. Parse QR payload
2. Validate QR & TPS (in transaction)
3. Check election phase
4. Check voter eligibility
5. Create `tps_checkins` row with status PENDING
6. Log audit

### 2. ApproveCheckin - Panitia Approve
```go
result, err := service.ApproveCheckin(ctx, operatorUserID, tpsID, checkinID)
// Updates check-in to APPROVED with 15-min expiry
```

**Flow:**
1. Validate operator access to TPS
2. Load check-in, ensure status PENDING
3. Validate election still open
4. Validate voter hasn't voted
5. Update to APPROVED with `expires_at`
6. Log audit

### 3. RejectCheckin - Panitia Reject
```go
result, err := service.RejectCheckin(ctx, operatorUserID, tpsID, checkinID, "Reason")
// Updates check-in to REJECTED
```

### 4. CastTPSVote - Voting di TPS
```go
votingService := voting.NewTPSVotingService(db)
receipt, err := votingService.CastTPSVote(ctx, voterID, electionID, candidateID)
// Casts vote using approved check-in
```

**Flow:**
1. Get latest APPROVED check-in
2. Validate not expired (< 15 min)
3. **Lock `voter_status` with FOR UPDATE** ← Critical section
4. Validate hasn't voted
5. Insert vote with `voted_via='TPS'`
6. Update `voter_status` and `tps_checkins`
7. Log audit

## 🔒 Transaction & Locking Strategy

### Check-in: No Lock (Optimized)
```go
// No FOR UPDATE needed for check-in operations
// - Only insert/update tps_checkins
// - Read voter_status without lock
// - Race condition not critical here
```

### Voting: With Lock (Critical)
```go
// FOR UPDATE only at voting
SELECT has_voted FROM voter_status
WHERE election_id = $1 AND voter_id = $2
FOR UPDATE  // Prevents double voting
```

**Why this pattern?**
- Check-in creates "permission", not critical if race occurs
- Voting is critical operation, needs strong consistency
- Row-level lock: only blocks same voter
- Better performance: lock only when necessary

## 🗄️ Database Schema

### tps_checkins (New)
```sql
CREATE TABLE tps_checkins (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL,
    voter_id BIGINT NOT NULL,
    election_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,  -- PENDING, APPROVED, REJECTED, USED, EXPIRED
    scan_at TIMESTAMP NOT NULL,
    approved_at TIMESTAMP,
    approved_by_id BIGINT,
    rejection_reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

### votes (Updated)
```sql
-- Add voted_via column
ALTER TABLE votes ADD COLUMN voted_via VARCHAR(20);  -- 'ONLINE' or 'TPS'
```

### voter_status (Updated)
```sql
-- Add tps_id column
ALTER TABLE voter_status ADD COLUMN tps_id BIGINT REFERENCES tps(id);
```

## 🌐 API Endpoints

### Student
- `POST /api/v1/voter/tps/scan` - Scan QR code
- `GET /api/v1/voter/tps/status` - Check check-in status

### TPS Panel
- `GET /api/v1/tps/:tpsID/checkins` - List check-ins
- `POST /api/v1/tps/:tpsID/checkins/:id/approve` - Approve check-in
- `POST /api/v1/tps/:tpsID/checkins/:id/reject` - Reject check-in

### Voting
- `POST /api/v1/voter/vote/tps` - Cast vote at TPS
- `GET /api/v1/voter/vote/tps/eligibility` - Check eligibility

## 🧪 Testing

```bash
# Run all tests
go test ./internal/tps/... -v

# With race detector
go test ./internal/tps/... -race -v

# Specific test
go test ./internal/tps/... -run TestCheckinScan_Success -v
```

**Test Coverage:**
- ✅ Happy paths
- ✅ Error cases
- ✅ Access control
- ✅ Race conditions
- ✅ Full integration flow

## 📊 Code Statistics

| File | Lines | Purpose |
|------|-------|---------|
| `service_checkin.go` | 605 | Core check-in logic |
| `http_handler_checkin.go` | 235 | REST API handlers |
| `service_tps.go` | 393 | TPS voting logic |
| `service_checkin_test.go` | 467 | Unit & integration tests |
| **Total** | **1,700** | Production code |

## ✨ Pattern Highlights

### 1. Transaction Management
```go
func (s *CheckinService) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
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
// Pass tx as parameter, not stored in struct
func (s *CheckinService) getTPSByID(ctx context.Context, tx pgx.Tx, id int64) (*TPS, error) {
    query := `SELECT ... FROM tps WHERE id = $1`
    var tps TPS
    err := tx.QueryRow(ctx, query, id).Scan(...)
    return &tps, err
}
```

### 3. Error Handling
```go
var (
    ErrTPSInactive = errors.New("TPS belum/tidak aktif")
    ErrAlreadyVoted = errors.New("Mahasiswa sudah pernah voting")
)

func GetErrorCode(err error) (string, int) {
    if ec, ok := errorCodeMap[err]; ok {
        return ec.Code, ec.HTTPStatus
    }
    return "INTERNAL_ERROR", 500
}
```

## 🚀 Quick Start

```go
package main

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "pemira-api/internal/tps"
    "pemira-api/internal/voting"
)

func main() {
    // Setup connection pool
    db, _ := pgxpool.New(ctx, "postgres://...")
    
    // Create services
    checkinService := tps.NewCheckinService(db)
    votingService := voting.NewTPSVotingService(db)
    
    // Use services
    result, _ := checkinService.CheckinScan(ctx, voterID, qrPayload)
    receipt, _ := votingService.CastTPSVote(ctx, voterID, electionID, candidateID)
}
```

## 🎓 Design Decisions

1. **pgxpool.Pool over database/sql**: Better performance, native PostgreSQL features
2. **Transaction function pattern**: Clean, composable, prevents leaks
3. **Repository per module**: Clear separation, easy to test
4. **Lock only at voting**: Optimized performance, clear critical section
5. **Domain errors**: Type-safe, easy to map to HTTP codes
6. **Audit logging**: Complete traceability
7. **15-min TTL**: Balance between convenience and security

## 🔐 Security Features

- ✅ QR validation with secret suffix
- ✅ TPS operator access control
- ✅ Check-in expiration (15 min)
- ✅ Double voting prevention with row lock
- ✅ Audit trail for all operations
- ✅ Domain error handling

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| `TPS_CHECKIN_IMPLEMENTATION.md` | Full technical docs |
| `TPS_CHECKIN_QUICK_REFERENCE.md` | Quick start guide |
| `README_CHECKIN.md` | Flow & integration |
| `setup_checkin_example.go` | Code examples |

## ✅ Checklist

- [x] CheckinScan implementation
- [x] ApproveCheckin implementation
- [x] RejectCheckin implementation
- [x] CastTPSVote implementation
- [x] GetTPSVotingEligibility implementation
- [x] HTTP handlers
- [x] Unit tests (15+ tests)
- [x] Integration tests
- [x] Race condition tests
- [x] Error handling
- [x] Audit logging
- [x] Transaction management
- [x] Documentation
- [x] Quick reference
- [x] Setup examples
- [x] Code compiles without errors

## 🎉 Ready for Production

Implementasi sudah lengkap dan siap digunakan dengan:
- ✅ Production-ready code patterns
- ✅ Comprehensive testing
- ✅ Complete documentation
- ✅ Security best practices
- ✅ Performance optimization

## 📞 Next Steps

1. **Database Migration**: Run migration untuk create `tps_checkins` table
2. **Integration**: Integrate dengan main application
3. **Testing**: Test di staging environment
4. **Monitoring**: Setup metrics & alerts
5. **Documentation**: Update API docs untuk frontend team

## 👨‍💻 Author Notes

Implementasi mengikuti pola yang diminta:
- ✅ `pgxpool.Pool` + `WithTx(ctx, func(tx pgx.Tx) error)`
- ✅ Repo per modul dengan parameter `tx pgx.Tx`
- ✅ Row-level lock hanya di voting
- ✅ No lock di check-in untuk performance
- ✅ Real, functional, production-ready code

Semua kode sudah di-compile dan di-test. Siap untuk production deployment! 🚀
