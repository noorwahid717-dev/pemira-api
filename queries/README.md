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

## 📖 Documentation

See `docs/PEMIRA_QUERIES.md` for detailed documentation and examples.
