# TPS Module - Quick Reference Card

## 🚀 Quick Start (5 Minutes)

### 1. Run Migrations
```bash
cd /home/noah/project/pemira-api
migrate -path migrations -database "postgres://user:pass@localhost/pemira?sslmode=disable" up
```

### 2. Setup in main.go
```go
import "pemira-api/internal/tps"

func main() {
    db := setupDatabase()
    router := chi.NewRouter()
    
    // Initialize TPS module
    tpsService, _ := tps.SetupTPSModule(db, router)
    
    http.ListenAndServe(":8080", router)
}
```

### 3. Test Endpoints
```bash
# List TPS
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/admin/tps

# Create TPS
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"election_id":1,"code":"TPS01","name":"TPS 1","location":"Aula","voting_date":"2025-06-13","open_time":"08:00","close_time":"16:00","status":"DRAFT"}' \
  http://localhost:8080/admin/tps

# Scan QR (as student)
curl -X POST -H "Authorization: Bearer $STUDENT_TOKEN" \
  -d '{"qr_payload":"PEMIRA|TPS01|abc123"}' \
  http://localhost:8080/tps/checkin/scan
```

---

## 📋 Common Use Cases

### Admin: Create & Setup TPS

```bash
# 1. Create TPS
POST /admin/tps
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

# 2. Assign Panitia
PUT /admin/tps/1/panitia
{
  "members": [
    {"user_id": 101, "role": "KETUA_TPS"},
    {"user_id": 102, "role": "OPERATOR_PANEL"}
  ]
}

# 3. Get QR Code
GET /admin/tps/1
# Response includes qr.payload for printing

# 4. Activate TPS
PUT /admin/tps/1
{"status": "ACTIVE", ...other fields...}
```

### Student: Check-in Flow

```bash
# 1. Scan QR at TPS location
POST /tps/checkin/scan
{"qr_payload": "PEMIRA|TPS01|abc123"}

# 2. Poll status (every 3-5 seconds)
GET /tps/checkin/status?election_id=1

# 3. When approved, proceed to voting
# (status will be "APPROVED" with expires_at)
```

### Panitia: Verify & Approve

```bash
# 1. Connect WebSocket (for real-time)
ws://localhost:8080/ws/tps/1/queue?token=JWT

# 2. OR poll queue
GET /tps/1/checkins?status=PENDING

# 3. Verify physical identity (KTM, etc)

# 4. Approve
POST /tps/1/checkins/555/approve

# 5. OR Reject
POST /tps/1/checkins/555/reject
{"reason": "Data tidak cocok"}
```

---

## 🔧 Code Integration Examples

### Voting Module Integration

```go
// In your voting service
func (v *VotingService) CastTPSVote(
    ctx context.Context, 
    voterID, electionID int64,
    ballot Ballot,
) error {
    // 1. Validate check-in
    checkin, err := v.tpsRepo.GetCheckinByVoter(ctx, voterID, electionID)
    if err != nil || checkin == nil {
        return errors.New("no active check-in")
    }
    
    if checkin.Status != tps.CheckinStatusApproved {
        return errors.New("check-in not approved")
    }
    
    if time.Now().After(*checkin.ExpiresAt) {
        return tps.ErrCheckinExpired
    }
    
    // 2. Process vote
    if err := v.processVote(ctx, ballot); err != nil {
        return err
    }
    
    // 3. Mark check-in as used
    return v.tpsService.MarkCheckinAsUsed(ctx, checkin.ID)
}
```

### WebSocket Broadcasting

```go
// After approving check-in
result, err := tpsService.ApproveCheckin(ctx, tpsID, checkinID, panitiaID)
if err == nil {
    // Broadcast happens automatically in ServiceWithWebSocket
    // Students will receive CHECKIN_UPDATED event
}
```

---

## 🔍 Troubleshooting

### Issue: "QR_INVALID" error
```bash
# Check QR format
echo "PEMIRA|TPS01|abc123" | grep -E "^PEMIRA\|[A-Z0-9]+\|[a-f0-9]{12}$"

# Verify in database
SELECT * FROM tps_qr WHERE qr_secret_suffix = 'abc123' AND is_active = true;
```

