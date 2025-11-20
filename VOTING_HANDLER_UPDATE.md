# Voting Handler Update - November 2025

## 📝 Perubahan Terbaru

### Request Body Update

Handler voting sekarang mengharuskan `election_id` dikirim dalam request body untuk memberikan kontrol eksplisit ke client.

#### Online Voting
**Endpoint**: `POST /api/v1/voting/online/cast`

**Request Body**:
```json
{
  "election_id": 1,
  "candidate_id": 2
}
```

**Response** (Success):
```json
{
  "message": "Suara online berhasil direkam."
}
```

#### TPS Voting
**Endpoint**: `POST /api/v1/voting/tps/cast`

**Request Body**:
```json
{
  "election_id": 1,
  "candidate_id": 2,
  "tps_id": 3
}
```

**Response** (Success):
```json
{
  "message": "Suara TPS berhasil direkam."
}
```

### Handler Changes

1. **auth.FromContext()**: Handler sekarang menggunakan `auth.FromContext(ctx)` untuk mendapatkan `AuthUser` dari context
2. **Explicit election_id**: Client harus mengirim `election_id` dalam request body
3. **Simple Response**: Success response hanya mengembalikan message, bukan full receipt
4. **Voter Mapping Check**: Service layer memeriksa apakah user sudah terhubung dengan voter

### Service Layer Changes

Signature method service berubah:

**Before**:
```go
CastOnlineVote(ctx, voterID, candidateID int64) (*VoteReceipt, error)
CastTPSVote(ctx, voterID, candidateID int64) (*VoteReceipt, error)
```

**After**:
```go
CastOnlineVote(ctx, authUser auth.AuthUser, req CastOnlineVoteRequest) error
CastTPSVote(ctx, authUser auth.AuthUser, req CastTPSVoteRequest) error
```

### Error Handling

Ditambahkan error baru:
- `ErrVoterMappingMissing` - Ketika user belum terhubung dengan data voter

HTTP Response:
```json
{
  "code": "VOTER_MAPPING_MISSING",
  "message": "Akun ini belum terhubung dengan data pemilih."
}
```

### Integration Example

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"student1","password":"pass123"}' \
  | jq -r '.data.access_token')

# 2. Get current election
ELECTION_ID=$(curl -s http://localhost:8080/api/v1/elections/current \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data.id')

# 3. Cast vote
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"election_id\": $ELECTION_ID,
    \"candidate_id\": 1
  }"
```

### Benefits

1. **Explicit Election Selection**: Client bisa memilih election mana yang akan di-vote
2. **Better Error Handling**: Error mapping lebih jelas dengan auth.AuthUser
3. **Simpler Response**: Response success tidak perlu membawa data berat
4. **Cleaner Code**: Separation of concerns lebih baik antara handler dan service

---

**Updated**: 2025-11-20
**Commit**: f54b717
