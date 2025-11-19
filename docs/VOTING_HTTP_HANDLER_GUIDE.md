# Voting HTTP Handler Guide

## Overview

HTTP handlers untuk voting endpoints dengan chi router, clean error mapping, dan context-based authentication.

## Endpoints

### 1. POST /voting/online/cast
Cast vote via online method

### 2. POST /voting/tps/cast
Cast vote via TPS after check-in approval

## Architecture

```
HTTP Request
    ↓
Auth Middleware (extract & validate JWT, set context)
    ↓
Handler (validate body, get voterID from context)
    ↓
Service (business logic with transaction)
    ↓
Repository (database operations)
    ↓
Response (success or error mapping)
```

## Request & Response Format

### Request Body (Both Endpoints)

```json
{
  "candidate_id": 123
}
```

**Validation:**
- `candidate_id`: required, must be > 0

### Success Response (200 OK)

```json
{
  "data": {
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
}
```

### TPS Success Response

```json
{
  "data": {
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
}
```

### Error Response Format

```json
{
  "code": "ALREADY_VOTED",
  "message": "Anda sudah menggunakan hak suara untuk pemilu ini.",
  "details": null
}
```

## Error Mapping

| Domain Error | HTTP Status | Error Code | Message |
|-------------|-------------|------------|---------|
| `ErrElectionNotFound` | 404 | ELECTION_NOT_FOUND | Pemilu aktif tidak ditemukan. |
| `ErrElectionNotOpen` | 400 | ELECTION_NOT_OPEN | Fase voting belum dibuka atau sudah ditutup. |
| `ErrNotEligible` | 403 | NOT_ELIGIBLE | Anda tidak termasuk dalam DPT atau tidak berhak memilih. |
| `ErrAlreadyVoted` | 409 | ALREADY_VOTED | Anda sudah menggunakan hak suara untuk pemilu ini. |
| `ErrCandidateNotFound` | 404 | CANDIDATE_NOT_FOUND | Kandidat tidak ditemukan untuk pemilu ini. |
| `ErrCandidateInactive` | 400 | CANDIDATE_INACTIVE | Kandidat tidak aktif. |
| `ErrMethodNotAllowed` | 400 | METHOD_NOT_ALLOWED | Metode voting ini tidak diizinkan untuk pemilu sekarang. |
| `ErrTPSCheckinNotFound` | 400 | TPS_CHECKIN_NOT_FOUND | Anda belum melakukan check-in TPS yang valid. |
| `ErrTPSCheckinNotApproved` | 400 | TPS_CHECKIN_NOT_APPROVED | Check-in Anda belum disetujui panitia TPS. |
| `ErrCheckinExpired` | 400 | CHECKIN_EXPIRED | Waktu validasi check-in Anda sudah habis, silakan ulangi di TPS. |
| `ErrTPSNotFound` | 404 | TPS_NOT_FOUND | TPS tidak ditemukan. |
| Other | 500 | INTERNAL_ERROR | Terjadi kesalahan pada sistem. |

## Handler Implementation

### File Structure

```
internal/voting/
├── http_handler.go         # HTTP handlers & error mapping
├── service.go              # Business logic
├── repository.go           # Repository interfaces
└── errors.go               # Domain errors
```

### Handler Code

```go
type Handler struct {
    service  *Service
    validate *validator.Validate
}

func NewVotingHandler(svc *Service) *Handler {
    return &Handler{
        service:  svc,
        validate: validator.New(),
    }
}
```

### Context Key Usage

```go
// Get voter ID from context (set by auth middleware)
voterID, ok := ctxkeys.GetVoterID(ctx)
if !ok {
    response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", 
        "Token tidak valid atau tidak memiliki akses.", nil)
    return
}
```

### Request Validation

```go
// Parse body
var req CastVoteRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", 
        "Format body tidak valid.", nil)
    return
}

// Validate struct
if err := h.validate.Struct(req); err != nil {
    response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", 
        "candidate_id wajib diisi.", map[string]string{
            "field":      "candidate_id",
            "constraint": "required",
        })
    return
}
```

