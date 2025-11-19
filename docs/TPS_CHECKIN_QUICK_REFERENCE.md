# TPS Check-in Quick Reference

Quick guide untuk menggunakan TPS check-in system.

## 🎯 Files Created

```
internal/tps/
├── service_checkin.go              ← Main check-in service
├── http_handler_checkin.go         ← HTTP handlers
├── service_checkin_test.go         ← Tests
├── setup_checkin_example.go        ← Setup examples
└── README_CHECKIN.md               ← Detailed docs

internal/voting/
└── service_tps.go                  ← TPS voting service

docs/
├── TPS_CHECKIN_IMPLEMENTATION.md   ← Full documentation
└── TPS_CHECKIN_QUICK_REFERENCE.md  ← This file
```

## 🚀 Quick Start

### 1. Setup Service

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    "pemira-api/internal/tps"
)

// Create connection pool
db, _ := pgxpool.New(ctx, "postgres://...")

// Create check-in service
checkinService := tps.NewCheckinService(db)

// Create voting service
votingService := voting.NewTPSVotingService(db)
```

### 2. Student Scans QR

```go
result, err := checkinService.CheckinScan(ctx, voterID, "PEMIRA|TPS01|abc123")
if err != nil {
    // Handle: ErrQRInvalid, ErrTPSInactive, ErrNotEligible, ErrAlreadyVoted
}

// result.CheckinID - ID untuk tracking
// result.Status - "PENDING"
```

### 3. Operator Approves

```go
result, err := checkinService.ApproveCheckin(ctx, operatorUserID, tpsID, checkinID)
if err != nil {
    // Handle: ErrTPSAccessDenied, ErrCheckinNotPending, ErrAlreadyVoted
}

// result.Status - "APPROVED"
// result.ApprovedAt - timestamp
// Check-in expires dalam 15 menit
```

### 4. Student Votes

```go
receipt, err := votingService.CastTPSVote(ctx, voterID, electionID, candidateID)
if err != nil {
    // Handle: ErrNoApprovedCheckin, ErrCheckinExpired, ErrAlreadyVoted
}

// receipt.Method - "TPS"
// receipt.TPS - TPS info
// Check-in status otomatis jadi "USED"
```

## 🔑 Key Functions

### Check-in Service

| Function | Purpose | Key Validations |
|----------|---------|----------------|
| `CheckinScan` | Mahasiswa scan QR | QR valid, TPS active, voter eligible, belum voting |
| `ApproveCheckin` | Operator approve | Operator has access, status pending, belum voting |
| `RejectCheckin` | Operator reject | Operator has access, status pending |

### Voting Service

| Function | Purpose | Key Validations |
|----------|---------|----------------|
| `CastTPSVote` | Vote di TPS | Check-in approved, not expired, belum voting, candidate valid |
| `GetTPSVotingEligibility` | Cek eligibility | Check-in approved & not expired |

## 🔒 Transaction Pattern

```go
// All operations use this pattern:
err := s.withTx(ctx, func(tx pgx.Tx) error {
    // 1. Load data
    data, err := s.getData(ctx, tx, id)
    if err != nil {
        return err
    }
    
    // 2. Validate
    if !data.Valid {
        return ErrInvalid
    }
    
    // 3. Update/Insert
    err = s.updateData(ctx, tx, data)
    if err != nil {
        return err
    }
    
    // 4. Audit (optional)
    _ = s.logAudit(ctx, tx, ...)
    
    return nil
})
```

## 🗄️ Key Database Tables

### tps_checkins

```sql
status:
- PENDING   → Waiting for approval
- APPROVED  → Approved, can vote (15 min TTL)
- REJECTED  → Rejected by operator
- USED      → Vote already cast
- EXPIRED   → TTL expired
```

### votes

```sql
voted_via:
- 'ONLINE' → Online voting
- 'TPS'    → TPS voting
```

### voter_status

```sql
has_voted: true/false
tps_id: Set if voted via TPS
```

## ⚠️ Error Codes

### Check-in Errors

| Error | HTTP | When |
|-------|------|------|
| `ErrQRInvalid` | 400 | QR format salah atau tidak ditemukan |
| `ErrQRRevoked` | 400 | QR sudah di-revoke |
| `ErrTPSInactive` | 400 | TPS status bukan ACTIVE |
| `ErrElectionNotOpen` | 400 | Election bukan di fase VOTING_OPEN |
| `ErrNotEligible` | 400 | Voter bukan DPT |
| `ErrAlreadyVoted` | 409 | Sudah voting |
| `ErrCheckinNotFound` | 404 | Check-in ID tidak ada |
| `ErrCheckinNotPending` | 400 | Status bukan PENDING |
| `ErrTPSAccessDenied` | 403 | Operator tidak assigned ke TPS |

### Voting Errors

| Error | When |
|-------|------|
| `ErrNoApprovedCheckin` | Belum ada check-in approved |
| `ErrCheckinExpired` | Check-in sudah lebih dari 15 menit |
| `ErrAlreadyVoted` | Sudah voting |
| `ErrInvalidCandidate` | Candidate tidak valid |

## 🧪 Testing

```bash
# Run all tests
go test ./internal/tps/... -v

