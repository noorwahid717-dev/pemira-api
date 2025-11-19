# PEMIRA Query Collection

SQL query files untuk reporting & analytics PEMIRA UNIWA.

## 📁 File Structure

### Core Queries (01-11)
- `01_total_votes_per_candidate.sql` - Total suara per kandidat
- `02_votes_breakdown_by_channel.sql` - Breakdown ONLINE vs TPS
- `03_votes_per_candidate_per_tps.sql` - Distribusi suara per TPS
- `04_turnout_per_faculty.sql` - Turnout per fakultas
- `05_turnout_per_prodi.sql` - Turnout per program studi
- `06_turnout_overall.sql` - Turnout total election
- `07_participation_per_tps.sql` - Usage & capacity per TPS
- `08_tps_checkin_summary.sql` - Summary check-in status
- `09_tps_checkin_queue_pending.sql` - Antrian pending approval
- `10_votes_by_channel_summary.sql` - Global channel breakdown
- `11_dashboard_admin_summary.sql` - All-in-one dashboard query

### Audit & Utility
- `audit_duplicate_votes.sql` - Detect duplicate token usage
- `top_5_busiest_tps.sql` - Ranking TPS tersibuk
- `voters_not_voted_yet.sql` - List pemilih belum vote

### Analytics & Advanced Reporting
- `analytics_01_timeline_votes_per_hour.sql` - Time-series suara per jam (total)
- `analytics_02_timeline_votes_by_channel.sql` - Time-series per channel (ONLINE/TPS)
- `analytics_03_timeline_votes_per_candidate.sql` - Time-series per kandidat
- `analytics_04_heatmap_faculty_candidate.sql` - Heatmap fakultas × kandidat (count)
- `analytics_05_heatmap_faculty_candidate_percent.sql` - Heatmap dengan % per fakultas
- `analytics_06_turnout_cumulative_timeline.sql` - Cumulative turnout over time
- `analytics_07_votes_by_cohort_candidate.sql` - Breakdown per angkatan
- `analytics_08_votes_by_prodi_candidate.sql` - Breakdown per prodi (granular)
- `analytics_09_peak_hours_analysis.sql` - Analisis jam tersibuk
- `analytics_10_voting_velocity.sql` - Velocity metrics (avg gap antar vote)

## 🚀 Usage

### Command Line (psql)
```bash
# Run query with parameter
psql -h localhost -U pemira -d pemira_db \
  -f queries/01_total_votes_per_candidate.sql \
  -v election_id=1

# Export to CSV
psql -h localhost -U pemira -d pemira_db \
  -f queries/11_dashboard_admin_summary.sql \
  -v election_id=1 \
  --csv > dashboard.csv
```

### Go (database/sql)
```go
import (
    "database/sql"
    _ "embed"
)

//go:embed queries/01_total_votes_per_candidate.sql
var queryCandidateVotes string

func GetCandidateVotes(db *sql.DB, electionID int64) ([]Result, error) {
    rows, err := db.Query(queryCandidateVotes, electionID)
    // ... process rows
}
```

### Go (sqlx)
```go
import (
    "github.com/jmoiron/sqlx"
    _ "embed"
)

//go:embed queries/06_turnout_overall.sql
var queryTurnoutOverall string

type TurnoutResult struct {
    TotalEligible   int64   `db:"total_eligible"`
    TotalVoted      int64   `db:"total_voted"`
    TurnoutPercent  float64 `db:"turnout_percent"`
}

func GetTurnout(db *sqlx.DB, electionID int64) (*TurnoutResult, error) {
    var result TurnoutResult
    err := db.Get(&result, queryTurnoutOverall, electionID)
    return &result, err
}
```

## 📊 Query Parameters

Most queries use `$1` for `election_id`, except:
- `09_tps_checkin_queue_pending.sql` - uses `$1` for `tps_id`

## ⚡ Performance Tips

1. **Indexes**: All queries optimized with schema indexes
2. **Caching**: Cache dashboard query (11) for 1-5 minutes
3. **Pagination**: Add `LIMIT` and `OFFSET` for large results
4. **Connection Pool**: Use connection pooling for concurrent queries

## 📈 Visualization Use Cases

### Timeline Charts (Line/Bar)
- `analytics_01`, `analytics_02`, `analytics_03` - Suara over time
- `analytics_06` - Cumulative turnout progression

### Heatmaps
- `analytics_04`, `analytics_05` - Faculty/Prodi × Candidate matrix
- `analytics_08` - Granular prodi breakdown

### Clustered/Grouped Charts
- `analytics_02` - ONLINE vs TPS stacked bar
- `analytics_07` - Votes by cohort year
- `analytics_09` - Peak hours ranking

### Statistical Analysis
- `analytics_10` - Velocity & gap analysis for capacity planning

## 📖 Documentation

See `docs/PEMIRA_QUERIES.md` for detailed documentation and examples.