## Router Setup

### Option 1: Using RegisterRoutes

```go
// internal/http/router.go
func NewRouter(deps Dependencies) http.Handler {
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.CORS)
    
    // Voting routes (requires auth)
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthRequired)
        r.Use(middleware.StudentOnly)
        
        votingHandler := voting.NewVotingHandler(deps.VotingService)
        votingHandler.RegisterRoutes(r)
    })
    
    return r
}
```

### Option 2: Using Mount

```go
func NewRouter(deps Dependencies) http.Handler {
    r := chi.NewRouter()
    
    // Voting group
    r.Route("/voting", func(r chi.Router) {
        r.Use(middleware.AuthRequired)
        r.Use(middleware.StudentOnly)
        
        votingHandler := voting.NewVotingHandler(deps.VotingService)
        votingHandler.Mount(r)
    })
    
    return r
}
```

### Option 3: Direct Registration

```go
r.Route("/voting", func(r chi.Router) {
    r.Use(middleware.AuthRequired)
    r.Use(middleware.StudentOnly)
    
    handler := voting.NewVotingHandler(votingService)
    
    r.Post("/online/cast", handler.CastOnlineVote)
    r.Post("/tps/cast", handler.CastTPSVote)
})
```

## Middleware Requirements

### AuthRequired Middleware

