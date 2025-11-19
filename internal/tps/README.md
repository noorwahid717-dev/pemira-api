# TPS Module

Modul untuk manajemen TPS (Tempat Pemungutan Suara) dengan fitur:
- Manajemen TPS di level admin (CRUD, assignment panitia)
- QR statis per TPS dengan regenerate capability
- Flow check-in mahasiswa (scan QR → queue → verifikasi panitia → siap voting)
- Panel TPS real-time untuk pantau antrian & approve/deny pemilih
- WebSocket untuk real-time queue updates

## 📁 File Structure

```
tps/
├── entity.go           # Domain entities (TPS, TPSQR, TPSPanitia, TPSCheckin)
├── dto.go              # Request/Response DTOs
├── errors.go           # Custom error definitions & codes
├── repository.go       # Repository interface
├── service.go          # Business logic layer
├── http_handler.go     # REST API handlers
├── websocket_handler.go # WebSocket handler for real-time updates
└── README.md           # This file
```

## 🎭 Actors & Permissions

### Admin Universitas / KPUM
- Kelola daftar TPS (create, read, update)
- Assign/manage panitia TPS
- Generate/regenerate QR codes
- View all TPS statistics

### Mahasiswa (Student)
- Scan QR di lokasi TPS
- Check status check-in
- Waiting for approval from panitia

### Panitia TPS (TPS Operator)
- View real-time check-in queue
- Approve/reject check-in after identity verification
- Monitor TPS statistics

## 🔄 Check-in Flow

```
1. Mahasiswa arrives at TPS physically
   ↓
2. Scan QR code via mobile app
   POST /tps/checkin/scan
   ↓
3. System validates:
   - QR valid & not revoked
   - TPS active
   - Election in VOTING_OPEN phase
   - Voter eligible (DPT)
   - Voter hasn't voted yet
   ↓
4. Create check-in record (status: PENDING)
   ↓
5. Broadcast to WebSocket → Panel TPS updates
   ↓
6. Panitia sees voter in queue
   Verifies physical identity (KTM, etc)
   ↓
7. Panitia approve/reject:
   POST /tps/:tps_id/checkins/:id/approve
   OR
   POST /tps/:tps_id/checkins/:id/reject
   ↓
8. If APPROVED:
   - Student can vote (via Voting module)
   - Check-in expires after 15 minutes
   ↓
9. After voting:
   - Mark check-in as USED
   - Prevent duplicate voting
```

## 🔐 QR Code Format

```
PEMIRA|<TPS_CODE>|<QR_SECRET_SUFFIX>
Example: PEMIRA|TPS01|c9423e5f97d4
```

- **PEMIRA**: Fixed prefix for validation
- **TPS_CODE**: Unique TPS identifier (e.g., TPS01, TPS02)
- **QR_SECRET_SUFFIX**: Cryptographically random 12-char hex string

### QR Security
- Secret rotates when regenerated
- Old QR marked as revoked (is_active = false)
- Prevents replay attacks with check-in expiry

## 📊 Status Types

### TPS Status
- `DRAFT`: TPS created but not yet active
- `ACTIVE`: TPS operational for voting
- `CLOSED`: TPS closed after voting period

### Check-in Status
- `PENDING`: Waiting for panitia verification
- `APPROVED`: Panitia approved, voter can vote
- `REJECTED`: Panitia rejected (identity mismatch, etc)
- `USED`: Voter has voted
- `EXPIRED`: Check-in approval expired (15 min timeout)

### Panitia Role
- `KETUA_TPS`: TPS leader
- `OPERATOR_PANEL`: Panel operator (can approve/reject)

## 🛰️ WebSocket Events

### Client → Server
```json
{
  "type": "PING"
}
```

### Server → Client

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

## 🔗 Integration Points

### With Voting Module
```go
// Voting module validates before casting vote
func (v *VotingService) CastTPSVote(ctx, voterID, electionID) error {
    // 1. Get active check-in
    checkin := tpsRepo.GetCheckinByVoter(voterID, electionID)
    
    // 2. Validate
    if checkin.Status != "APPROVED" {
        return ErrCheckinNotApproved
    }
    
    if time.Now().After(checkin.ExpiresAt) {
        return ErrCheckinExpired
    }
    
    // 3. Process vote...
    
    // 4. Mark check-in as USED
    checkin.Status = "USED"
    tpsRepo.UpdateCheckin(checkin)
}
```

