package voting

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pemira-api/internal/election"
	"pemira-api/internal/shared"
	"pemira-api/internal/tps"
)

type Service struct {
	db            *pgxpool.Pool
	repo          Repository
	electionRepo  election.Repository
	voterRepo     VoterRepository
	candidateRepo CandidateRepository
	voteRepo      VoteRepository
	statsRepo     VoteStatsRepository
	auditSvc      AuditService
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewVotingService(
	db *pgxpool.Pool,
	electionRepo election.Repository,
	voterRepo VoterRepository,
	candidateRepo CandidateRepository,
	voteRepo VoteRepository,
	statsRepo VoteStatsRepository,
	auditSvc AuditService,
) *Service {
	return &Service{
		db:            db,
		electionRepo:  electionRepo,
		voterRepo:     voterRepo,
		candidateRepo: candidateRepo,
		voteRepo:      voteRepo,
		statsRepo:     statsRepo,
		auditSvc:      auditSvc,
	}
}

// withTx executes a function within a transaction
func (s *Service) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
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

// GetVotingConfig returns voting configuration and voter eligibility
func (s *Service) GetVotingConfig(ctx context.Context, voterID int64) (*VotingConfigResponse, error) {
	// TODO: Implement with actual database queries
	// This is a stub implementation
	return &VotingConfigResponse{}, nil
}

// CastOnlineVote handles online voting with full validation and transaction
func (s *Service) CastOnlineVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error) {
	if s.db == nil || s.electionRepo == nil {
		return nil, errors.New("not implemented")
	}

	// 1. Get current election
	election, err := s.electionRepo.GetCurrentElection(ctx)
	if err != nil {
		return nil, translateNotFound(err, ErrElectionNotFound)
	}

	// 2. Validate election status
	// TODO: Check current_phase from election table or phase schedule
	// For now, assume election has a current_phase field or we check is_active
	if election.Status != election.ElectionStatusVotingOpen {
		return nil, ErrElectionNotOpen
	}

	// 3. Validate online mode enabled
	if !election.OnlineEnabled {
		return nil, ErrMethodNotAllowed
	}

	// 4. Cast vote with transaction
	result, err := s.castVote(ctx, election.ID, voterID, candidateID, "ONLINE", nil)
	if err != nil {
		return nil, err
	}

	// 5. Convert to DTO
	return &VoteReceipt{
		ElectionID: result.ElectionID,
		VoterID:    result.VoterID,
		Method:     result.Method,
		VotedAt:    result.VotedAt,
		Receipt: ReceiptDetail{
			TokenHash: result.Receipt.TokenHash,
			Note:      "Your vote has been recorded securely",
		},
		TPS: result.TPS,
	}, nil
}

