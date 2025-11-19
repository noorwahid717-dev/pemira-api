package voting

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TPSVotingService struct {
	db *pgxpool.Pool
}

func NewTPSVotingService(db *pgxpool.Pool) *TPSVotingService {
	return &TPSVotingService{
		db: db,
	}
}

func (s *TPSVotingService) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// CastTPSVote - Mahasiswa voting di TPS setelah check-in approved
func (s *TPSVotingService) CastTPSVote(
	ctx context.Context,
	voterID int64,
	electionID int64,
	candidateID int64,
) (*VoteReceipt, error) {
	var receipt *VoteReceipt

	err := s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Ambil latest tps_checkins dengan status APPROVED untuk (election_id, voter_id)
		checkin, err := s.getApprovedCheckin(ctx, tx, voterID, electionID)
		if err != nil {
			return err
		}

		// 2. Pastikan ExpiresAt > now()
		now := time.Now().UTC()
		if checkin.ExpiresAt == nil || checkin.ExpiresAt.Before(now) {
			return ErrCheckinExpired
		}

		// 3. Lock voter_status dengan FOR UPDATE untuk mencegah double voting
		voterStatus, err := s.getVoterStatusForUpdate(ctx, tx, electionID, voterID)
		if err != nil {
			return err
		}

		if voterStatus.HasVoted {
			return ErrAlreadyVoted
		}

		// 4. Validasi candidate exists dan eligible
		candidateValid, err := s.validateCandidate(ctx, tx, electionID, candidateID)
		if err != nil || !candidateValid {
			return ErrInvalidCandidate
		}

		// 5. Generate receipt token hash
		tokenHash, err := s.generateTokenHash()
		if err != nil {
			return err
		}

		// 6. Insert ke table votes
		voteID, err := s.insertVote(ctx, tx, electionID, candidateID, tokenHash, "TPS", now)
		if err != nil {
			return err
		}

		// 7. Update voter_status: has_voted = true, voted_at, tps_id
		err = s.updateVoterStatusAfterVote(ctx, tx, electionID, voterID, checkin.TPSID, now)
		if err != nil {
			return err
		}

		// 8. Update tps_checkins.Status = USED
		err = s.updateCheckinToUsed(ctx, tx, checkin.ID)
		if err != nil {
			return err
		}

		// 9. Audit log
		_ = s.logVoteAudit(ctx, tx, voterID, voteID, electionID, checkin.TPSID)

		// 10. Load TPS info
		tpsInfo, _ := s.getTPSInfo(ctx, tx, checkin.TPSID)

		// Build receipt
		receipt = &VoteReceipt{
			ElectionID: electionID,
			VoterID:    voterID,
			Method:     "TPS",
			VotedAt:    now,
			Receipt: ReceiptDetail{
				TokenHash: tokenHash,
				Note:      "Voting berhasil dilakukan di TPS",
			},
			TPS: tpsInfo,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetTPSVotingEligibility - Cek apakah voter eligible untuk voting TPS
func (s *TPSVotingService) GetTPSVotingEligibility(
	ctx context.Context,
	voterID int64,
	electionID int64,
) (*TPSVotingEligibility, error) {
	var eligibility *TPSVotingEligibility

	err := s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Cek voter status
		voterStatus, err := s.getVoterStatus(ctx, tx, electionID, voterID)
		if err != nil {
			return err
		}

		if voterStatus.HasVoted {
			eligibility = &TPSVotingEligibility{
				Eligible: false,
				Reason:   "Anda sudah voting",
			}
			return nil
		}

		// 2. Ambil latest approved checkin
		checkin, err := s.getApprovedCheckin(ctx, tx, voterID, electionID)
		if err != nil {
			eligibility = &TPSVotingEligibility{
				Eligible: false,
				Reason:   "Check-in belum disetujui panitia TPS",
			}
			return nil
		}

		// 3. Cek expires
		now := time.Now().UTC()
		if checkin.ExpiresAt == nil || checkin.ExpiresAt.Before(now) {
			eligibility = &TPSVotingEligibility{
				Eligible: false,
				Reason:   "Waktu voting sudah habis. Silakan check-in ulang.",
			}
			return nil
		}

		// 4. Load TPS info
		tps, _ := s.getTPSInfo(ctx, tx, checkin.TPSID)

		eligibility = &TPSVotingEligibility{
			Eligible:  true,
			Reason:    "Anda dapat melakukan voting",
			TPSID:     checkin.TPSID,
			TPSCode:   tps.Code,
			TPSName:   tps.Name,
			ExpiresAt: checkin.ExpiresAt,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return eligibility, nil
}

// ===== Repository methods with tx =====

type TPSCheckin struct {
	ID         int64
	TPSID      int64
	VoterID    int64
	ElectionID int64
	Status     string
	ExpiresAt  *time.Time
}

func (s *TPSVotingService) getApprovedCheckin(ctx context.Context, tx pgx.Tx, voterID, electionID int64) (*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, expires_at
		FROM tps_checkins
		WHERE voter_id = $1 AND election_id = $2 AND status = 'APPROVED'
		ORDER BY approved_at DESC
		LIMIT 1
	`

	var checkin TPSCheckin
	err := tx.QueryRow(ctx, query, voterID, electionID).Scan(
		&checkin.ID, &checkin.TPSID, &checkin.VoterID,
		&checkin.ElectionID, &checkin.Status, &checkin.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoApprovedCheckin
		}
		return nil, err
	}

	return &checkin, nil
}

type VoterStatus struct {
	HasVoted bool
	VotedAt  *time.Time
}

func (s *TPSVotingService) getVoterStatusForUpdate(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatus, error) {
	// Row-level lock dengan FOR UPDATE
	query := `
		SELECT has_voted, voted_at
		FROM voter_status
		WHERE election_id = $1 AND voter_id = $2
		FOR UPDATE
	`

	var status VoterStatus
	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(&status.HasVoted, &status.VotedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Jika belum ada row, insert dulu
			insertQuery := `
				INSERT INTO voter_status (election_id, voter_id, has_voted, created_at, updated_at)
				VALUES ($1, $2, false, NOW(), NOW())
				RETURNING has_voted, voted_at
			`
			err = tx.QueryRow(ctx, insertQuery, electionID, voterID).Scan(&status.HasVoted, &status.VotedAt)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &status, nil
}

func (s *TPSVotingService) getVoterStatus(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatus, error) {
	query := `
		SELECT has_voted, voted_at
		FROM voter_status
		WHERE election_id = $1 AND voter_id = $2
	`

	var status VoterStatus
	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(&status.HasVoted, &status.VotedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &VoterStatus{HasVoted: false}, nil
		}
		return nil, err
	}

	return &status, nil
}

func (s *TPSVotingService) validateCandidate(ctx context.Context, tx pgx.Tx, electionID, candidateID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM candidates
			WHERE id = $1 AND election_id = $2
		)
	`

	var exists bool
	err := tx.QueryRow(ctx, query, candidateID, electionID).Scan(&exists)
	return exists, err
}

func (s *TPSVotingService) generateTokenHash() (string, error) {
	// Generate simple hash - in production use proper crypto
	return "HASH-" + time.Now().Format("20060102150405"), nil
}

func (s *TPSVotingService) insertVote(ctx context.Context, tx pgx.Tx, electionID, candidateID int64, tokenHash, votedVia string, votedAt time.Time) (int64, error) {
	query := `
		INSERT INTO votes (election_id, candidate_id, token_hash, voted_via, voted_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(ctx, query,
		electionID, candidateID, tokenHash, votedVia, votedAt,
	).Scan(&id)

	return id, err
}

func (s *TPSVotingService) updateVoterStatusAfterVote(ctx context.Context, tx pgx.Tx, electionID, voterID, tpsID int64, votedAt time.Time) error {
	query := `
		UPDATE voter_status
		SET has_voted = true, voted_at = $1, tps_id = $2, updated_at = NOW()
		WHERE election_id = $3 AND voter_id = $4
	`

	_, err := tx.Exec(ctx, query, votedAt, tpsID, electionID, voterID)
	return err
}

func (s *TPSVotingService) updateCheckinToUsed(ctx context.Context, tx pgx.Tx, checkinID int64) error {
	query := `
		UPDATE tps_checkins
		SET status = 'USED', updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, checkinID)
	return err
}

func (s *TPSVotingService) getTPSInfo(ctx context.Context, tx pgx.Tx, tpsID int64) (*TPSInfo, error) {
	query := `SELECT id, code, name FROM tps WHERE id = $1`

	var info TPSInfo
	err := tx.QueryRow(ctx, query, tpsID).Scan(&info.ID, &info.Code, &info.Name)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *TPSVotingService) logVoteAudit(ctx context.Context, tx pgx.Tx, voterID, voteID, electionID, tpsID int64) error {
	// Simplified audit - adjust to your schema
	query := `
		INSERT INTO audit_logs (actor_voter_id, action, entity_type, entity_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	
	metadata := `{"election_id": ` + fmt.Sprintf("%d", electionID) + `, "tps_id": ` + fmt.Sprintf("%d", tpsID) + `}`
	_, _ = tx.Exec(ctx, query, voterID, "VOTE_CAST_TPS", "VOTE", voteID, metadata)
	
	return nil
}

// ===== DTOs =====

type TPSVotingEligibility struct {
	Eligible  bool       `json:"eligible"`
	Reason    string     `json:"reason"`
	TPSID     int64      `json:"tps_id,omitempty"`
	TPSCode   string     `json:"tps_code,omitempty"`
	TPSName   string     `json:"tps_name,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ===== Errors =====

var (
	ErrCheckinExpired    = VotingError{Code: "CHECKIN_EXPIRED", Message: "Waktu check-in sudah habis"}
	ErrAlreadyVoted      = VotingError{Code: "ALREADY_VOTED", Message: "Anda sudah melakukan voting"}
	ErrInvalidCandidate  = VotingError{Code: "INVALID_CANDIDATE", Message: "Kandidat tidak valid"}
	ErrNoApprovedCheckin = VotingError{Code: "NO_APPROVED_CHECKIN", Message: "Belum ada check-in yang disetujui"}
)

type VotingError struct {
	Code    string
	Message string
}

func (e VotingError) Error() string {
	return e.Message
}
