# Candidate Package

Package candidate provides domain models, repository, and service for candidate management in PEMIRA UNIWA.

## Architecture

```
Handler (HTTP) → Service → Repository → PostgreSQL
                    ↓
              StatsProvider (Analytics)
```

## Components

### Model (`model.go`)

Domain entities:
- `Candidate` - Main candidate entity
- `MainProgram` - Program details (stored as JSONB)
- `Media` - Media assets (stored as JSONB)
- `SocialLink` - Social media links (stored as JSONB)
- `CandidateStats` - Voting statistics

### Repository (`repository.go`, `repository_pgx.go`)

Data access layer:
- `CandidateRepository` interface
- `PgCandidateRepository` implementation with pgxpool
- Handles JSONB scanning for nested structs

**Methods:**
- `ListByElection(ctx, electionID, filter)` - List with filters & pagination
- `GetByID(ctx, electionID, candidateID)` - Get single candidate

### Service (`service.go`)

Business logic layer:
- Combines candidate data with voting statistics
- Filters published candidates for students
- Returns DTOs ready for API responses

**Methods:**
- `ListPublicCandidates(ctx, electionID, search, page, limit)` - For student API
- `GetPublicCandidateDetail(ctx, electionID, candidateID)` - For student API

**DTOs:**
- `CandidateListItemDTO` - List view
- `CandidateDetailDTO` - Detail view
- `Pagination` - Pagination metadata

### Stats Adapter (`stats_adapter.go`)

Integrates with analytics package:
- `AnalyticsStatsAdapter` - Adapts analytics to StatsProvider
- Decouples candidate service from analytics implementation

## Database Schema

```sql
CREATE TABLE candidates (
    id                  BIGSERIAL PRIMARY KEY,
    election_id         BIGINT NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    number              INTEGER NOT NULL,
    name                TEXT NOT NULL,
    photo_url           TEXT NOT NULL,
    short_bio           TEXT NOT NULL,
    long_bio            TEXT,
    tagline             TEXT NOT NULL,
    faculty_name        TEXT,
    study_program_name  TEXT,
    cohort_year         INTEGER,
    vision              TEXT,
    missions            JSONB NOT NULL DEFAULT '[]',
    main_programs       JSONB NOT NULL DEFAULT '[]',
    media               JSONB NOT NULL DEFAULT '{}'::jsonb,
    social_links        JSONB NOT NULL DEFAULT '[]',
    status              TEXT NOT NULL DEFAULT 'DRAFT',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT ux_candidates_election_number UNIQUE (election_id, number)
);

CREATE INDEX idx_candidates_election ON candidates(election_id);
CREATE INDEX idx_candidates_status ON candidates(status);
```

## Usage

### Setup

**Option 1: Using PgStatsProvider (Recommended)**

```go
import (
    "github.com/pemira/internal/candidate"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Initialize dependencies
pool, _ := pgxpool.New(ctx, connString)

// Setup repositories
candidateRepo := candidate.NewPgCandidateRepository(pool)
statsProvider := candidate.NewPgStatsProvider(pool) // Direct implementation

// Setup service
candidateService := candidate.NewService(candidateRepo, statsProvider)
```

**Option 2: Using Analytics Package Adapter**

```go
import (
    "github.com/pemira/internal/candidate"
    "github.com/pemira/internal/analytics"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Initialize dependencies
pool, _ := pgxpool.New(ctx, connString)

// Setup repositories
candidateRepo := candidate.NewPgCandidateRepository(pool)
analyticsRepo := analytics.NewAnalyticsRepo(pool)

// Setup stats adapter (if analytics has GetCandidateVoteStats method)
statsAdapter := candidate.NewAnalyticsStatsAdapter(analyticsRepo)

// Setup service
candidateService := candidate.NewService(candidateRepo, statsAdapter)
```

### List Candidates (Student View)

```go
candidates, pagination, err := candidateService.ListPublicCandidates(
    ctx,
    electionID,
    "", // search query
    1,  // page
    10, // limit
)

if err != nil {
    // handle error
}

// candidates contains []CandidateListItemDTO with stats
// pagination contains page info
```

### Get Candidate Detail (Student View)

```go
detail, err := candidateService.GetPublicCandidateDetail(
    ctx,
    electionID,
    candidateID,
)

if errors.Is(err, candidate.ErrCandidateNotFound) {
    // 404 Not Found
}

if errors.Is(err, candidate.ErrCandidateNotPublished) {
    // 404 Not Found (for students)
}

// detail contains full candidate info with stats
```

## HTTP Handler Integration

See `docs/CANDIDATE_API.md` for complete API specification.

**Public endpoints:**
```
GET /elections/{election_id}/candidates
GET /elections/{election_id}/candidates/{candidate_id}
```

**Handler example:**
```go
func (h *Handler) ListCandidates(w http.ResponseWriter, r *http.Request) {
    electionID := chi.URLParam(r, "electionID")
    
    // Parse query params
    search := r.URL.Query().Get("search")
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    
    // Call service
    items, pag, err := h.service.ListPublicCandidates(
        r.Context(),
        electionID,
        search,
        page,
        limit,
    )
    
    if err != nil {
        // handle error
        return
    }
    
    // Send response
    response.Success(w, http.StatusOK, map[string]any{
        "items": items,
        "pagination": pag,
    })
}
```

