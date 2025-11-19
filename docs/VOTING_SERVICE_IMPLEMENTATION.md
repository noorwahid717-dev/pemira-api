# Voting Service Implementation

## Overview

Implementasi lengkap voting service dengan konsep repository pattern dan transaction safety untuk mencegah double voting.

## Architecture

### 1. Service Layer (`internal/voting/service.go`)

Service utama dengan dependency injection:

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

**Constructor:**
```go
func NewVotingService(
    db *pgxpool.Pool,
    electionRepo election.Repository,
    voterRepo VoterRepository,
    candidateRepo CandidateRepository,
    voteRepo VoteRepository,
    statsRepo VoteStatsRepository,
    auditSvc AuditService,
) *Service
```

### 2. Repository Interfaces

#### VoterRepository (`repository_voter.go`)
- `GetStatusForUpdate(ctx, tx, electionID, voterID)` - Lock voter status dengan `FOR UPDATE`
- `UpdateStatus(ctx, tx, status)` - Update voter status setelah voting

#### CandidateRepository (`repository_candidate.go`)
- `GetByIDWithTx(ctx, tx, candidateID)` - Get candidate dalam transaction

#### VoteRepository (`repository_vote.go`)
- `InsertToken(ctx, tx, token)` - Insert vote token untuk receipt
- `InsertVote(ctx, tx, vote)` - Insert anonymous vote
- `MarkTokenUsed(ctx, tx, electionID, tokenHash, usedAt)` - Mark token sebagai used
- `GetLatestApprovedCheckin(ctx, tx, electionID, voterID)` - Get TPS check-in terbaru
- `GetTPSByID(ctx, tx, tpsID)` - Get TPS info
- `MarkCheckinUsed(ctx, tx, checkinID, usedAt)` - Mark check-in sebagai used

#### VoteStatsRepository (`repository_stats.go`)
- `IncrementCandidateCount(ctx, tx, electionID, candidateID, channel, tpsID)` - Update vote count

### 3. Core Methods

#### CastOnlineVote
```go
func (s *Service) CastOnlineVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error)
```

**Flow:**
1. Get current election
2. Validate election status (VOTING_OPEN)
3. Validate online mode enabled
4. Call `castVote()` with channel="ONLINE"
5. Return receipt

#### CastTPSVote
```go
func (s *Service) CastTPSVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error)
```

**Flow:**
1. Get current election
2. Validate election status
3. Validate TPS mode enabled
4. Get latest approved check-in
5. Validate check-in not expired (15 min TTL)
6. Get TPS info
7. Call `castVote()` with channel="TPS" and tpsID
8. Mark check-in as USED
9. Return receipt with TPS info

#### castVote (Core Logic)
```go
func (s *Service) castVote(
    ctx context.Context,
    electionID, voterID, candidateID int64,
    channel string,
    tpsID *int64,
) (*VoteResultEntity, error)
```

**Transaction Flow:**
1. **Lock voter_status** dengan `FOR UPDATE` (prevents race condition)
2. **Validate eligibility:** Check IsEligible dan !HasVoted
3. **Validate candidate:** Get candidate, check ElectionID match, check IsActive
4. **Generate token:** Secure random token hash
5. **Insert vote_tokens:** Untuk voter's receipt
6. **Insert votes:** Anonymous vote record
7. **Update voter_status:** Mark HasVoted=true, set method, tps_id, voted_at
8. **Update stats:** Increment candidate count (optional, for live count)
9. **Audit log:** Record voting action
10. **Build result:** Return VoteResultEntity dengan receipt

## Transaction Safety

### Row-Level Locking
```sql
SELECT id, election_id, voter_id, has_voted, status, voted_via, tps_id, voted_at, token_hash
FROM voter_election_status
WHERE election_id = $1 AND voter_id = $2
FOR UPDATE
```

**Mencegah:**
- Double voting dari concurrent requests
- Race conditions saat 2+ requests parallel dari voter yang sama

### Transaction Pattern
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

## Entity Models

### VoterStatusEntity
```go
type VoterStatusEntity struct {
    ID           int64
    ElectionID   int64
    VoterID      int64
    IsEligible   bool
    HasVoted     bool
    VotingMethod *string    // "ONLINE" | "TPS"
    TPSID        *int64
    VotedAt      *time.Time
    TokenHash    *string
    Status       string
}
```

### Vote
```go
type Vote struct {
    ID          int64
    ElectionID  int64
    CandidateID int64
    TokenHash   string
    Channel     string     // "ONLINE" | "TPS"
    TPSID       *int64
    CastAt      time.Time
}
```

### VoteToken
```go
type VoteToken struct {
    ID         int64
    ElectionID int64
    VoterID    int64
    TokenHash  string
    IssuedAt   time.Time
    UsedAt     *time.Time
    Method     string     // "ONLINE" | "TPS"
    TPSID      *int64
}
```

## Security Features

### 1. Token Generation
```go
func generateTokenHash(electionID, voterID int64) string
```
- Uses `crypto/rand` for secure random bytes
- SHA256 hash of random bytes + electionID:voterID
- Prefix "vt_" + 12 bytes hex

### 2. Anonymity
- Votes tidak link langsung ke voter
- Hanya link via token_hash
- Token hash secure random, tidak predictable

