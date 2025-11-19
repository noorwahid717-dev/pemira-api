# TPS Check-in System

Implementasi sistem check-in TPS dengan pola:
- `pgxpool.Pool` untuk connection pooling
- `WithTx(ctx, func(tx pgx.Tx) error)` untuk transaksi
- Repository per modul
- Row-level integrity (FOR UPDATE hanya di voting)

## Flow Check-in

### 1. CheckinScan - Mahasiswa scan QR di TPS

```go
service := tps.NewCheckinService(db)

result, err := service.CheckinScan(ctx, voterID, qrPayload)
if err != nil {
    // Handle error: ErrQRInvalid, ErrTPSInactive, ErrNotEligible, ErrAlreadyVoted
    return err
}

// result contains:
// - CheckinID
// - TPS info (ID, Code, Name)
// - Status: PENDING
// - Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS."
```

**Alur internal:**
1. Parse QR payload: `PEMIRA|TPS01|c9423e5f97d4`
2. Validasi QR active & TPS active
3. Cek election dalam fase VOTING_OPEN
4. Cek voter eligible & belum voting
5. Buat row `tps_checkins` dengan status PENDING
6. Log audit

### 2. ApproveCheckin - Panitia TPS menyetujui check-in

```go
service := tps.NewCheckinService(db)

result, err := service.ApproveCheckin(ctx, operatorUserID, tpsID, checkinID)
if err != nil {
    // Handle error: ErrTPSAccessDenied, ErrCheckinNotPending, ErrAlreadyVoted
    return err
}

// result contains:
// - CheckinID
// - Status: APPROVED
// - Voter info (ID, NIM, Name)
// - TPS info (ID, Code, Name)
// - ApprovedAt
// - ExpiresAt (15 menit dari sekarang)
```

**Alur internal:**
1. Validasi operator punya akses ke TPS (dari tabel `tps_panitia`)
2. Ambil check-in, pastikan status PENDING
3. Cek election masih VOTING_OPEN
4. Cek voter belum voting
5. Update check-in ke APPROVED, set approved_by_id, approved_at, expires_at
6. Log audit

### 3. RejectCheckin - Panitia TPS menolak check-in

```go
service := tps.NewCheckinService(db)

result, err := service.RejectCheckin(ctx, operatorUserID, tpsID, checkinID, "Identitas tidak sesuai")
if err != nil {
    // Handle error
    return err
}

// result contains:
// - CheckinID
// - Status: REJECTED
// - Reason
```

## Database Schema

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

### tps_panitia

```sql
CREATE TABLE tps_panitia (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL, -- KETUA_TPS, OPERATOR_PANEL
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tps_id, user_id)
);
```

### tps_qr

```sql
CREATE TABLE tps_qr (
    id BIGSERIAL PRIMARY KEY,
    tps_id BIGINT NOT NULL REFERENCES tps(id),
    qr_secret_suffix VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tps_id, qr_secret_suffix)
);
```

## Integrasi dengan Voting

### CastTPSVote - Voting setelah check-in approved

```go
votingService := voting.NewTPSVotingService(db)

receipt, err := votingService.CastTPSVote(ctx, voterID, electionID, candidateID)
if err != nil {
    // Handle error: ErrNoApprovedCheckin, ErrCheckinExpired, ErrAlreadyVoted
    return err
}

// receipt contains:
// - ReceiptToken
// - VotedAt
// - Method: "TPS"
// - TPSID
```

**Alur internal:**
1. Ambil latest `tps_checkins` dengan status APPROVED
2. Pastikan `expires_at > now()` (masih dalam window 15 menit)
3. Lock `voter_status` dengan FOR UPDATE
4. Validasi belum voting & candidate valid
5. Generate receipt token
6. Insert ke `votes` dengan tps_id
7. Update `voter_status`: has_voted=true, tps_id
8. Update `tps_checkins.status = USED`
9. Log audit

### GetTPSVotingEligibility - Cek eligibility

```go
votingService := voting.NewTPSVotingService(db)

eligibility, err := votingService.GetTPSVotingEligibility(ctx, voterID, electionID)
if err != nil {
    return err
}

if eligibility.Eligible {
    // Voter dapat voting
    // eligibility.ExpiresAt menunjukkan batas waktu
} else {
    // Tampilkan eligibility.Reason
}
```

## Error Handling