### With User/Auth Module
```go
// Repository needs to fetch voter info
type Repository interface {
    GetVoterInfo(ctx, voterID int64) (*VoterInfo, error)
    IsVoterEligible(ctx, voterID, electionID int64) (bool, error)
    HasVoterVoted(ctx, voterID, electionID int64) (bool, error)
}
```

## 🧪 Testing Checklist

### Unit Tests
- [ ] Service: Create TPS with auto QR generation
- [ ] Service: Regenerate QR (revoke old, create new)
- [ ] Service: Scan QR validation (invalid format, revoked, inactive)
- [ ] Service: Check-in validation (eligibility, already voted)
- [ ] Service: Approve/reject access control

### Integration Tests
- [ ] HTTP: Admin CRUD TPS
- [ ] HTTP: Assign panitia
- [ ] HTTP: Student scan QR → pending
- [ ] HTTP: Panitia approve → check-in status updates
- [ ] WebSocket: Real-time queue updates

### E2E Tests
- [ ] Complete flow: Create TPS → Student scan → Panitia approve → Vote
- [ ] Error cases: Invalid QR, already voted, expired check-in
- [ ] Access control: Non-panitia can't approve

## 📝 Usage Examples

### Admin Creates TPS
```go
req := &CreateTPSRequest{
    ElectionID:       1,
    Code:             "TPS01",
    Name:             "TPS 1 – Aula Utama",
    Location:         "Aula Utama, Gedung A Lt.1",
    VotingDate:       "2025-06-13",
    OpenTime:         "08:00",
    CloseTime:        "16:00",
    CapacityEstimate: 1000,
    Status:           StatusDraft,
}

tpsID, err := tpsService.Create(ctx, req)
```

### Student Scans QR
```go
req := &ScanQRRequest{
    QRPayload: "PEMIRA|TPS01|c9423e5f97d4",
}

result, err := tpsService.ScanQR(ctx, voterID, req)
// result.Status = "PENDING"
// result.Message = "Silakan menunggu verifikasi panitia TPS."
```

### Panitia Approves Check-in
```go
result, err := tpsService.ApproveCheckin(ctx, tpsID, checkinID, panitiaUserID)
// result.Status = "APPROVED"
// result.ExpiresAt = now + 15 minutes

// Broadcast via WebSocket
wsHub.BroadcastCheckinUpdated(tpsID, checkinID, "APPROVED")
```

## 🚀 Deployment Notes

### Environment Variables
```env
TPS_CHECKIN_EXPIRY_MINUTES=15
TPS_QR_SECRET_LENGTH=12
```

### Database Indexes
Critical indexes for performance:
- `idx_tps_checkins_voter_id` - Fast lookup voter's check-in
- `idx_tps_checkins_status` - Filter pending queue
- `idx_tps_panitia_user_id` - Access control checks

### WebSocket Considerations
- Use Redis for multi-instance WebSocket coordination
- Implement heartbeat/ping-pong for connection health
- Rate limit reconnection attempts

## 📚 API Documentation

See [docs/TPS_API.md](../../docs/TPS_API.md) for complete API reference.

## 🐛 Common Issues

### Issue: QR scan fails with QR_INVALID
**Solution**: Check QR format matches `PEMIRA|<CODE>|<SECRET>`

### Issue: Panitia can't approve check-in
**Solution**: Verify panitia is assigned to the TPS via `tps_panitia` table

### Issue: WebSocket disconnects frequently
**Solution**: Implement client-side reconnection logic with exponential backoff

### Issue: Check-in expires before student can vote
**Solution**: Increase `TPS_CHECKIN_EXPIRY_MINUTES` or optimize voting flow

## 🔮 Future Enhancements

- [ ] QR code rotation schedule (auto-regenerate daily)
- [ ] SMS/push notification when check-in approved
- [ ] Offline mode for TPS panel (local cache)
- [ ] Multi-language support for error messages
- [ ] Analytics dashboard for TPS performance
- [ ] Facial recognition for automated identity verification
- [ ] Queue position indicator for students
- [ ] Estimated wait time calculation

## 📞 Support

For issues or questions:
- Check API documentation first
- Review error codes in `errors.go`
- Contact backend team: backend@pemira.example.com