### 3. Audit Trail
```go
type AuditEntry struct {
    ActorVoterID *int64
    Action       string         // "CAST_VOTE_ONLINE" | "CAST_VOTE_TPS"
    EntityType   string         // "VOTE"
    EntityID     int64
    Metadata     map[string]any
    CreatedAt    time.Time
}
```

## Error Handling

Custom errors di `errors.go`:
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

## Usage Example

### Initialize Service
```go
db := pgxpool.New(ctx, connString)
electionRepo := election.NewPostgresRepository(db)
voterRepo := voting.NewVoterRepository()
candidateRepo := voting.NewCandidateRepository()
voteRepo := voting.NewVoteRepository()
statsRepo := voting.NewVoteStatsRepository()
auditSvc := voting.NewAuditService()

votingService := voting.NewVotingService(
    db,
    electionRepo,
    voterRepo,
    candidateRepo,
    voteRepo,
    statsRepo,
    auditSvc,
)
```

### Cast Online Vote
```go
receipt, err := votingService.CastOnlineVote(ctx, voterID, candidateID)
if err != nil {
    if errors.Is(err, voting.ErrAlreadyVoted) {
        // Handle already voted (409 Conflict)
    } else if errors.Is(err, voting.ErrElectionNotOpen) {
        // Handle election not open (400 Bad Request)
    }
    // Handle other errors
}

// Success: receipt.Receipt.TokenHash untuk voter's receipt
```

### Cast TPS Vote
```go
receipt, err := votingService.CastTPSVote(ctx, voterID, candidateID)
if err != nil {
    if errors.Is(err, voting.ErrTPSCheckinNotFound) {
        // Handle no check-in (400 Bad Request)
    } else if errors.Is(err, voting.ErrCheckinExpired) {
        // Handle expired check-in (400 Bad Request)
    }
    // Handle other errors
}

// Success: receipt dengan TPS info
```

## Database Schema Requirements

### Tables
1. **voter_election_status** - Voter eligibility & status
2. **votes** - Anonymous votes (token_hash, candidate_id, channel, tps_id)
3. **vote_tokens** - Vote receipts (voter_id, token_hash, method)
4. **vote_stats** - Materialized vote counts (optional)
5. **tps_checkins** - TPS check-in records
6. **tps** - TPS master data
7. **candidates** - Candidate master data
8. **elections** - Election master data
9. **audit_logs** - Audit trail (optional)

### Key Indexes
```sql
CREATE INDEX idx_voter_status_election_voter ON voter_election_status(election_id, voter_id);
CREATE INDEX idx_votes_election ON votes(election_id);
CREATE INDEX idx_votes_token ON votes(token_hash);
CREATE INDEX idx_vote_tokens_election_voter ON vote_tokens(election_id, voter_id);
CREATE INDEX idx_tps_checkins_election_voter ON tps_checkins(election_id, voter_id);
```

## Testing Checklist

### Unit Tests
- [ ] VoterRepository GetStatusForUpdate
- [ ] VoterRepository UpdateStatus
- [ ] CandidateRepository GetByIDWithTx
- [ ] VoteRepository InsertToken
- [ ] VoteRepository InsertVote
- [ ] Token generation uniqueness

### Integration Tests
- [ ] CastOnlineVote success flow
- [ ] CastTPSVote success flow
- [ ] Double voting prevention (concurrent requests)
- [ ] Invalid candidate rejection
- [ ] Expired check-in rejection
- [ ] Election not open rejection

### Stress Tests
- [ ] 1000 concurrent vote requests
- [ ] Row lock contention handling
- [ ] Transaction rollback scenarios

## Future Enhancements

1. **Async Audit Logging**
   - Queue-based audit log writes
   - Non-blocking audit trail

2. **Stats Caching**
   - Redis-based vote count cache
   - Asynchronous stats update

3. **Rate Limiting**
   - Per-voter rate limits
   - DDoS protection

4. **Vote Verification**
   - Public token verification endpoint
   - Zero-knowledge proof support

## Files Created/Modified

### New Files
- `internal/voting/repository_voter.go` - Voter status repository
- `internal/voting/repository_candidate.go` - Candidate repository
- `internal/voting/repository_vote.go` - Vote & token repository
- `internal/voting/repository_stats.go` - Vote stats repository
- `internal/voting/audit.go` - Audit service interface
- `internal/voting/errors.go` - Custom error definitions
- `internal/voting/token.go` - Token generation utilities

### Modified Files
- `internal/voting/entity.go` - Updated entity models
- `internal/voting/repository.go` - Added repository interfaces
- `internal/voting/service.go` - Implemented core voting logic
- `internal/voting/service_tps.go` - Removed duplicate errors
- `internal/election/http_handler.go` - Fixed syntax error

## Notes

1. **Transaction Isolation:** Uses default PostgreSQL READ COMMITTED isolation level
2. **Lock Timeout:** Default PostgreSQL lock timeout applies (consider tuning for high load)
3. **Token Expiry:** Vote tokens tidak expire (permanent receipt)
4. **Check-in TTL:** TPS check-in expires setelah 15 menit (configurable via expires_at)
5. **Audit Async:** Audit service dapat diimplementasikan secara async untuk performance
