package monitoring

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

// GetVoteStats returns aggregated votes per candidate (online vs TPS).
func (r *PgRepository) GetVoteStats(ctx context.Context, electionID int64) ([]*VoteStats, error) {
	const q = `
SELECT
    election_id,
    candidate_id,
    COUNT(*) AS total_votes,
    SUM(CASE WHEN channel = 'ONLINE' THEN 1 ELSE 0 END) AS total_votes_online,
    SUM(CASE WHEN channel = 'TPS' THEN 1 ELSE 0 END) AS total_votes_tps,
    COALESCE(MAX(cast_at), NOW()) AS updated_at
FROM votes
WHERE election_id = $1
GROUP BY election_id, candidate_id
ORDER BY candidate_id
`
	rows, err := r.db.Query(ctx, q, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*VoteStats
	for rows.Next() {
		var s VoteStats
		if err := rows.Scan(
			&s.ElectionID,
			&s.CandidateID,
			&s.TotalVotes,
			&s.TotalVotesOnline,
			&s.TotalVotesTPS,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}
	return stats, rows.Err()
}

// GetParticipationStats returns eligible/voted counts and participation pct.
func (r *PgRepository) GetParticipationStats(ctx context.Context, electionID int64) (*ParticipationStats, error) {
	const q = `
SELECT
    election_id,
    COUNT(*) FILTER (WHERE is_eligible = TRUE) AS total_eligible,
    COUNT(*) FILTER (WHERE has_voted = TRUE)  AS total_voted
FROM voter_status
WHERE election_id = $1
GROUP BY election_id
`
	var s ParticipationStats
	err := r.db.QueryRow(ctx, q, electionID).Scan(
		&s.ElectionID,
		&s.TotalEligible,
		&s.TotalVoted,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No voter_status rows; return zeroed stats
			return &ParticipationStats{ElectionID: electionID}, nil
		}
		return nil, err
	}
	if s.TotalEligible > 0 {
		s.ParticipationPct = float64(s.TotalVoted) * 100.0 / float64(s.TotalEligible)
	}
	return &s, nil
}

// GetTPSStats aggregates votes and pending check-ins per TPS for an election.
func (r *PgRepository) GetTPSStats(ctx context.Context, electionID int64) ([]*TPSStats, error) {
	const q = `
SELECT
    t.id AS tps_id,
    t.name AS tps_name,
    COALESCE(v.total_votes, 0) AS total_votes,
    COALESCE(c.pending_checkins, 0) AS pending_checkins
FROM tps t
LEFT JOIN (
    SELECT tps_id, COUNT(*) AS total_votes
    FROM votes
    WHERE election_id = $1
    GROUP BY tps_id
) v ON v.tps_id = t.id
LEFT JOIN (
    SELECT tps_id, COUNT(*) AS pending_checkins
    FROM tps_checkins
    WHERE election_id = $1 AND status = 'PENDING'
    GROUP BY tps_id
) c ON c.tps_id = t.id
WHERE t.election_id = $1
ORDER BY t.id
`
	rows, err := r.db.Query(ctx, q, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*TPSStats
	for rows.Next() {
		var s TPSStats
		if err := rows.Scan(
			&s.TPSID,
			&s.TPSName,
			&s.TotalVotes,
			&s.PendingCheckins,
		); err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}
	return stats, rows.Err()
}

// GetLiveCount returns map[candidate_id]total_votes for quick lookups.
func (r *PgRepository) GetLiveCount(ctx context.Context, electionID int64) (map[int64]int64, error) {
	const q = `
SELECT candidate_id, COUNT(*) AS total_votes
FROM votes
WHERE election_id = $1
GROUP BY candidate_id
`
	rows, err := r.db.Query(ctx, q, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]int64)
	for rows.Next() {
		var candidateID, total int64
		if err := rows.Scan(&candidateID, &total); err != nil {
			return nil, err
		}
		result[candidateID] = total
	}
	return result, rows.Err()
}
