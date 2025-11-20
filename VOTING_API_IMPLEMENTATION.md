# Voting API Implementation

Implementasi lengkap sistem voting untuk Pemira API dengan dukungan ONLINE dan TPS voting.

## 📋 Overview

Sistem voting ini mendukung dua metode voting:
1. **ONLINE Voting** - Pemilih voting secara online melalui web/mobile
2. **TPS Voting** - Pemilih voting di TPS setelah check-in disetujui

## 🔐 Security Features

- **Row-level Locking**: Menggunakan `FOR UPDATE` untuk mencegah double voting
- **Transaction Safety**: Semua operasi voting dalam transaction
- **Token-based Receipt**: Setiap vote menghasilkan unique token untuk verifikasi
- **Audit Trail**: Semua voting activity dicatat dalam audit log
- **No Vote Revealing**: Sistem tidak menyimpan kaitan langsung voter-candidate yang mudah di-query

## 📁 Struktur Database

### Table: voter_status
```sql
CREATE TABLE voter_status (
    id                BIGSERIAL PRIMARY KEY,
    election_id       BIGINT NOT NULL REFERENCES elections(id),
    voter_id          BIGINT NOT NULL REFERENCES voters(id),
    
    is_eligible       BOOLEAN NOT NULL DEFAULT TRUE,
    has_voted         BOOLEAN NOT NULL DEFAULT FALSE,
    
    voting_method     voting_method NULL,  -- 'ONLINE' | 'TPS'
    tps_id            BIGINT NULL REFERENCES tps(id),
    voted_at          TIMESTAMPTZ NULL,
    vote_token_hash   TEXT NULL,
    
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (election_id, voter_id),
    CHECK (
        (has_voted = FALSE AND voting_method IS NULL AND tps_id IS NULL AND voted_at IS NULL)
     OR (has_voted = TRUE AND voting_method IS NOT NULL AND voted_at IS NOT NULL)
    )
);
```

