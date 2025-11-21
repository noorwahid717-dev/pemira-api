package voting

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pemira-api/internal/auth"
	"pemira-api/internal/election"
	"pemira-api/internal/shared"
	"pemira-api/internal/shared/constants"
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

type SetMethodRequest struct {
	ElectionID int64
	Method     string
	TPSID      *int64
}

type ScanCandidateRequest struct {
	TPSID     int64
	CheckinID int64
	Payload   string
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

// CastOnlineVote handles online voting with full validation
func (s *Service) CastOnlineVote(ctx context.Context, authUser auth.AuthUser, req CastOnlineVoteRequest) error {
	if s.db == nil || s.electionRepo == nil {
		return errors.New("not implemented")
	}

	// Check if user has voter mapping
	if authUser.Role != constants.RoleStudent || authUser.VoterID == nil {
		return ErrVoterMappingMissing
	}

	voterID := *authUser.VoterID

	// 1. Get election
	election, err := s.electionRepo.GetByID(ctx, req.ElectionID)
	if err != nil {
		return translateNotFound(err, ErrElectionNotFound)
	}

	// 2. Validate election status
	if election.Status != "VOTING_OPEN" {
		return ErrElectionNotOpen
	}

	// 3. Validate online mode enabled
	if !election.OnlineEnabled {
		return ErrMethodNotAllowed
	}

	// 4. Cast vote with transaction
	_, err = s.castVote(ctx, req.ElectionID, voterID, req.CandidateID, "ONLINE", nil)
	return err
}

// CastTPSVote handles TPS voting after check-in approval
func (s *Service) CastTPSVote(ctx context.Context, authUser auth.AuthUser, req CastTPSVoteRequest) error {
	if s.db == nil || s.electionRepo == nil {
		return errors.New("not implemented")
	}

	// Check if user has voter mapping
	if authUser.Role != constants.RoleStudent || authUser.VoterID == nil {
		return ErrVoterMappingMissing
	}

	voterID := *authUser.VoterID

	// 1. Get election
	election, err := s.electionRepo.GetByID(ctx, req.ElectionID)
	if err != nil {
		return translateNotFound(err, ErrElectionNotFound)
	}

	// 2. Validate election status
	if election.Status != "VOTING_OPEN" {
		return ErrElectionNotOpen
	}

	// 3. Validate TPS mode enabled
	if !election.TPSEnabled {
		return ErrMethodNotAllowed
	}

	// 4. Get & validate latest approved check-in
	var checkin *tps.TPSCheckin

	err = s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		checkin, err = s.voteRepo.GetLatestApprovedCheckin(ctx, tx, req.ElectionID, voterID)
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

		// Validate TPS ID matches request
		if checkin.TPSID != req.TPSID {
			return ErrTPSNotFound
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 5. Cast vote with TPS info
	_, err = s.castVote(ctx, req.ElectionID, voterID, req.CandidateID, "TPS", &req.TPSID)
	if err != nil {
		return err
	}

	// 6. Mark check-in as used
	_ = s.withTx(ctx, func(tx pgx.Tx) error {
		return s.voteRepo.MarkCheckinUsed(ctx, tx, checkin.ID, time.Now().UTC())
	})

	return nil
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
		// Note: IsActive field removed from Candidate struct
		// if !cand.IsActive {
		// 	return ErrCandidateInactive
		// }

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

// SetVoterMethod sets the preferred voting method (ONLINE or TPS) for a voter in a specific election.
func (s *Service) SetVoterMethod(ctx context.Context, authUser auth.AuthUser, req SetMethodRequest) error {
	if s.db == nil {
		return errors.New("service not initialized")
	}

	if authUser.Role != constants.RoleStudent || authUser.VoterID == nil {
		return ErrVoterMappingMissing
	}

	voterID := *authUser.VoterID

	// Validate method
	method := strings.ToUpper(req.Method)
	if method != "ONLINE" && method != "TPS" {
		return ErrMethodNotAllowed
	}

	return s.withTx(ctx, func(tx pgx.Tx) error {
		// Get election to validate status and channel availability
		election, err := s.electionRepo.GetByID(ctx, req.ElectionID)
		if err != nil {
			return translateNotFound(err, ErrElectionNotFound)
		}

		if election.Status == "CLOSED" || election.Status == "ARCHIVED" {
			return ErrElectionNotOpen
		}

		if method == "ONLINE" && !election.OnlineEnabled {
			return ErrMethodNotAllowed
		}
		if method == "TPS" && !election.TPSEnabled {
			return ErrMethodNotAllowed
		}

		// Lock voter_status row
		status, err := s.voterRepo.GetStatusForUpdate(ctx, tx, req.ElectionID, voterID)
		if err != nil {
			return translateNotFound(err, ErrNotEligible)
		}

		if status.HasVoted {
			return ErrAlreadyVoted
		}

		var tpsID *int64
		if method == "TPS" {
			if req.TPSID == nil || *req.TPSID <= 0 {
				return ErrTPSNotFound
			}
			id := *req.TPSID
			tpsID = &id
		}

		status.VotingMethod = &method
		status.TPSID = tpsID

		return s.voterRepo.UpdateStatus(ctx, tx, status)
	})
}

// ScanCandidateAtTPS handles QR ballot scan after check-in approval.
func (s *Service) ScanCandidateAtTPS(ctx context.Context, authUser auth.AuthUser, req ScanCandidateRequest) (*VoteResultEntity, error) {
	if s.db == nil {
		return nil, errors.New("service not initialized")
	}

	// Only TPS operators (or ketua_tps via admin role) with matching TPSID
	if authUser.Role != constants.RoleTPSOperator && authUser.Role != constants.RoleAdmin {
		return nil, ErrMethodNotAllowed
	}
	if authUser.TPSID == nil || *authUser.TPSID != req.TPSID {
		return nil, ErrTPSNotFound
	}

	qr, err := parseBallotQR(req.Payload)
	if err != nil {
		return nil, ErrInvalidBallotQR
	}

	var result *VoteResultEntity
	err = s.withTx(ctx, func(tx pgx.Tx) error {
		checkin, err := s.voteRepo.GetCheckinByID(ctx, tx, req.CheckinID)
		if err != nil {
			return translateNotFound(err, ErrTPSCheckinNotFound)
		}
		if checkin.TPSID != req.TPSID {
			return ErrTPSNotFound
		}
		if checkin.Status != tps.CheckinStatusApproved {
			return ErrTPSCheckinNotApproved
		}
		if checkin.ExpiresAt != nil && checkin.ExpiresAt.Before(time.Now().UTC()) {
			return ErrCheckinExpired
		}

		// Election match
		if checkin.ElectionID != qr.ElectionID {
			return ErrElectionMismatch
		}

		// Get election
		election, err := s.electionRepo.GetByID(ctx, qr.ElectionID)
		if err != nil {
			return translateNotFound(err, ErrElectionNotFound)
		}
		if election.Status != "VOTING_OPEN" || !election.TPSEnabled {
			return ErrElectionNotOpen
		}

		// Lock voter_status
		status, err := s.voterRepo.GetStatusForUpdate(ctx, tx, qr.ElectionID, checkin.VoterID)
		if err != nil {
			return translateNotFound(err, ErrNotEligible)
		}
		if status.HasVoted {
			return ErrAlreadyVoted
		}

		// Validate candidate belongs to election
		cand, err := s.candidateRepo.GetByIDWithTx(ctx, tx, qr.CandidateID)
		if err != nil || cand.ElectionID != qr.ElectionID {
			return ErrCandidateNotFound
		}

		now := time.Now().UTC()
		tokenHash := generateTokenHash(qr.ElectionID, checkin.VoterID)

		// Insert vote
		vote := &Vote{
			ElectionID:  qr.ElectionID,
			CandidateID: qr.CandidateID,
			TokenHash:   tokenHash,
			Channel:     "TPS",
			TPSID:       &req.TPSID,
			CastAt:      now,
		}
		if err := s.voteRepo.InsertVote(ctx, tx, vote); err != nil {
			return err
		}

		// Update voter_status
		status.HasVoted = true
		status.VotingMethod = stringPtr("TPS")
		status.TPSID = &req.TPSID
		status.VotedAt = &now
		status.TokenHash = &tokenHash
		if err := s.voterRepo.UpdateStatus(ctx, tx, status); err != nil {
			return err
		}

		// Update checkin status to USED
		if err := s.voteRepo.MarkCheckinUsed(ctx, tx, checkin.ID, now); err != nil {
			return err
		}

		result = &VoteResultEntity{
			ElectionID: qr.ElectionID,
			VoterID:    checkin.VoterID,
			Method:     "TPS",
			VotedAt:    now,
			TPS:        &TPSInfo{ID: checkin.TPSID},
			Receipt: ReceiptDetail{
				TokenHash: tokenHash,
				Note:      "Vote dicatat melalui scan QR surat suara.",
			},
		}
		return nil
	})

	return result, err
}