```go
func AuthRequired(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract JWT from Authorization header
        token := extractToken(r)
        if token == "" {
            response.Unauthorized(w, "Token tidak ditemukan")
            return
        }
        
        // Validate & parse JWT
        claims, err := validateJWT(token)
        if err != nil {
            response.Unauthorized(w, "Token tidak valid atau sudah expire")
            return
        }
        
        // Set context values
        ctx := r.Context()
        ctx = context.WithValue(ctx, ctxkeys.UserIDKey, claims.UserID)
        ctx = context.WithValue(ctx, ctxkeys.VoterIDKey, claims.VoterID)
        ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, claims.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### StudentOnly Middleware

```go
func StudentOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role, ok := ctxkeys.GetUserRole(r.Context())
        if !ok || role != "STUDENT" {
            response.Forbidden(w, "Akses ditolak. Endpoint ini hanya untuk mahasiswa.")
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Service Initialization

### In main.go or dependency injection

```go
// Initialize repositories
voterRepo := voting.NewVoterRepository()
candidateRepo := voting.NewCandidateRepository()
voteRepo := voting.NewVoteRepository()
statsRepo := voting.NewVoteStatsRepository()
auditSvc := voting.NewAuditService()

// Initialize election repository (assuming already exists)
electionRepo := election.NewPostgresRepository(db)

// Initialize voting service
votingService := voting.NewVotingService(
    db,
    electionRepo,
    voterRepo,
    candidateRepo,
    voteRepo,
    statsRepo,
    auditSvc,
)

// Initialize handler
votingHandler := voting.NewVotingHandler(votingService)
```

## Testing Examples

### Unit Test: Handler

```go
func TestCastOnlineVote_Success(t *testing.T) {
    // Mock service
    mockService := &MockVotingService{}
    handler := voting.NewVotingHandler(mockService)
    
    // Setup expected response
    mockService.On("CastOnlineVote", mock.Anything, int64(123), int64(456)).
        Return(&voting.VoteReceipt{
            ElectionID: 1,
            VoterID:    123,
            Method:     "ONLINE",
            VotedAt:    time.Now(),
            Receipt: voting.ReceiptDetail{
                TokenHash: "vt_test123",
                Note:      "Success",
            },
        }, nil)
    
    // Create request with context
    body := `{"candidate_id": 456}`
    req := httptest.NewRequest("POST", "/voting/online/cast", strings.NewReader(body))
    ctx := context.WithValue(req.Context(), ctxkeys.VoterIDKey, int64(123))
    req = req.WithContext(ctx)
    
    // Execute
    w := httptest.NewRecorder()
    handler.CastOnlineVote(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}
```

### Integration Test

```go
func TestCastOnlineVote_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Initialize real service
    service := voting.NewVotingService(
        db,
        electionRepo,
        voterRepo,
        candidateRepo,
        voteRepo,
        statsRepo,
        auditSvc,
    )
    
    handler := voting.NewVotingHandler(service)
    
    // Create test data
    setupTestElection(t, db)
    setupTestVoter(t, db, 123)
    setupTestCandidate(t, db, 456)
    
    // Execute request
    body := `{"candidate_id": 456}`
    req := httptest.NewRequest("POST", "/voting/online/cast", strings.NewReader(body))
    ctx := context.WithValue(req.Context(), ctxkeys.VoterIDKey, int64(123))
    req = req.WithContext(ctx)
    
    w := httptest.NewRecorder()
    handler.CastOnlineVote(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var resp struct {
        Data struct {
            ElectionID int64  `json:"election_id"`
            VoterID    int64  `json:"voter_id"`
            Method     string `json:"method"`
        } `json:"data"`
    }
    json.NewDecoder(w.Body).Decode(&resp)
    
    assert.Equal(t, int64(1), resp.Data.ElectionID)
    assert.Equal(t, int64(123), resp.Data.VoterID)
    assert.Equal(t, "ONLINE", resp.Data.Method)
}
```

## cURL Examples

### Online Vote

```bash
curl -X POST http://localhost:8080/voting/online/cast \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"candidate_id": 123}'
```

### TPS Vote

```bash
curl -X POST http://localhost:8080/voting/tps/cast \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"candidate_id": 456}'
```

## Common Issues & Solutions

### Issue: "UNAUTHORIZED" response

**Cause:** JWT token not provided or invalid

**Solution:** 
1. Check Authorization header format: `Bearer <token>`
2. Verify JWT is valid and not expired
3. Ensure middleware sets context correctly

### Issue: "NOT_ELIGIBLE" response

**Cause:** Voter not in DPT or voter_status check fails

**Solution:**
1. Verify voter exists in `voter_election_status` table
2. Check `status = 'ELIGIBLE'`
3. Verify `is_eligible = true`

### Issue: "ALREADY_VOTED" response

**Cause:** Duplicate vote attempt

**Solution:** This is expected behavior. Voter can only vote once.

### Issue: "CHECKIN_EXPIRED" response

**Cause:** TPS check-in TTL exceeded (default 15 minutes)

**Solution:** Voter must check-in again at TPS

## Security Considerations

1. **Rate Limiting:** Implement per-voter rate limits on voting endpoints
2. **CSRF Protection:** Use CSRF tokens for state-changing operations
3. **Audit Logging:** Log all voting attempts (success & failure)
4. **Input Validation:** Strict validation of candidate_id
5. **Token Security:** Use secure JWT signing, short expiry times

## Performance Tips

1. **Connection Pooling:** Ensure database connection pool is properly configured
2. **Index Optimization:** Verify indexes on `(election_id, voter_id)` exist
3. **Cache Stats:** Consider caching live vote counts with Redis
4. **Async Audit:** Implement async audit logging to reduce latency
5. **Load Testing:** Test with concurrent requests to verify locking works

## Monitoring

### Metrics to Track

- Vote latency (p50, p95, p99)
- Error rate by error type
- Concurrent vote attempts
- Database lock wait time
- Transaction duration

### Logging

```go
// Add structured logging in handler
log.Info("vote cast attempt",
    "voter_id", voterID,
    "candidate_id", req.CandidateID,
    "method", "ONLINE",
)

// Log errors with context
log.Error("vote cast failed",
    "voter_id", voterID,
    "error", err,
    "error_type", reflect.TypeOf(err).String(),
)
```

## Next Steps

1. ✅ Implement handlers (DONE)
2. ✅ Add error mapping (DONE)
3. ✅ Add context helpers (DONE)
4. TODO: Wire handlers to router in main.go
5. TODO: Implement auth middleware
6. TODO: Add integration tests
7. TODO: Add monitoring & metrics
8. TODO: Load testing
