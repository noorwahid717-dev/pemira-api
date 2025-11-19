# Candidate Package Integration Guide

Complete guide untuk wire candidate package ke aplikasi.

## 🔧 Dependencies Setup

### 1. Initialize Database Connection

```go
package main

import (
    "context"
    "log"
    
    "github.com/jackc/pgx/v5/pgxpool"
)

func setupDatabase(connString string) *pgxpool.Pool {
    pool, err := pgxpool.New(context.Background(), connString)
    if err != nil {
        log.Fatal("Unable to connect to database:", err)
    }
    
    // Test connection
    if err := pool.Ping(context.Background()); err != nil {
        log.Fatal("Unable to ping database:", err)
    }
    
    log.Println("Database connected successfully")
    return pool
}
```

### 2. Initialize Candidate Package

```go
import (
    "pemira-api/internal/candidate"
)

func setupCandidateService(pool *pgxpool.Pool) *candidate.Service {
    // Setup repository
    candidateRepo := candidate.NewPgCandidateRepository(pool)
    
    // Setup stats provider
    statsProvider := candidate.NewPgStatsProvider(pool)
    
    // Create service
    service := candidate.NewService(candidateRepo, statsProvider)
    
    return service
}
```

### 3. Complete Initialization

```go
func main() {
    // Load config
    connString := os.Getenv("DATABASE_URL")
    
    // Setup database
    pool := setupDatabase(connString)
    defer pool.Close()
    
    // Setup candidate service
    candidateService := setupCandidateService(pool)
    
    // Setup HTTP server
    router := setupRouter(candidateService)
    
    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", router)
}
```

## 🌐 HTTP Handler Implementation

### 1. Create Handler Struct

```go
package handler

import (
    "context"
    "errors"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
    "pemira-api/internal/candidate"
    "pemira-api/internal/http/response"
)

type CandidateHandler struct {
    service *candidate.Service
}

func NewCandidateHandler(service *candidate.Service) *CandidateHandler {
    return &CandidateHandler{service: service}
}
```

### 2. Implement List Handler

```go
func (h *CandidateHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse election ID from URL
    electionIDStr := chi.URLParam(r, "electionID")
    electionID, err := strconv.ParseInt(electionIDStr, 10, 64)
    if err != nil {
        response.BadRequest(w, "Invalid election ID", nil)
        return
    }
    
    // Parse query parameters
    search := r.URL.Query().Get("search")
    
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page <= 0 {
        page = 1
    }
    
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit <= 0 {
        limit = 10
    }
    
    // Call service
    items, pagination, err := h.service.ListPublicCandidates(
        ctx,
        electionID,
        search,
        page,
        limit,
    )
    
    if err != nil {
        response.InternalServerError(w, "Failed to fetch candidates")
        return
    }
    
    // Send response
    response.Success(w, http.StatusOK, map[string]any{
        "items":      items,
        "pagination": pagination,
    })
}
```

### 3. Implement Detail Handler

```go
func (h *CandidateHandler) GetCandidateDetail(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse IDs
    electionIDStr := chi.URLParam(r, "electionID")
    candidateIDStr := chi.URLParam(r, "candidateID")
    
    electionID, err := strconv.ParseInt(electionIDStr, 10, 64)
    if err != nil {
        response.BadRequest(w, "Invalid election ID", nil)
        return
    }
    
    candidateID, err := strconv.ParseInt(candidateIDStr, 10, 64)
    if err != nil {
        response.BadRequest(w, "Invalid candidate ID", nil)
        return
    }
    
    // Call service
    detail, err := h.service.GetPublicCandidateDetail(ctx, electionID, candidateID)
    
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    // Send response
    response.Success(w, http.StatusOK, detail)
}
```

### 4. Error Handling

```go
func (h *CandidateHandler) handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, candidate.ErrCandidateNotFound):
        response.NotFound(w, "Kandidat tidak ditemukan")
        
    case errors.Is(err, candidate.ErrCandidateNotPublished):
        response.NotFound(w, "Kandidat tidak ditemukan")
        
    default:
        response.InternalServerError(w, "Terjadi kesalahan pada sistem")
    }
}
```

## 🛣️ Router Setup

### Complete Router with Authentication

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "pemira-api/internal/candidate"
    candidateHandler "pemira-api/internal/handler"
    authMiddleware "pemira-api/internal/middleware"
)

func setupRouter(candidateService *candidate.Service) *chi.Mux {
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RealIP)
    r.Use(middleware.RequestID)
    
    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    // Public API routes (student)
    r.Route("/elections/{electionID}", func(er chi.Router) {
        // Auth middleware for students
        er.Use(authMiddleware.AuthRequired)
        er.Use(authMiddleware.StudentOnly)
        
        // Initialize handler
        handler := candidateHandler.NewCandidateHandler(candidateService)
        
        // Candidate endpoints
        er.Get("/candidates", handler.ListCandidates)
        er.Get("/candidates/{candidateID}", handler.GetCandidateDetail)
    })
    
    return r
}
```

## 🧪 Testing Examples

### 1. Repository Test

```go
package candidate_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "pemira-api/internal/candidate"
)