## JSONB Fields

### Missions ([]string)
```json
[
  "Meningkatkan transparansi anggaran dan program kerja BEM.",
  "Membangun ekosistem kolaborasi antar UKM dan komunitas."
]
```

### Main Programs ([]MainProgram)
```json
[
  {
    "title": "UNIWA Aspiration Hub",
    "description": "Platform digital untuk aspirasi mahasiswa.",
    "category": "Transparansi & Aspirasi"
  }
]
```

### Media (Media)
```json
{
  "video_url": "https://www.youtube.com/watch?v=abc123",
  "gallery_photos": [
    "https://cdn.pemira.uniwa.ac.id/candidates/1/gallery/1.jpg"
  ],
  "document_manifesto_url": "https://cdn.pemira.uniwa.ac.id/candidates/1/visi-misi.pdf"
}
```

### Social Links ([]SocialLink)
```json
[
  {
    "platform": "instagram",
    "url": "https://instagram.com/paslon_a"
  },
  {
    "platform": "tiktok",
    "url": "https://tiktok.com/@paslon_a"
  }
]
```

## Error Handling

```go
err := service.GetPublicCandidateDetail(ctx, electionID, candidateID)

switch {
case errors.Is(err, candidate.ErrCandidateNotFound):
    // HTTP 404: Candidate not found
    
case errors.Is(err, candidate.ErrCandidateNotPublished):
    // HTTP 404: Candidate exists but not published (hide from students)
    
default:
    // HTTP 500: Internal server error
}
```

## Testing

### Repository Tests
```go
func TestPgCandidateRepository_ListByElection(t *testing.T) {
    pool := setupTestDB(t)
    repo := candidate.NewPgCandidateRepository(pool)
    
    // Insert test data
    seedCandidates(t, pool)
    
    // Test list
    candidates, total, err := repo.ListByElection(context.Background(), 1, candidate.Filter{
        Status: ptrStatus(candidate.CandidateStatusPublished),
        Limit: 10,
    })
    
    assert.NoError(t, err)
    assert.Greater(t, total, int64(0))
    assert.NotEmpty(t, candidates)
}
```

### Service Tests (with mock)
```go
type mockRepo struct{}
type mockStats struct{}

func (m *mockRepo) ListByElection(ctx, id, filter) ([]candidate.Candidate, int64, error) {
    return []candidate.Candidate{{...}}, 1, nil
}

func (m *mockStats) GetCandidateStats(ctx, id) (candidate.CandidateStatsMap, error) {
    return candidate.CandidateStatsMap{
        1: {TotalVotes: 100, Percentage: 50.0},
    }, nil
}

func TestService_ListPublicCandidates(t *testing.T) {
    repo := &mockRepo{}
    stats := &mockStats{}
    service := candidate.NewService(repo, stats)
    
    items, pag, err := service.ListPublicCandidates(context.Background(), 1, "", 1, 10)
    
    assert.NoError(t, err)
    assert.Len(t, items, 1)
    assert.Equal(t, int64(50.0), items[0].Stats.Percentage)
}
```

## Stats Provider Implementation

### SQL Query

See `queries/candidate_vote_stats.sql`:

```sql
WITH total_election AS (
    SELECT COUNT(*)::NUMERIC AS total_votes
    FROM votes WHERE election_id = $1
)
SELECT
    v.candidate_id,
    COUNT(*) AS candidate_votes,
    CASE
        WHEN te.total_votes = 0 THEN 0
        ELSE ROUND(COUNT(*)::NUMERIC / te.total_votes * 100, 2)
    END AS percentage
FROM votes v
CROSS JOIN total_election te
WHERE v.election_id = $1
GROUP BY v.candidate_id, te.total_votes;
```

**Features:**
- Calculates total votes in election
- Computes votes per candidate
- Calculates percentage with 2 decimal precision
- Handles zero votes case
- Only returns candidates with votes (map lookup handles missing entries)

### PgStatsProvider

Direct implementation using pgxpool:

```go
statsProvider := candidate.NewPgStatsProvider(pool)
stats, err := statsProvider.GetCandidateStats(ctx, electionID)
// stats is map[int64]CandidateStats
// stats[candidateID] = {TotalVotes: 100, Percentage: 45.67}
```

**Behavior:**
- Only candidates with votes appear in map
- Candidates without votes get default {0, 0.0} from service layer
- Uses go:embed for SQL query

## Future Enhancements

- [ ] Add admin CRUD methods (Create, Update, Delete)
- [ ] Add publish/unpublish methods
- [ ] Add image upload handling
- [ ] Add validation layer
- [ ] Add caching for frequently accessed candidates
- [ ] Add full-text search
- [ ] Add audit logging
- [ ] Add real-time stats updates (WebSocket)

## See Also

- `/docs/CANDIDATE_API.md` - Complete API specification
- `/internal/analytics` - Analytics package for voting statistics
- `/queries/01_total_votes_per_candidate.sql` - Stats query
