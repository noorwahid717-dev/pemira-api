package voting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
	
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"pemira-api/internal/shared"
	"pemira-api/internal/shared/constants"
)

// CastVoteWithTransaction implements the critical voting path with proper locking
// This is the MISSION CRITICAL function that prevents double voting
func (s *Service) CastVoteWithTransaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	electionID, voterID, candidateID int64,
	votedVia string,
	tpsID *int64,
) (string, error) {
	// Start transaction with proper isolation level
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// Step 1: Get voter status with row-level lock (FOR UPDATE)
	// This prevents race conditions and double voting
	var voterStatus struct {
		ID         int64
		HasVoted   bool
		Status     constants.VoterStatus
		ElectionID int64
	}

	err = tx.QueryRow(ctx, `
		SELECT id, has_voted, status, election_id
		FROM voter_election_status
		WHERE voter_id = $1 AND election_id = $2
		FOR UPDATE
	`, voterID, electionID).Scan(
		&voterStatus.ID,
		&voterStatus.HasVoted,
		&voterStatus.Status,
		&voterStatus.ElectionID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", shared.ErrVoterNotEligible
		}
		return "", err
	}

	// Step 2: Validate voter eligibility
	if voterStatus.HasVoted {
		return "", shared.ErrAlreadyVoted
	}

	if voterStatus.Status != constants.VoterStatusEligible {
		return "", shared.ErrVoterNotEligible
	}

	// Step 3: Validate election phase
	var electionPhase string
	err = tx.QueryRow(ctx, `
		SELECT current_phase
		FROM elections
		WHERE id = $1 AND is_active = true
	`, electionID).Scan(&electionPhase)

	if err != nil {
		return "", shared.ErrInvalidPhase
	}

	if electionPhase != string(constants.PhaseVoting) {
		return "", shared.ErrInvalidPhase
	}

	// Step 4: Generate anonymous vote token
	token, tokenHash, err := s.generateVoteToken()
	if err != nil {
		return "", err
	}

	// Step 5: Insert vote token (for voter's receipt)
	_, err = tx.Exec(ctx, `
		INSERT INTO vote_tokens (voter_id, election_id, token, created_at)
		VALUES ($1, $2, $3, $4)
	`, voterID, electionID, tokenHash, time.Now())

	if err != nil {
		return "", err
	}

	// Step 6: Insert anonymous vote
	_, err = tx.Exec(ctx, `
		INSERT INTO votes (election_id, candidate_id, token_hash, voted_at, voted_via, tps_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, electionID, candidateID, tokenHash, time.Now(), votedVia, tpsID)

	if err != nil {
		return "", err
	}

	// Step 7: Update voter status (mark as voted)
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE voter_election_status
		SET has_voted = true,
		    voted_at = $1,
		    voted_via = $2,
		    tps_id = $3,
		    updated_at = $4
		WHERE id = $5
	`, now, votedVia, tpsID, now, voterStatus.ID)

	if err != nil {
		return "", err
	}

	// Step 8: Update vote statistics (materialized view/cache)
	_, err = tx.Exec(ctx, `
		INSERT INTO vote_stats (election_id, candidate_id, total_votes, updated_at)
		VALUES ($1, $2, 1, $3)
		ON CONFLICT (election_id, candidate_id)
		DO UPDATE SET
			total_votes = vote_stats.total_votes + 1,
			updated_at = $3
	`, electionID, candidateID, time.Now())

	if err != nil {
		return "", err
	}

	// Step 9: Insert audit log
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, entity_type, entity_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, voterID, "VOTE_CAST", "VOTE", candidateID, 
		map[string]interface{}{
			"election_id": electionID,
			"voted_via":   votedVia,
			"tps_id":      tpsID,
		}, time.Now())

	if err != nil {
		return "", err
	}

	// Step 10: Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	// Return the vote token as receipt
	return token, nil
}

func (s *Service) generateVoteToken() (token, hash string, err error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := s.generateToken(); err != nil {
		return "", "", err
	}

	token = hex.EncodeToString(tokenBytes)
	
	// Hash the token for storage
	hashBytes := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(hashBytes[:])
	
	return token, hash, nil
}

// ValidateVoteToken allows voter to verify their vote was recorded
func (s *Service) ValidateVoteToken(ctx context.Context, pool *pgxpool.Pool, token string) (bool, error) {
	// Hash the provided token
	hashBytes := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(hashBytes[:])

	// Check if token exists in votes table
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM votes WHERE token_hash = $1
		)
	`, hash).Scan(&exists)

	return exists, err
}
