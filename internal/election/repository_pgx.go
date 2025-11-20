package election

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

func NewRepository(db *pgxpool.Pool) Repository {
	return NewPgRepository(db)
}

var (
	ErrElectionNotFound    = fmt.Errorf("election not found")
	ErrVoterStatusNotFound = fmt.Errorf("voter status not found")
)

func (r *PgRepository) GetCurrentElection(ctx context.Context) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    code,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM elections
WHERE status = 'VOTING_OPEN'
ORDER BY voting_start_at NULLS LAST, id DESC
LIMIT 1
`
	var e Election
	err := r.db.QueryRow(ctx, q).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *PgRepository) GetByID(ctx context.Context, id int64) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    code,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM elections
WHERE id = $1
`
	var e Election
	err := r.db.QueryRow(ctx, q, id).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *PgRepository) GetVoterStatus(
	ctx context.Context,
	electionID, voterID int64,
) (*MeStatusRow, error) {
	const q = `
SELECT
    vs.election_id,
    vs.voter_id,
    vs.is_eligible,
    vs.has_voted,
    vs.voted_at,
    vs.voting_method,
    vs.tps_id,
    e.online_enabled,
    e.tps_enabled
FROM voter_status vs
JOIN elections e
  ON e.id = vs.election_id
WHERE vs.election_id = $1
  AND vs.voter_id = $2
`
	var row MeStatusRow
	var method *string

	err := r.db.QueryRow(ctx, q, electionID, voterID).Scan(
		&row.ElectionID,
		&row.VoterID,
		&row.IsEligible,
		&row.HasVoted,
		&row.LastVoteAt,
		&method,
		&row.LastTPSID,
		&row.OnlineEnabled,
		&row.TPSEnabled,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVoterStatusNotFound
		}
		return nil, err
	}

	row.LastVoteChannel = method
	return &row, nil
}
