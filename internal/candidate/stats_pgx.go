package candidate

import (
"context"
_ "embed"

"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../queries/candidate_vote_stats.sql
var qCandidateStats string

// PgStatsProvider implements StatsProvider using PostgreSQL
type PgStatsProvider struct {
db *pgxpool.Pool
}

// NewPgStatsProvider creates a new PostgreSQL stats provider
func NewPgStatsProvider(db *pgxpool.Pool) *PgStatsProvider {
return &PgStatsProvider{db: db}
}

// GetCandidateStats returns voting statistics for all candidates in an election
func (p *PgStatsProvider) GetCandidateStats(
ctx context.Context,
electionID int64,
) (CandidateStatsMap, error) {
rows, err := p.db.Query(ctx, qCandidateStats, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

statsMap := make(CandidateStatsMap)

for rows.Next() {
var candidateID int64
var totalVotes int64
var percentage float64

if err := rows.Scan(&candidateID, &totalVotes, &percentage); err != nil {
return nil, err
}

statsMap[candidateID] = CandidateStats{
TotalVotes: totalVotes,
Percentage: percentage,
}
}

if err := rows.Err(); err != nil {
return nil, err
}

return statsMap, nil
}