### Table: votes
```sql
CREATE TABLE votes (
    id              BIGSERIAL PRIMARY KEY,
    election_id     BIGINT NOT NULL REFERENCES elections(id),
    candidate_id    BIGINT NOT NULL REFERENCES candidates(id),
    
    token_hash      TEXT NOT NULL UNIQUE,
    channel         vote_channel NOT NULL,  -- 'ONLINE' | 'TPS'
    tps_id          BIGINT NULL REFERENCES tps(id),
    
    cast_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Table: vote_tokens
```sql
CREATE TABLE vote_tokens (
    id               BIGSERIAL PRIMARY KEY,
    election_id      BIGINT NOT NULL REFERENCES elections(id),
    voter_id         BIGINT NOT NULL REFERENCES voters(id),
    token            TEXT NOT NULL UNIQUE,
    issued_at        TIMESTAMPTZ NOT NULL,
    used_at          TIMESTAMPTZ NULL,
    method           TEXT NOT NULL,  -- 'ONLINE' | 'TPS'
    tps_id           BIGINT NULL REFERENCES tps(id)
);
```

## 🛠️ Core Components

### 1. Repository Layer

#### VoterRepository
```go
type VoterRepository interface {
    // Lock voter_status row untuk prevent race condition
    GetStatusForUpdate(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatusEntity, error)
    
    // Update status setelah voting
    UpdateStatus(ctx context.Context, tx pgx.Tx, status *VoterStatusEntity) error
}
```

#### VoteRepository
```go
type VoteRepository interface {
    // Insert vote token untuk audit trail
    InsertToken(ctx context.Context, tx pgx.Tx, token *VoteToken) error
    
    // Insert actual vote
    InsertVote(ctx context.Context, tx pgx.Tx, vote *Vote) error
    
    // Mark token sebagai used
    MarkTokenUsed(ctx context.Context, tx pgx.Tx, electionID int64, tokenHash string, usedAt time.Time) error
    
    // Get approved TPS check-in untuk TPS voting
    GetLatestApprovedCheckin(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*tps.TPSCheckin, error)
    
    // Get TPS info
    GetTPSByID(ctx context.Context, tx pgx.Tx, tpsID int64) (*tps.TPS, error)
    
    // Mark check-in sebagai used
    MarkCheckinUsed(ctx context.Context, tx pgx.Tx, checkinID int64, usedAt time.Time) error
}
```

#### CandidateRepository
```go
type CandidateRepository interface {
    // Get candidate dalam transaction untuk validasi
    GetByIDWithTx(ctx context.Context, tx pgx.Tx, candidateID int64) (*candidate.Candidate, error)
}
```

### 2. Service Layer

#### Service Structure
```go
type Service struct {
    db            *pgxpool.Pool
    electionRepo  election.Repository
    voterRepo     VoterRepository
    candidateRepo CandidateRepository
    voteRepo      VoteRepository
    statsRepo     VoteStatsRepository
    auditSvc      AuditService
}
```

#### Core Methods

**CastOnlineVote**
```go
func (s *Service) CastOnlineVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error)
```

Flow:
1. Get current election
2. Validate election status (must be VOTING_OPEN)
3. Validate online voting enabled
4. Execute castVote in transaction

**CastTPSVote**
```go
func (s *Service) CastTPSVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error)
```

Flow:
1. Get current election
2. Validate election status (must be VOTING_OPEN)
3. Validate TPS voting enabled
4. Get & validate latest approved TPS check-in
5. Validate check-in not expired (15 min TTL)
6. Execute castVote in transaction
7. Mark check-in as used

**castVote (Core Transaction Logic)**
```go
func (s *Service) castVote(
    ctx context.Context,
    electionID, voterID, candidateID int64,
    channel string,
    tpsID *int64,
) (*VoteResultEntity, error)
```

Transaction steps:
1. **Lock voter_status** with `FOR UPDATE`
2. **Check eligibility**: is_eligible = true, has_voted = false
3. **Validate candidate**: exists, active, same election
4. **Generate token hash**: unique random token
5. **Insert vote_token**: for audit trail
6. **Insert vote**: actual vote record
7. **Update voter_status**: set has_voted, voting_method, voted_at, token_hash
8. **Update stats** (optional): increment candidate count
9. **Audit log** (async): record voting action
10. **Build result**: return receipt with token

## 🌐 API Endpoints

### POST /api/v1/voting/online/cast

**Description**: Cast online vote

**Auth**: Required (JWT) - Student only

**Request**:
```json
{
  "candidate_id": 1
}
```

**Response** (200 OK):
```json
{
  "data": {
    "election_id": 1,
    "voter_id": 123,
    "method": "ONLINE",
    "voted_at": "2025-11-20T15:30:00Z",
    "receipt": {
      "token_hash": "vt_a1b2c3d4e5f6",
      "note": "Your vote has been recorded securely"
    }
  }
}
```

**Errors**:
- `401 UNAUTHORIZED` - Token tidak valid
- `403 FORBIDDEN` - Bukan role STUDENT
- `404 ELECTION_NOT_FOUND` - Pemilu aktif tidak ditemukan
- `400 ELECTION_NOT_OPEN` - Fase voting belum/sudah ditutup
- `400 METHOD_NOT_ALLOWED` - Online voting tidak diizinkan
- `403 NOT_ELIGIBLE` - Tidak termasuk DPT
- `409 ALREADY_VOTED` - Sudah voting
- `404 CANDIDATE_NOT_FOUND` - Kandidat tidak ditemukan
- `400 CANDIDATE_INACTIVE` - Kandidat tidak aktif

### POST /api/v1/voting/tps/cast

**Description**: Cast TPS vote after check-in approval

**Auth**: Required (JWT) - Student only

**Request**:
```json
{
  "candidate_id": 2
}
```

**Response** (200 OK):
```json
{
  "data": {
    "election_id": 1,
    "voter_id": 123,
    "method": "TPS",
    "voted_at": "2025-11-20T15:30:00Z",
    "tps": {
      "id": 3,
      "code": "TPS-003",
      "name": "TPS Aula Utama"
    },
    "receipt": {
      "token_hash": "vt_x9y8z7w6v5u4",
      "note": "Your vote has been recorded securely at TPS"
    }
  }
}
```

**Errors**:
- Same as online voting, plus:
- `400 TPS_CHECKIN_NOT_FOUND` - Belum check-in TPS yang valid
- `400 TPS_CHECKIN_NOT_APPROVED` - Check-in belum disetujui
- `400 CHECKIN_EXPIRED` - Waktu validasi check-in habis (>15 min)
- `404 TPS_NOT_FOUND` - TPS tidak ditemukan

### GET /api/v1/voting/tps/status

**Description**: Check TPS voting eligibility

**Auth**: Required (JWT) - Student only

**Response** (200 OK):
```json
{
  "data": {
    "eligible": true,
    "tps": {
      "id": 3,
      "code": "TPS-003",
      "name": "TPS Aula Utama"
    },
    "expires_at": "2025-11-20T15:45:00Z"
  }
}
```

### GET /api/v1/voting/receipt

**Description**: Get voting receipt (without revealing candidate)

**Auth**: Required (JWT) - Student only

**Response** (200 OK):
```json
{
  "data": {
    "has_voted": true,
    "election_id": 1,
    "method": "ONLINE",
    "voted_at": "2025-11-20T15:30:00Z",
    "receipt": {
      "token_hash": "vt_a1b2c3d4e5f6",
      "note": "Your vote has been recorded"
    }
  }
}
```

## 🔄 Voting Flow

### Online Voting Flow
```
1. User login → Get JWT token
2. Check election status → GET /elections/{id}/me/status
3. If eligible & online_allowed:
   - Select candidate
   - POST /voting/online/cast