func TestPgCandidateRepository_GetByID(t *testing.T) {
    // Setup test database
    pool := setupTestDB(t)
    defer pool.Close()
    
    // Seed test data
    electionID := seedElection(t, pool)
    candidateID := seedCandidate(t, pool, electionID)
    
    // Create repository
    repo := candidate.NewPgCandidateRepository(pool)
    
    // Test GetByID
    c, err := repo.GetByID(context.Background(), electionID, candidateID)
    
    require.NoError(t, err)
    assert.NotNil(t, c)
    assert.Equal(t, candidateID, c.ID)
    assert.Equal(t, electionID, c.ElectionID)
}

func TestPgCandidateRepository_ListByElection(t *testing.T) {
    pool := setupTestDB(t)
    defer pool.Close()
    
    electionID := seedElection(t, pool)
    seedCandidate(t, pool, electionID)
    seedCandidate(t, pool, electionID)
    
    repo := candidate.NewPgCandidateRepository(pool)
    
    // Test list with published filter
    status := candidate.CandidateStatusPublished
    filter := candidate.Filter{
        Status: &status,
        Limit:  10,
    }
    
    candidates, total, err := repo.ListByElection(context.Background(), electionID, filter)
    
    require.NoError(t, err)
    assert.Greater(t, total, int64(0))
    assert.NotEmpty(t, candidates)
}
```

### 2. Stats Provider Test

```go
func TestPgStatsProvider_GetCandidateStats(t *testing.T) {
    pool := setupTestDB(t)
    defer pool.Close()
    
    electionID := seedElection(t, pool)
    candidateID1 := seedCandidate(t, pool, electionID)
    candidateID2 := seedCandidate(t, pool, electionID)
    
    // Seed votes
    seedVotes(t, pool, electionID, candidateID1, 60)
    seedVotes(t, pool, electionID, candidateID2, 40)
    
    // Test stats provider
    provider := candidate.NewPgStatsProvider(pool)
    stats, err := provider.GetCandidateStats(context.Background(), electionID)
    
    require.NoError(t, err)
    assert.Len(t, stats, 2)
    
    // Verify stats
    assert.Equal(t, int64(60), stats[candidateID1].TotalVotes)
    assert.Equal(t, 60.0, stats[candidateID1].Percentage)
    
    assert.Equal(t, int64(40), stats[candidateID2].TotalVotes)
    assert.Equal(t, 40.0, stats[candidateID2].Percentage)
}
```

### 3. Service Test with Mocks

```go
type mockRepo struct{}
type mockStats struct{}

func (m *mockRepo) ListByElection(ctx context.Context, id int64, filter candidate.Filter) ([]candidate.Candidate, int64, error) {
    return []candidate.Candidate{
        {ID: 1, ElectionID: id, Number: 1, Name: "Test", Status: candidate.CandidateStatusPublished},
    }, 1, nil
}

func (m *mockRepo) GetByID(ctx context.Context, electionID, candidateID int64) (*candidate.Candidate, error) {
    return &candidate.Candidate{
        ID: candidateID, 
        ElectionID: electionID,
        Status: candidate.CandidateStatusPublished,
    }, nil
}

func (m *mockStats) GetCandidateStats(ctx context.Context, id int64) (candidate.CandidateStatsMap, error) {
    return candidate.CandidateStatsMap{
        1: {TotalVotes: 100, Percentage: 50.0},
    }, nil
}

func TestService_ListPublicCandidates(t *testing.T) {
    repo := &mockRepo{}
    stats := &mockStats{}
    service := candidate.NewService(repo, stats)
    
    items, pag, err := service.ListPublicCandidates(
        context.Background(),
        1,
        "",
        1,
        10,
    )
    
    require.NoError(t, err)
    assert.Len(t, items, 1)
    assert.Equal(t, int64(100), items[0].Stats.TotalVotes)
    assert.Equal(t, 50.0, items[0].Stats.Percentage)
    assert.Equal(t, int64(1), pag.TotalItems)
}
```

## 📝 Example cURL Requests

### List Candidates

```bash
curl -X GET "http://localhost:8080/elections/1/candidates?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### Get Candidate Detail

```bash
curl -X GET "http://localhost:8080/elections/1/candidates/1" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### Search Candidates

```bash
curl -X GET "http://localhost:8080/elections/1/candidates?search=paslon&page=1&limit=5" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

## 🚀 Production Checklist

- [ ] Database migrations run successfully
- [ ] Indexes created on candidates table
- [ ] Connection pool configured (min/max connections)
- [ ] Authentication middleware integrated
- [ ] CORS configured for frontend
- [ ] Rate limiting added
- [ ] Logging configured
- [ ] Monitoring/metrics setup
- [ ] Error tracking (Sentry/similar)
- [ ] Load testing completed
- [ ] API documentation published

## 🔍 Troubleshooting

### "candidate not found" errors

```go
// Check if candidate exists in database
// Check if candidate status is PUBLISHED
// Verify electionID matches
```

### Stats not showing

```go
// Verify votes table has data
// Check if votes.election_id matches
// Run query manually to debug
```

### Slow queries

```go
// Add indexes:
CREATE INDEX idx_candidates_election_status ON candidates(election_id, status);
CREATE INDEX idx_votes_election_candidate ON votes(election_id, candidate_id);
```

## 📚 See Also

- `/docs/CANDIDATE_API.md` - API specification
- `/internal/candidate/README.md` - Package documentation
- `/queries/candidate_vote_stats.sql` - Stats SQL query