### TPS Check-in Errors

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `ErrQRInvalid` | 400 | Payload QR tidak valid |
| `ErrQRRevoked` | 400 | QR sudah direvoke |
| `ErrTPSInactive` | 400 | TPS belum/tidak aktif |
| `ErrElectionNotOpen` | 400 | Pemilu bukan di fase voting |
| `ErrNotEligible` | 400 | Mahasiswa bukan DPT |
| `ErrAlreadyVoted` | 409 | Sudah voting |
| `ErrCheckinNotFound` | 404 | Data check-in tidak ada |
| `ErrCheckinNotPending` | 400 | Check-in bukan status PENDING |
| `ErrTPSAccessDenied` | 403 | Panitia tidak assigned ke TPS |

### Voting Errors

| Error | Description |
|-------|-------------|
| `ErrNoApprovedCheckin` | Belum ada check-in approved |
| `ErrCheckinExpired` | Waktu check-in sudah habis (> 15 menit) |
| `ErrAlreadyVoted` | Sudah melakukan voting |
| `ErrInvalidCandidate` | Kandidat tidak valid |

## Transaksi & Locking

### Check-in (no lock)

Check-in relatif aman tanpa FOR UPDATE karena:
- Hanya insert data baru
- Validasi `has_voted` untuk prevent, tapi locking berat ada di voting

### Voting (with lock)

```go
// Lock voter_status untuk prevent double voting
SELECT has_voted, voted_at
FROM voter_status
WHERE election_id = $1 AND voter_id = $2
FOR UPDATE
```

Row-level locking hanya di voting untuk performa optimal.

## WebSocket Integration (Optional)

### Broadcast ke mahasiswa setelah approved

```go
// Di ApproveCheckin, setelah commit
wsHub.PublishToVoter(voterID, WSMessage{
    Type: "CHECKIN_APPROVED",
    Data: map[string]interface{}{
        "checkin_id": checkin.ID,
        "tps_id":     tpsID,
        "expires_at": expiresAt,
    },
})
```

### Broadcast ke panel TPS

```go
// Di CheckinScan, setelah commit
wsHub.PublishToTPS(tpsID, WSMessage{
    Type: "NEW_CHECKIN",
    Data: map[string]interface{}{
        "checkin_id": checkin.ID,
        "voter_id":   voterID,
        "status":     "PENDING",
    },
})
```

## Testing

### Unit Test Example

```go
func TestCheckinScan_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    service := tps.NewCheckinService(db)
    
    // Setup test data
    electionID := createTestElection(t, db)
    tpsID := createTestTPS(t, db, electionID)
    voterID := createTestVoter(t, db, electionID)
    qr := createTestQR(t, db, tpsID)
    
    payload := fmt.Sprintf("PEMIRA|%s|%s", tps.Code, qr.QRSecretSuffix)
    
    // Execute
    result, err := service.CheckinScan(context.Background(), voterID, payload)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "PENDING", result.Status)
    assert.Equal(t, tpsID, result.TPS.ID)
}
```

## Production Considerations

### TTL & Cleanup

```sql
-- Expire check-in yang sudah lewat 15 menit
UPDATE tps_checkins
SET status = 'EXPIRED'
WHERE status = 'APPROVED'
  AND expires_at < NOW();
```

Run as cron job atau background worker.

### Monitoring

Key metrics:
- Check-in creation rate
- Approval rate (approved vs rejected)
- Average approval time
- Check-in expiration rate
- Voting completion rate (approved → voted)

### Audit Trail

Semua operasi check-in dan voting sudah di-log ke `audit_logs`:
- `TPS_CHECKIN_CREATED`
- `TPS_CHECKIN_APPROVED`
- `TPS_CHECKIN_REJECTED`
- `VOTE_CAST_TPS`

## API Endpoints

### Student Side

```
POST /api/v1/voter/tps/scan
Body: { "qr_payload": "PEMIRA|TPS01|c9423e..." }
Response: ScanQRResponse

GET /api/v1/voter/tps/status?election_id=1
Response: CheckinStatusResponse
```

### TPS Panel

```
GET /api/v1/tps/:tpsID/checkins?status=PENDING
Response: CheckinQueueResponse

POST /api/v1/tps/:tpsID/checkins/:checkinID/approve
Response: ApproveCheckinResponse

POST /api/v1/tps/:tpsID/checkins/:checkinID/reject
Body: { "reason": "..." }
Response: RejectCheckinResponse
```

### Voting

```
POST /api/v1/voter/vote/tps
Body: { "candidate_id": 123 }
Response: VoteReceipt

GET /api/v1/voter/vote/tps/eligibility?election_id=1
Response: TPSVotingEligibility
```