# Run with race detector
go test ./internal/tps/... -race -v

# Run specific test
go test ./internal/tps/... -run TestCheckinScan_Success -v
```

## 📊 Flow Diagram

```
Student                TPS Panel              System
   |                       |                     |
   | 1. Scan QR            |                     |
   |-------------------------------------->      |
   |                       |          CheckinScan|
   |                       |          (status=PENDING)
   |<--------------------------------------      |
   |                       |                     |
   |                       | 2. View queue       |
   |                       |<------------------- |
   |                       |                     |
   |                       | 3. Approve          |
   |                       |-------------------> |
   |                       |       ApproveCheckin|
   |                       |   (status=APPROVED, |
   |                       |    expires_at=+15m) |
   |                       |<------------------- |
   |                       |                     |
   | 4. Notification       |                     |
   |<-------------------------------------------  |
   |                       |                     |
   | 5. Vote               |                     |
   |-------------------------------------->      |
   |                       |           CastTPSVote
   |                       |     (status=USED)   |
   |<--------------------------------------      |
   |                       |                     |
```

## 🔧 Common Operations

### Check if voter has active check-in

```go
// Query database directly or add method to service
SELECT status, expires_at
FROM tps_checkins
WHERE voter_id = $1 AND election_id = $2
  AND status IN ('PENDING', 'APPROVED')
ORDER BY created_at DESC
LIMIT 1
```

### List pending check-ins for TPS

```go
SELECT c.id, c.voter_id, c.status, c.scan_at,
       v.nim, v.name, v.faculty
FROM tps_checkins c
JOIN voters v ON v.id = c.voter_id
WHERE c.tps_id = $1 AND c.status = 'PENDING'
ORDER BY c.scan_at ASC
```

### Expire old check-ins (background job)

```go
UPDATE tps_checkins
SET status = 'EXPIRED', updated_at = NOW()
WHERE status = 'APPROVED' AND expires_at < NOW()
```

## 🎯 Best Practices

1. **Always use transactions** for multi-step operations
2. **Pass `tx pgx.Tx` as parameter** to repository methods
3. **Lock only when necessary** (voting, not check-in)
4. **Handle errors gracefully** with domain-specific errors
5. **Log important operations** for audit trail
6. **Set appropriate TTL** for check-ins (15 minutes default)
7. **Validate access control** for TPS operators

## 📞 Need Help?

- Full docs: `docs/TPS_CHECKIN_IMPLEMENTATION.md`
- Examples: `internal/tps/setup_checkin_example.go`
- Tests: `internal/tps/service_checkin_test.go`
- README: `internal/tps/README_CHECKIN.md`

## 🔍 Debugging Tips

### Check-in not created

```sql
-- Check if QR is active
SELECT * FROM tps_qr WHERE tps_id = ? AND is_active = true;

-- Check if TPS is active
SELECT * FROM tps WHERE id = ? AND status = 'ACTIVE';

-- Check if voter is eligible
SELECT * FROM voters WHERE id = ? AND is_eligible = true;
```

### Can't approve check-in

```sql
-- Check operator assignment
SELECT * FROM tps_panitia WHERE tps_id = ? AND user_id = ?;

-- Check check-in status
SELECT * FROM tps_checkins WHERE id = ?;
```

### Can't vote

```sql
-- Check approved check-in
SELECT * FROM tps_checkins
WHERE voter_id = ? AND election_id = ?
  AND status = 'APPROVED' AND expires_at > NOW();

-- Check voter status
SELECT * FROM voter_status WHERE election_id = ? AND voter_id = ?;
```

## 🚦 Status Flow

```
PENDING → APPROVED → USED
   ↓         ↓
REJECTED  EXPIRED
```

- **PENDING**: Just scanned, waiting approval
- **APPROVED**: Approved, can vote (15 min window)
- **REJECTED**: Rejected by operator
- **USED**: Vote already cast
- **EXPIRED**: Approval expired (> 15 min)
