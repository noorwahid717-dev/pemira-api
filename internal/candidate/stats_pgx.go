package candidate

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const qCandidateStats = `
SELECT 
    c.id AS candidate_id,
    COALESCE(COUNT(v.id), 0) AS total_votes,
    COALESCE(
        COUNT(v.id)::FLOAT / NULLIF((SELECT COUNT(*) FROM votes WHERE election_id = $1), 0) * 100,
        0
    ) AS percentage
FROM candidates c
LEFT JOIN votes v ON v.candidate_id = c.id AND v.election_id = $1
WHERE c.election_id = $1 AND c.status = 'APPROVED'
GROUP BY c.id
ORDER BY total_votes DESC, c.number ASC
`

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