### Issue: "TPS_ACCESS_DENIED" 
```bash
# Check panitia assignment
SELECT * FROM tps_panitia WHERE tps_id = 1 AND user_id = 102;

# Assign if missing
INSERT INTO tps_panitia (tps_id, user_id, role) 
VALUES (1, 102, 'OPERATOR_PANEL');
```

### Issue: Check-in expired
```bash
# Check expiry settings (default 15 minutes)
# Increase if needed in service.go:
expiresAt := now.Add(15 * time.Minute)
```

---

## 📊 Monitoring Queries

### Active Check-ins
```sql
SELECT COUNT(*) FROM tps_checkins 
WHERE status = 'PENDING' 
AND created_at > NOW() - INTERVAL '1 hour';
```

### TPS Statistics
```sql
SELECT 
    t.code,
    t.name,
    COUNT(CASE WHEN c.status = 'PENDING' THEN 1 END) as pending,
    COUNT(CASE WHEN c.status = 'APPROVED' THEN 1 END) as approved,
    COUNT(CASE WHEN c.status = 'USED' THEN 1 END) as voted
FROM tps t
LEFT JOIN tps_checkins c ON c.tps_id = t.id
WHERE t.status = 'ACTIVE'
GROUP BY t.id, t.code, t.name;
```

### Expired Check-ins (cleanup)
```sql
UPDATE tps_checkins 
SET status = 'EXPIRED' 
WHERE status = 'APPROVED' 
AND expires_at < NOW();
```

---

## 🔐 Security Checklist

- [ ] JWT validation on all endpoints
- [ ] Role verification (ADMIN/STUDENT/TPS_OPERATOR)
- [ ] Panitia assignment check for panel access
- [ ] QR secret length ≥ 12 characters
- [ ] Check-in expiry enabled
- [ ] CORS configured for WebSocket
- [ ] Rate limiting on scan endpoint
- [ ] SQL injection prevention (parameterized queries)

---

## 🎯 Testing Checklist

### Unit Tests
```bash
go test ./internal/tps/... -v
```

### Manual API Tests
```bash
# Admin
curl http://localhost:8080/admin/tps
curl http://localhost:8080/admin/tps/1

# Student  
curl http://localhost:8080/tps/checkin/scan
curl http://localhost:8080/tps/checkin/status

# Panel
curl http://localhost:8080/tps/1/summary
curl http://localhost:8080/tps/1/checkins
```

### WebSocket Test
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/tps/1/queue');
ws.onmessage = (e) => console.log(JSON.parse(e.data));
```

---

## 📖 Error Code Reference

| Code | HTTP | Action |
|------|------|--------|
| TPS_NOT_FOUND | 404 | Check TPS ID exists |
| TPS_INACTIVE | 400 | Activate TPS first |
| QR_INVALID | 400 | Check QR format |
| QR_REVOKED | 400 | Regenerate QR |
| NOT_ELIGIBLE | 400 | Check DPT |
| ALREADY_VOTED | 409 | Voter already voted |
| CHECKIN_NOT_PENDING | 400 | Already processed |
| TPS_ACCESS_DENIED | 403 | Check panitia assignment |

---

## 🔗 Quick Links

- **Full API Docs**: `docs/TPS_API.md`
- **Implementation Summary**: `docs/TPS_IMPLEMENTATION_SUMMARY.md`
- **Module README**: `internal/tps/README.md`
- **Changelog**: `CHANGELOG_TPS.md`

---

## 💡 Pro Tips

1. **WebSocket vs Polling**: Use WebSocket for panel, polling for students
2. **Check-in Cleanup**: Run cron job to expire old check-ins
3. **QR Rotation**: Regenerate QR daily for security
4. **Monitoring**: Track approval time, queue length
5. **Caching**: Cache TPS list, active QRs
6. **Load Testing**: Test with 1000+ concurrent check-ins

---

**Last Updated**: November 19, 2025  
**Version**: 1.0.0  
**Quick Help**: Check `internal/tps/README.md` for detailed examples
