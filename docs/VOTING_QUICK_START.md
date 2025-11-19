# Voting Service - Quick Start Guide

## Konsep Implementasi

Service voting menggunakan **repository pattern** dengan **transaction safety** untuk mencegah double voting.

## Struktur Utama

```
internal/voting/
├── service.go                  # Service utama dengan CastOnlineVote & CastTPSVote
├── repository.go               # Repository interfaces
├── repository_voter.go         # Voter status repository (dengan row locking)
├── repository_candidate.go     # Candidate repository
├── repository_vote.go          # Vote & token repository
├── repository_stats.go         # Vote statistics repository
├── entity.go                   # Entity models
├── dto.go                      # DTOs untuk HTTP
├── errors.go                   # Custom errors
├── audit.go                    # Audit service interface
├── token.go                    # Token generation utilities
└── transaction.go              # Legacy transaction helper
```

## Key Components

### 1. Service Initialization

```go
votingService := voting.NewVotingService(
    db,                              // *pgxpool.Pool
    electionRepo,                    // election.Repository
    voting.NewVoterRepository(),     // VoterRepository
    voting.NewCandidateRepository(), // CandidateRepository
    voting.NewVoteRepository(),      // VoteRepository
    voting.NewVoteStatsRepository(), // VoteStatsRepository
    voting.NewAuditService(),        // AuditService
)
```

### 2. Cast Online Vote

```go
receipt, err := votingService.CastOnlineVote(ctx, voterID, candidateID)
```

**Validations:**
- ✅ Election active
- ✅ Online mode enabled
- ✅ Voter eligible & not voted
- ✅ Candidate active

### 3. Cast TPS Vote

```go
receipt, err := votingService.CastTPSVote(ctx, voterID, candidateID)
```

**Validations:**
- ✅ Election active
- ✅ TPS mode enabled
- ✅ TPS check-in APPROVED
- ✅ Check-in not expired (< 15 min)
- ✅ Voter eligible & not voted
- ✅ Candidate active

## Transaction Safety

### Row-Level Locking

```go
// Di VoterRepository.GetStatusForUpdate()
SELECT id, election_id, voter_id, has_voted, status, ...
FROM voter_election_status
WHERE election_id = $1 AND voter_id = $2
FOR UPDATE  -- 🔒 Lock baris ini
```

**Benefit:** Concurrent requests dari voter yang sama akan di-queue, mencegah race condition.

### Core Vote Flow

```
1. BEGIN TRANSACTION
2. LOCK voter_status (FOR UPDATE) 
3. CHECK has_voted → jika true: ROLLBACK + return ErrAlreadyVoted
4. VALIDATE candidate
5. GENERATE secure token
6. INSERT vote_tokens (voter receipt)
7. INSERT votes (anonymous)
8. UPDATE voter_status (has_voted=true)
9. UPDATE vote_stats (optional)
10. INSERT audit_logs (optional)
11. COMMIT TRANSACTION
```

## Error Mapping untuk HTTP Handler

```go
switch {
case errors.Is(err, voting.ErrAlreadyVoted):
    return 409 Conflict
    
case errors.Is(err, voting.ErrElectionNotOpen):
    return 400 Bad Request
    
case errors.Is(err, voting.ErrNotEligible):
    return 403 Forbidden
    
case errors.Is(err, voting.ErrCandidateNotFound):
    return 404 Not Found
    
case errors.Is(err, voting.ErrCandidateInactive):
    return 400 Bad Request
    
case errors.Is(err, voting.ErrMethodNotAllowed):
    return 400 Bad Request
    
case errors.Is(err, voting.ErrTPSCheckinNotFound):
    return 400 Bad Request
    
case errors.Is(err, voting.ErrTPSCheckinNotApproved):
    return 400 Bad Request
    
case errors.Is(err, voting.ErrCheckinExpired):
    return 400 Bad Request
    
default:
    return 500 Internal Server Error
}
```

## Response Format

### Success Response

```json
{
  "election_id": 1,
  "voter_id": 12345,
  "method": "ONLINE",
  "voted_at": "2024-11-20T01:23:45Z",
  "receipt": {
    "token_hash": "vt_a1b2c3d4e5f6",
    "note": "Your vote has been recorded securely"
  },
  "tps": null
}
```

### TPS Vote Response

