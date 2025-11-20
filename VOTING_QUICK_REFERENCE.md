# Voting API - Quick Reference

Quick reference untuk menggunakan Voting API di Pemira system.

## 🚀 Quick Start

### 1. Login & Get Token
```bash
# Login as student
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student123",
    "password": "password123"
  }'

# Response
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "ref_abc123...",
    "expires_in": 3600
  }
}

# Set token untuk request berikutnya
export TOKEN="eyJhbGc..."
```

### 2. Cast Online Vote
```bash
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "candidate_id": 1
  }'
```

### 3. Cast TPS Vote
```bash
# After TPS check-in approved
curl -X POST http://localhost:8080/api/v1/voting/tps/cast \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "candidate_id": 2
  }'
```

### 4. Get Receipt
```bash
curl http://localhost:8080/api/v1/voting/receipt \
  -H "Authorization: Bearer $TOKEN"
```

## 📋 API Endpoints

| Method | Endpoint | Auth | Role | Description |
|--------|----------|------|------|-------------|
| POST | `/api/v1/voting/online/cast` | ✅ | STUDENT | Cast online vote |
| POST | `/api/v1/voting/tps/cast` | ✅ | STUDENT | Cast TPS vote |
| GET | `/api/v1/voting/tps/status` | ✅ | STUDENT | Check TPS eligibility |
| GET | `/api/v1/voting/receipt` | ✅ | STUDENT | Get vote receipt |

## 📝 Request/Response Schemas

### Cast Vote Request
```typescript
{
  "candidate_id": number  // Required, min: 1
}
```

### Vote Receipt Response
```typescript
{
  "data": {
    "election_id": number,
    "voter_id": number,
    "method": "ONLINE" | "TPS",
    "voted_at": string,     // ISO 8601
    "tps": {                // Only for TPS voting
      "id": number,
      "code": string,
      "name": string
    },
    "receipt": {
      "token_hash": string, // Format: vt_<12_hex>
      "note": string
    }
  }
}
```

### TPS Status Response
```typescript
{
  "data": {
    "eligible": boolean,
    "tps": {
      "id": number,
      "code": string,
      "name": string
    },
    "expires_at": string,   // ISO 8601, optional
    "reason": string        // Optional, if not eligible
  }
}
```

### Receipt Response
```typescript
{
  "data": {
    "has_voted": boolean,
    "election_id": number,
    "method": "ONLINE" | "TPS",
    "voted_at": string,
    "tps": {...},           // Optional
    "receipt": {
      "token_hash": string,
      "note": string
    }
  }
}
```

## ⚠️ Error Codes

### Common Errors
| Code | Status | Description |
|------|--------|-------------|
| `UNAUTHORIZED` | 401 | Token tidak valid / expired |
| `FORBIDDEN` | 403 | Bukan role STUDENT |
| `VALIDATION_ERROR` | 400/422 | Request body tidak valid |

### Voting Errors
| Code | Status | Description |
|------|--------|-------------|
| `ELECTION_NOT_FOUND` | 404 | Pemilu aktif tidak ditemukan |
| `ELECTION_NOT_OPEN` | 400 | Fase voting belum/sudah ditutup |
| `METHOD_NOT_ALLOWED` | 400 | Metode voting tidak diizinkan |
| `NOT_ELIGIBLE` | 403 | Tidak termasuk DPT |
| `ALREADY_VOTED` | 409 | Sudah menggunakan hak suara |
| `CANDIDATE_NOT_FOUND` | 404 | Kandidat tidak ditemukan |
| `CANDIDATE_INACTIVE` | 400 | Kandidat tidak aktif |

### TPS-specific Errors
| Code | Status | Description |
|------|--------|-------------|
| `TPS_CHECKIN_NOT_FOUND` | 400 | Belum check-in TPS |
| `TPS_CHECKIN_NOT_APPROVED` | 400 | Check-in belum disetujui |
| `CHECKIN_EXPIRED` | 400 | Waktu validasi habis (>15 min) |
| `TPS_NOT_FOUND` | 404 | TPS tidak ditemukan |

## 🔄 Voting Flow

### Online Voting
```
1. Login → Get JWT token
2. Check eligibility → GET /elections/{id}/me/status
   Response: {
     "eligible": true,
     "has_voted": false,
     "online_allowed": true
   }
3. Select candidate in UI
4. Submit vote → POST /voting/online/cast
5. Show receipt with token_hash
6. Optionally verify → GET /voting/receipt
```

### TPS Voting
```
1. Voter arrives at TPS
2. TPS Operator scans QR / manual entry
3. TPS Operator approves check-in
   → tps_checkins.status = 'APPROVED'
   → expires_at = now + 15 minutes

4. Voter opens app & logs in
5. Check TPS status → GET /voting/tps/status
   Response: {
     "eligible": true,
     "tps": {...},
     "expires_at": "..."
   }
   
6. If eligible:
   - Select candidate
   - Submit → POST /voting/tps/cast
   
7. Show receipt with TPS info
8. Check-in status → 'USED'
```

## 🔐 Authentication