4. Receive receipt with token_hash
5. Can verify vote with GET /voting/receipt
```

### TPS Voting Flow
```
1. User arrives at TPS
2. TPS Operator scans QR / manual check-in
3. TPS Operator approves check-in
4. User opens voting app
5. Check TPS status → GET /voting/tps/status
6. If eligible (approved & not expired):
   - Select candidate  
   - POST /voting/tps/cast
7. Receive receipt with token_hash + TPS info
8. Check-in marked as USED
```

## 🛡️ Validation & Business Rules

### Pre-voting Validation
1. **Election Status**: Must be in VOTING_OPEN phase
2. **Voting Method**: Online/TPS must be enabled for election
3. **Voter Eligibility**: is_eligible = true in voter_status
4. **No Double Vote**: has_voted must be false
5. **Candidate Active**: Candidate must be active and belong to same election

### TPS-specific Validation
1. **Check-in Required**: Must have approved TPS check-in
2. **Not Expired**: Check-in expires_at > now (15 min TTL)
3. **Not Used**: Check-in status != USED

### Transaction Guarantees
1. **Atomicity**: Vote + Status Update + Token creation = single transaction
2. **Isolation**: Row-level lock prevents concurrent voting
3. **Consistency**: CHECK constraints ensure data integrity
4. **Durability**: Commit only after all steps succeed

## 📊 Vote Token System

### Purpose
- **Audit Trail**: Track who voted when without revealing choice
- **Verification**: Voter can verify their vote was recorded
- **Anonymity**: Token doesn't reveal candidate choice

### Token Format
```
vt_<12_hex_chars>
Example: vt_a1b2c3d4e5f6
```

### Token Generation
```go
func generateTokenHash(electionID, voterID int64) string {
    randomBytes := make([]byte, 32)
    rand.Read(randomBytes)
    
    data := append(randomBytes, []byte(fmt.Sprintf("%d:%d", electionID, voterID))...)
    hash := sha256.Sum256(data)
    
    return "vt_" + hex.EncodeToString(hash[:12])
}
```

## 🔍 Error Handling

### Domain Errors
```go
var (
    ErrElectionNotFound       = errors.New("election not found")
    ErrElectionNotOpen        = errors.New("election is not open for voting")
    ErrNotEligible            = errors.New("voter is not eligible")
    ErrAlreadyVoted           = errors.New("voter has already voted")
    ErrCandidateNotFound      = errors.New("candidate not found")
    ErrCandidateInactive      = errors.New("candidate is not active")
    ErrMethodNotAllowed       = errors.New("voting method not allowed")
    ErrTPSCheckinNotFound     = errors.New("TPS check-in not found")
    ErrTPSCheckinNotApproved  = errors.New("TPS check-in not approved")
    ErrCheckinExpired         = errors.New("TPS check-in has expired")
    ErrTPSNotFound            = errors.New("TPS not found")
)
```

### HTTP Error Mapping
Handler `handleVotingError()` maps domain errors to appropriate HTTP responses with Indonesian messages.

## 🧪 Testing

### Manual Testing with curl

**1. Login as Student**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student123",
    "password": "password123"
  }'
```

**2. Cast Online Vote**
```bash
curl -X POST http://localhost:8080/api/v1/voting/online/cast \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "candidate_id": 1
  }'
```

**3. Get Voting Receipt**
```bash
curl http://localhost:8080/api/v1/voting/receipt \
  -H "Authorization: Bearer <access_token>"
```

## 📝 Integration Points

### With Auth System
- Uses JWT middleware for authentication
- Extracts voterID from token claims
- Enforces STUDENT role for voting endpoints

### With Election System
- Queries election status and configuration
- Validates online_enabled / tps_enabled flags
- Uses election_id for all operations

### With TPS System
- Validates TPS check-in approval
- Enforces 15-minute expiry window
- Marks check-in as USED after voting
- Links vote to specific TPS

### With Voter System
- Reads voter_status for eligibility
- Updates voter_status after voting
- Maintains has_voted flag

## 🔐 Security Considerations

1. **No Direct Vote Query**: votes table tidak punya voter_id
2. **Token-based Audit**: Link via token_hash, bukan voter_id
3. **Row Locking**: FOR UPDATE prevents race conditions
4. **Transaction Safety**: All-or-nothing voting operation
5. **Method Separation**: Clear separation between ONLINE vs TPS
6. **Expiry Enforcement**: TPS check-ins expire after 15 minutes
7. **Single Vote**: CHECK constraint enforces voting rules

## 📈 Future Enhancements

- [ ] Vote revocation/change (within time window)
- [ ] Multi-round voting support
- [ ] Ranked-choice voting
- [ ] Vote verification via blockchain
- [ ] Real-time vote count streaming
- [ ] Advanced fraud detection
- [ ] Voter notification system

---

**Status**: ✅ Implemented & Committed
**Last Updated**: 2025-11-20