```json
{
  "election_id": 1,
  "voter_id": 12345,
  "method": "TPS",
  "voted_at": "2024-11-20T01:23:45Z",
  "receipt": {
    "token_hash": "vt_x1y2z3a4b5c6",
    "note": "Your vote has been recorded securely at TPS"
  },
  "tps": {
    "id": 5,
    "code": "TPS001",
    "name": "TPS Gedung A"
  }
}
```

## Security Features

### 1. Anonymity
- Vote record tidak menyimpan voter_id
- Link via token_hash yang secure random
- Tidak bisa trace vote → voter

### 2. Receipt Verification
- Voter dapat verifikasi token_hash mereka
- Token permanent (tidak expire)
- Bisa implement public verification endpoint

### 3. Audit Trail
- Setiap voting di-log (actor, action, timestamp)
- Metadata lengkap (election_id, channel, tps_id)
- Dapat async untuk performance

## Testing Strategy

### 1. Unit Tests

```go
func TestCastOnlineVote_Success(t *testing.T) {
    // Test successful online vote
}

func TestCastOnlineVote_AlreadyVoted(t *testing.T) {
    // Test duplicate vote prevention
}

func TestCastTPSVote_ExpiredCheckin(t *testing.T) {
    // Test expired check-in rejection
}
```

### 2. Concurrency Test

```go
func TestConcurrentVotes_PreventDouble(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    // Spawn 10 goroutines mencoba vote bersamaan
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := service.CastOnlineVote(ctx, voterID, candidateID)
            if err != nil {
                errors <- err
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // Assert: hanya 1 sukses, 9 lainnya ErrAlreadyVoted
    successCount := 0
    alreadyVotedCount := 0
    
    for err := range errors {
        if errors.Is(err, voting.ErrAlreadyVoted) {
            alreadyVotedCount++
        }
    }
    
    assert.Equal(t, 1, successCount)
    assert.Equal(t, 9, alreadyVotedCount)
}
```

## Integration dengan Existing Code

### Option 1: Replace Existing Service (Recommended)

Ganti service initialization di `cmd/api/main.go` atau handler setup:

```go
// Old
votingService := voting.NewService(votingRepo)

// New
votingService := voting.NewVotingService(
    db,
    electionRepo,
    voting.NewVoterRepository(),
    voting.NewCandidateRepository(),
    voting.NewVoteRepository(),
    voting.NewVoteStatsRepository(),
    voting.NewAuditService(),
)
```

### Option 2: Side-by-Side (Migration)

Keep both, gradually migrate:

```go
// Old service untuk backward compatibility
oldVotingService := voting.NewService(votingRepo)

// New service untuk new endpoints
newVotingService := voting.NewVotingService(...)

handler := voting.NewHandler(oldVotingService, newVotingService)
```

## Next Steps

1. **Setup Database:**
   - Pastikan tabel `voter_election_status`, `votes`, `vote_tokens`, dll ada
   - Add indexes untuk performance

2. **Update HTTP Handlers:**
   - Inject new service ke handlers
   - Map errors ke HTTP status codes

3. **Testing:**
   - Write unit tests
   - Test concurrent voting
   - Test TPS flow end-to-end

4. **Monitoring:**
   - Add metrics (vote latency, error rate)
   - Alert on high error rates
   - Monitor lock contention

5. **Documentation:**
   - API documentation
   - Sequence diagrams
   - Runbook untuk troubleshooting

## Troubleshooting

### Issue: "voter has already voted" tapi UI belum update

**Solution:** Check voter_election_status.has_voted field, bisa jadi race condition sebelumnya.

### Issue: Transaction timeout

**Solution:** 
1. Check database connection pool size
2. Reduce transaction scope
3. Add index on frequently queried columns

### Issue: TPS check-in expired

**Solution:** Check `tps_checkins.expires_at`, pastikan TTL 15 menit cukup.

## Performance Tips

1. **Connection Pooling:** Set `max_connections` appropriate untuk load
2. **Indexes:** Pastikan index ada di `(election_id, voter_id)` columns
3. **Async Audit:** Implementasi audit logging secara async
4. **Stats Cache:** Consider Redis untuk live vote counts
5. **Read Replicas:** Query non-transactional dari read replica

## Support

Untuk pertanyaan atau issue:
1. Check dokumentasi lengkap di `VOTING_SERVICE_IMPLEMENTATION.md`
2. Review error logs
3. Check audit_logs table untuk debugging
