package voting

import (
	"context"
	"fmt"
	"time"
	
	"github.com/jackc/pgx/v5"
	"pemira-api/internal/shared"
	"pemira-api/internal/tps"
)

type voteRepository struct{}

func NewVoteRepository() VoteRepository {
	return &voteRepository{}
}

func (r *voteRepository) InsertToken(ctx context.Context, tx pgx.Tx, token *VoteToken) error {
	query := `
		INSERT INTO vote_tokens (election_id, voter_id, token, issued_at, method, tps_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	err := tx.QueryRow(ctx, query,
		token.ElectionID,
		token.VoterID,
		token.TokenHash,
		token.IssuedAt,
		token.Method,
		token.TPSID,
	).Scan(&token.ID)
	
	if err != nil {
		return fmt.Errorf("insert vote token: %w", err)
	}
	
	return nil
}

func (r *voteRepository) InsertVote(ctx context.Context, tx pgx.Tx, vote *Vote) error {
	query := `
		INSERT INTO votes (election_id, candidate_id, token_hash, channel, tps_id, cast_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	err := tx.QueryRow(ctx, query,
		vote.ElectionID,
		vote.CandidateID,
		vote.TokenHash,
		vote.Channel,
		vote.TPSID,
		vote.CastAt,
	).Scan(&vote.ID)
	
	if err != nil {
		return fmt.Errorf("insert vote: %w", err)
	}
	
	return nil
}

func (r *voteRepository) MarkTokenUsed(ctx context.Context, tx pgx.Tx, electionID int64, tokenHash string, usedAt time.Time) error {
	query := `
		UPDATE vote_tokens
		SET used_at = $1
		WHERE election_id = $2 AND token = $3
	`
	
	_, err := tx.Exec(ctx, query, usedAt, electionID, tokenHash)
	if err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}
	
	return nil
}

func (r *voteRepository) GetLatestApprovedCheckin(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*tps.TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at, 
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE election_id = $1 AND voter_id = $2 AND status = $3
		ORDER BY approved_at DESC
		LIMIT 1
	`
	
	var checkin tps.TPSCheckin
	
	err := tx.QueryRow(ctx, query, electionID, voterID, tps.CheckinStatusApproved).Scan(
		&checkin.ID,
		&checkin.TPSID,
		&checkin.VoterID,
		&checkin.ElectionID,
		&checkin.Status,
		&checkin.ScanAt,
		&checkin.ApprovedAt,
		&checkin.ApprovedByID,
		&checkin.RejectionReason,
		&checkin.ExpiresAt,
		&checkin.CreatedAt,
		&checkin.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get latest approved checkin: %w", err)
	}
	
	return &checkin, nil
}

func (r *voteRepository) GetTPSByID(ctx context.Context, tx pgx.Tx, tpsID int64) (*tps.TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, 
		       voting_date, open_time, close_time, capacity_estimate, 
		       area_faculty_id, created_at, updated_at
		FROM tps
		WHERE id = $1
	`
	
	var t tps.TPS
	
	err := tx.QueryRow(ctx, query, tpsID).Scan(
		&t.ID,
		&t.ElectionID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Status,
		&t.VotingDate,
		&t.OpenTime,
		&t.CloseTime,
		&t.CapacityEstimate,
		&t.AreaFacultyID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get TPS: %w", err)
	}
	
	return &t, nil
}

func (r *voteRepository) MarkCheckinUsed(ctx context.Context, tx pgx.Tx, checkinID int64, usedAt time.Time) error {
	query := `
		UPDATE tps_checkins
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := tx.Exec(ctx, query, tps.CheckinStatusUsed, usedAt, checkinID)
	if err != nil {
		return fmt.Errorf("mark checkin used: %w", err)
	}
	
	return nil
}