### Required Header
```
Authorization: Bearer <access_token>
```

### Token Structure (JWT Claims)
```json
{
  "user_id": 123,
  "role": "STUDENT",
  "voter_id": 456,
  "exp": 1700000000,
  "iat": 1699996400
}
```

## 🛡️ Business Rules

### Pre-voting Checks
1. ✅ Election must be in `VOTING_OPEN` phase
2. ✅ Voting method (ONLINE/TPS) must be enabled
3. ✅ Voter must be eligible (`is_eligible = true`)
4. ✅ Voter must not have voted (`has_voted = false`)
5. ✅ Candidate must be active and in same election

### TPS-specific Rules
1. ✅ Must have approved TPS check-in
2. ✅ Check-in must not be expired (15 min window)
3. ✅ Check-in must not be already used

### Post-voting State
After successful vote:
- `voter_status.has_voted` → `true`
- `voter_status.voting_method` → `'ONLINE'` or `'TPS'`
- `voter_status.voted_at` → current timestamp
- `voter_status.vote_token_hash` → generated token
- `voter_status.tps_id` → TPS ID (for TPS voting)
- Vote token issued
- Actual vote recorded (anonymized)

## 📊 Database State

### Before Voting
```sql
-- voter_status
{
  "has_voted": false,
  "voting_method": null,
  "tps_id": null,
  "voted_at": null,
  "vote_token_hash": null
}
```

### After Online Voting
```sql
-- voter_status
{
  "has_voted": true,
  "voting_method": "ONLINE",
  "tps_id": null,
  "voted_at": "2025-11-20 15:30:00+00",
  "vote_token_hash": "vt_a1b2c3d4e5f6"
}

-- vote_tokens
INSERT (election_id, voter_id, token, method, issued_at)

-- votes
INSERT (election_id, candidate_id, token_hash, channel, cast_at)
```

### After TPS Voting
```sql
-- voter_status
{
  "has_voted": true,
  "voting_method": "TPS",
  "tps_id": 3,
  "voted_at": "2025-11-20 15:30:00+00",
  "vote_token_hash": "vt_x9y8z7w6v5u4"
}

-- tps_checkins
UPDATE status = 'USED' WHERE id = ...

-- vote_tokens + votes (same as online)
```

## 🧪 Testing Scenarios

### Test 1: Successful Online Vote
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"student1","password":"pass123"}' \
  | jq -r '.data.access_token')

# 2. Vote
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"candidate_id":1}'

# Expected: 200 OK with receipt
```

### Test 2: Double Vote (Should Fail)
```bash
# Vote again with same token
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"candidate_id":2}'

# Expected: 409 ALREADY_VOTED
```

### Test 3: Invalid Candidate
```bash
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"candidate_id":9999}'

# Expected: 404 CANDIDATE_NOT_FOUND
```

### Test 4: TPS Voting Without Check-in
```bash
curl -X POST http://localhost:8080/api/v1/voting/tps/cast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"candidate_id":1}'

# Expected: 400 TPS_CHECKIN_NOT_FOUND
```

## 🔍 Debugging Tips

### Check Voter Status
```sql
SELECT * FROM voter_status 
WHERE election_id = 1 AND voter_id = 123;
```

### Check if Vote Recorded
```sql
SELECT v.*, vt.token 
FROM votes v
JOIN vote_tokens vt ON v.token_hash = vt.token
WHERE vt.voter_id = 123 AND v.election_id = 1;
```

### Check TPS Check-in
```sql
SELECT * FROM tps_checkins
WHERE voter_id = 123 AND election_id = 1
ORDER BY created_at DESC LIMIT 1;
```

### Verify Token
```bash
# Get receipt
curl http://localhost:8080/api/v1/voting/receipt \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.receipt.token_hash'

# Cross-check with DB
psql -c "SELECT * FROM vote_tokens WHERE token = 'vt_...'"
```

## 🚨 Common Issues

### Issue: "UNAUTHORIZED" 
**Cause**: Token expired or invalid  
**Solution**: Login again to get new token

### Issue: "ALREADY_VOTED"
**Cause**: Voter sudah voting sebelumnya  
**Solution**: Check voter_status, cannot vote twice

### Issue: "ELECTION_NOT_OPEN"
**Cause**: Election status bukan VOTING_OPEN  
**Solution**: Wait for voting phase or check election config

### Issue: "TPS_CHECKIN_NOT_APPROVED"
**Cause**: TPS check-in masih PENDING  
**Solution**: Wait for TPS operator to approve

### Issue: "CHECKIN_EXPIRED"
**Cause**: >15 minutes since check-in approval  
**Solution**: Check-in ulang di TPS

## 📚 Related Documentation

- **Full Implementation**: `VOTING_API_IMPLEMENTATION.md`
- **Setup Summary**: `VOTING_SETUP_SUMMARY.md`
- **Auth System**: `AUTH_IMPLEMENTATION.md`
- **Auth Quick Ref**: `AUTH_QUICK_REFERENCE.md`

---

**Version**: 1.0  
**Last Updated**: 2025-11-20
