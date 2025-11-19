# Analytics Package

Package analytics provides data access and business logic for PEMIRA election analytics and reporting.

## Architecture

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access) → PostgreSQL
```

## Components

### Models (`models.go`)
Data structures for analytics results:
- `HourlyVotes` - Time-series votes with channel breakdown
- `HourlyCandidateVotes` - Votes per candidate over time
- `FacultyCandidateHeatmapRow` - Faculty preference matrix
- `TurnoutPoint` - Cumulative turnout progression
- `CohortCandidateVotes` - Demographic breakdown by cohort
- `PeakHour` - Busiest voting hours
- `VotingVelocity` - Statistical speed metrics

### Repository (`repository.go`)
Data access layer using pgxpool with embedded SQL queries:
- Uses `//go:embed` to load queries from `/queries/analytics_*.sql`
- Interface-based for testability
- Context-aware for cancellation

### Service (`service.go`)
Business logic layer:
- `GetDashboardCharts()` - Fetches all analytics in parallel using errgroup
- Individual methods wrap repository calls
- Add custom business logic here (e.g., caching, aggregations)

### Handler (`handler.go`)
HTTP/REST layer using Gin:
- RESTful endpoints under `/analytics/elections/:id/...`
- JSON responses
- Error handling

## Usage

### Initialization

```go
import (
    "github.com/pemira/internal/analytics"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Setup
pool, _ := pgxpool.New(context.Background(), connString)
repo := analytics.NewAnalyticsRepo(pool)
service := analytics.NewService(repo)
handler := analytics.NewHandler(service)

// Register routes
r := gin.Default()
api := r.Group("/api/v1")
handler.RegisterRoutes(api)
```

### API Endpoints

**Dashboard (all charts in one call):**
```
GET /api/v1/analytics/elections/{id}/dashboard
```

**Individual charts:**
```
GET /api/v1/analytics/elections/{id}/hourly-votes
GET /api/v1/analytics/elections/{id}/hourly-by-candidate
GET /api/v1/analytics/elections/{id}/faculty-heatmap
GET /api/v1/analytics/elections/{id}/turnout-timeline
GET /api/v1/analytics/elections/{id}/cohort-breakdown
GET /api/v1/analytics/elections/{id}/peak-hours
GET /api/v1/analytics/elections/{id}/voting-velocity
```

### Example Response

**GET /analytics/elections/1/hourly-votes**
```json
[
  {
    "bucket_start": "2025-01-15T08:00:00Z",
    "total_votes": 150,
    "votes_online": 100,
    "votes_tps": 50
  },
  {
    "bucket_start": "2025-01-15T09:00:00Z",
    "total_votes": 320,
    "votes_online": 200,
    "votes_tps": 120
  }
]
```

**GET /analytics/elections/1/dashboard**
```json
{
  "hourly_votes": [...],
  "hourly_by_candidate": [...],
  "faculty_heatmap": [...],
  "turnout_timeline": [...],
  "cohort_breakdown": [...],
  "peak_hours": [...],
  "voting_velocity": {
    "total_intervals": 1234,
    "avg_gap_minutes": 2.5,
    "median_gap_minutes": 1.8,
    "p95_gap_minutes": 8.5
  }
}
```

## Testing

### Repository Tests
```go
func TestAnalyticsRepo_GetHourlyVotesByChannel(t *testing.T) {
    pool := setupTestDB(t)
    repo := analytics.NewAnalyticsRepo(pool)
    
    result, err := repo.GetHourlyVotesByChannel(context.Background(), 1)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

### Service Tests (with mock)
```go
type mockRepo struct{}

func (m *mockRepo) GetHourlyVotesByChannel(ctx context.Context, id int64) ([]analytics.HourlyVotes, error) {
    return []analytics.HourlyVotes{{...}}, nil
}

func TestService_GetDashboardCharts(t *testing.T) {
    repo := &mockRepo{}
    service := analytics.NewService(repo)
    
    result, err := service.GetDashboardCharts(context.Background(), 1)
    
    assert.NoError(t, err)
    assert.NotNil(t, result.HourlyVotes)
}
```

## Performance

- **Dashboard endpoint**: Fetches 7 queries in parallel using errgroup
- **Caching**: Add Redis caching in service layer for frequently accessed data
- **Pagination**: Not needed for time-series (bounded by election duration)
- **Indexes**: All queries optimized with existing schema indexes

## SQL Queries

All SQL queries are embedded from `/queries/analytics_*.sql`:
- `analytics_02_timeline_votes_by_channel.sql`
- `analytics_03_timeline_votes_per_candidate.sql`
- `analytics_05_heatmap_faculty_candidate_percent.sql`
- `analytics_06_turnout_cumulative_timeline.sql`
- `analytics_07_votes_by_cohort_candidate.sql`
- `analytics_09_peak_hours_analysis.sql`
- `analytics_10_voting_velocity.sql`

See `/queries/README.md` for query documentation.

## Future Enhancements

- [ ] Add caching layer (Redis)
- [ ] Add real-time WebSocket updates
- [ ] Export to CSV/Excel
- [ ] Add date range filters
- [ ] Add comparison between elections
- [ ] Add predictive analytics