// CastTPSVote handles TPS voting after check-in approval
func (s *Service) CastTPSVote(ctx context.Context, voterID, candidateID int64) (*VoteReceipt, error) {
	if s.db == nil || s.electionRepo == nil {
		return nil, errors.New("not implemented")
	}

	// 1. Get current election
	election, err := s.electionRepo.GetCurrentElection(ctx)
	if err != nil {
		return nil, translateNotFound(err, ErrElectionNotFound)
	}

	// 2. Validate election status
	if election.Status != election.ElectionStatusVotingOpen {
		return nil, ErrElectionNotOpen
	}

	// 3. Validate TPS mode enabled
	if !election.TPSEnabled {
		return nil, ErrMethodNotAllowed
	}

	// 4. Get latest approved check-in (must be done in transaction for consistency)
	var checkin *tps.TPSCheckin
	var tpsEntry *tps.TPS

	err = s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		checkin, err = s.voteRepo.GetLatestApprovedCheckin(ctx, tx, election.ID, voterID)
		if err != nil {
			return translateNotFound(err, ErrTPSCheckinNotFound)
		}

		// Validate check-in status
		if checkin.Status != tps.CheckinStatusApproved {
			return ErrTPSCheckinNotApproved
		}

		// Validate not expired (15 minutes TTL)
		if checkin.ExpiresAt != nil && checkin.ExpiresAt.Before(time.Now().UTC()) {
			return ErrCheckinExpired
		}

		// Get TPS info
		tpsEntry, err = s.voteRepo.GetTPSByID(ctx, tx, checkin.TPSID)
		if err != nil {
			return translateNotFound(err, ErrTPSNotFound)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Cast vote with TPS info
	result, err := s.castVote(ctx, election.ID, voterID, candidateID, "TPS", &tpsEntry.ID)
	if err != nil {
		return nil, err
	}

	// 6. Mark check-in as used
	_ = s.withTx(ctx, func(tx pgx.Tx) error {
		return s.voteRepo.MarkCheckinUsed(ctx, tx, checkin.ID, time.Now().UTC())
	})

	// 7. Convert to DTO
	return &VoteReceipt{
		ElectionID: result.ElectionID,
		VoterID:    result.VoterID,
		Method:     result.Method,
		VotedAt:    result.VotedAt,
		Receipt: ReceiptDetail{
			TokenHash: result.Receipt.TokenHash,
			Note:      "Your vote has been recorded securely at TPS",
		},
		TPS: result.TPS,
	}, nil
}

// castVote is the core voting logic with transaction safety
func (s *Service) castVote(
	ctx context.Context,
	electionID, voterID, candidateID int64,
	channel string,
	tpsID *int64,
) (*VoteResultEntity, error) {
	var result *VoteResultEntity

	err := s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Lock voter_status with FOR UPDATE
		vs, err := s.voterRepo.GetStatusForUpdate(ctx, tx, electionID, voterID)
		if err != nil {
			return translateNotFound(err, ErrNotEligible)
		}

		// 2. Check eligibility
		if !vs.IsEligible {
			return ErrNotEligible
		}
		if vs.HasVoted {
			return ErrAlreadyVoted
		}

		// 3. Get and validate candidate
		cand, err := s.candidateRepo.GetByIDWithTx(ctx, tx, candidateID)
		if err != nil {
			return translateNotFound(err, ErrCandidateNotFound)
		}
		if cand.ElectionID != electionID {
			return ErrCandidateNotFound
		}
		if !cand.IsActive {
			return ErrCandidateInactive
		}

		// 4. Generate token hash
		now := time.Now().UTC()
		tokenHash := generateTokenHash(electionID, voterID)

		// 5. Insert vote token
		token := &VoteToken{
			ElectionID: electionID,
			VoterID:    voterID,
			TokenHash:  tokenHash,
			IssuedAt:   now,
			Method:     channel,
			TPSID:      tpsID,
		}
		if err := s.voteRepo.InsertToken(ctx, tx, token); err != nil {
			return err
		}

		// 6. Insert vote
		vote := &Vote{
			ElectionID:  electionID,
			CandidateID: cand.ID,
			TokenHash:   tokenHash,
			Channel:     channel,
			TPSID:       tpsID,
			CastAt:      now,
		}
		if err := s.voteRepo.InsertVote(ctx, tx, vote); err != nil {
			return err
		}

		// 7. Update voter_status
		vs.HasVoted = true
		method := channel
		vs.VotingMethod = &method
		vs.TPSID = tpsID
		vs.VotedAt = &now
		vs.TokenHash = &tokenHash

		if err := s.voterRepo.UpdateStatus(ctx, tx, vs); err != nil {
			return err
		}

		// 8. Update stats (optional)
		if s.statsRepo != nil {
			if err := s.statsRepo.IncrementCandidateCount(ctx, tx, electionID, cand.ID, channel, tpsID); err != nil {
				return err
			}
		}

		// 9. Audit log (async-safe, errors ignored)
		if s.auditSvc != nil {
			_ = s.auditSvc.Log(ctx, AuditEntry{
				ActorVoterID: &voterID,
				Action:       "CAST_VOTE_" + channel,
				EntityType:   "VOTE",
				EntityID:     vote.ID,
				Metadata: map[string]any{
					"election_id": electionID,
					"channel":     channel,
					"tps_id":      tpsID,
				},
			})
		}

		// 10. Build result
		var tpsInfo *TPSInfo
		if tpsID != nil && channel == "TPS" {
			tpsEntry, err := s.voteRepo.GetTPSByID(ctx, tx, *tpsID)
			if err == nil {
				tpsInfo = &TPSInfo{
					ID:   tpsEntry.ID,
					Code: tpsEntry.Code,
					Name: tpsEntry.Name,
				}
			}
		}

		result = &VoteResultEntity{
			ElectionID: electionID,
			VoterID:    voterID,
			Method:     channel,
			VotedAt:    now,
			TPS:        tpsInfo,
			Receipt: ReceiptDetail{
				TokenHash: tokenHash,
				Note:      "",
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTPSVotingStatus checks if voter is eligible for TPS voting
func (s *Service) GetTPSVotingStatus(ctx context.Context, voterID int64) (*TPSVotingStatus, error) {
	// TODO: Implement
	// Check latest TPS check-in and its status

	return &TPSVotingStatus{
		Eligible: false,
		Reason:   stringPtr("TPS_REQUIRED"),
	}, nil
}

// GetVotingReceipt returns vote receipt without revealing candidate
func (s *Service) GetVotingReceipt(ctx context.Context, voterID int64) (*ReceiptResponse, error) {
	// TODO: Implement
	// Query voter_status and return receipt info

	return &ReceiptResponse{
		HasVoted: false,
	}, nil
}

func (s *Service) GetLiveCount(ctx context.Context, electionID int64) (map[int64]int64, error) {
	if s.repo == nil {
		return nil, errors.New("repository not initialized")
	}
	return s.repo.GetVoteCount(ctx, electionID)
}

func (s *Service) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", shared.ErrInternalServer
	}
	return hex.EncodeToString(bytes), nil
}

func stringPtr(s string) *string {
	return &s
}
