package election

import (
	"context"
	"errors"
	"time"

	"pemira-api/internal/auth"
)

type Service struct {
	repo      Repository
	adminRepo AdminRepository
}

func NewService(repo Repository, adminRepo AdminRepository) *Service {
	return &Service{
		repo:      repo,
		adminRepo: adminRepo,
	}
}

func (s *Service) GetCurrentElection(ctx context.Context) (*CurrentElectionDTO, error) {
	// Try to get VOTING_OPEN election first
	e, err := s.repo.GetCurrentElection(ctx)
	if err != nil {
		// If no VOTING_OPEN, get the most recent non-archived election
		elections, listErr := s.repo.ListPublicElections(ctx)
		if listErr != nil || len(elections) == 0 {
			return nil, ErrElectionNotFound
		}
		e = &elections[0] // Use first (most recent by year DESC)
	}

	dto := &CurrentElectionDTO{
		ID:            e.ID,
		Year:          e.Year,
		Name:          e.Name,
		Slug:          e.Slug,
		Status:        e.Status,
		VotingStartAt: e.VotingStartAt,
		VotingEndAt:   e.VotingEndAt,
		OnlineEnabled: e.OnlineEnabled,
		TPSEnabled:    e.TPSEnabled,
	}

	s.enrichWithPhases(ctx, dto)

	return dto, nil
}

func (s *Service) ListPublicElections(ctx context.Context) ([]CurrentElectionDTO, error) {
	elections, err := s.repo.ListPublicElections(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CurrentElectionDTO, len(elections))
	for i, e := range elections {
		result[i] = CurrentElectionDTO{
			ID:            e.ID,
			Year:          e.Year,
			Name:          e.Name,
			Slug:          e.Slug,
			Status:        e.Status,
			VotingStartAt: e.VotingStartAt,
			VotingEndAt:   e.VotingEndAt,
			OnlineEnabled: e.OnlineEnabled,
			TPSEnabled:    e.TPSEnabled,
		}

		// Only enrich current/focused elections to reduce overhead
		s.enrichWithPhases(ctx, &result[i])
	}

	return result, nil
}

func (s *Service) GetPublicPhases(ctx context.Context, electionID int64) (*ElectionPhasesResponse, error) {
	// Verify election exists
	_, err := s.repo.GetByID(ctx, electionID)
	if err != nil {
		return nil, err
	}

	// Get election with phases from admin repo
	dto, err := s.adminRepo.GetPhases(ctx, electionID)
	if err != nil {
		return nil, err
	}

	// Build phases response
	response := &ElectionPhasesResponse{
		ElectionID: dto.ID,
		Phases:     buildPhasesList(dto),
	}

	return response, nil
}

func (s *Service) enrichWithPhases(ctx context.Context, dto *CurrentElectionDTO) {
	if s.adminRepo == nil || dto == nil {
		return
	}

	adm, err := s.adminRepo.GetPhases(ctx, dto.ID)
	if err != nil || adm == nil {
		return
	}

	phases := buildPhasesList(adm)
	dto.Phases = phases

	// Prefer voting window from phases if available
	for _, ph := range phases {
		if ph.Key == PhaseKeyVoting && ph.StartAt != nil && ph.EndAt != nil {
			dto.VotingStartAt = ph.StartAt
			dto.VotingEndAt = ph.EndAt
			break
		}
	}

	// Set current phase
	if current := deriveCurrentPhase(phases); current != "" {
		dto.CurrentPhase = current
	}
}

func deriveCurrentPhase(phases []ElectionPhaseDTO) string {
	now := time.Now()
	for _, ph := range phases {
		if ph.StartAt != nil && ph.EndAt != nil {
			if (now.Equal(*ph.StartAt) || now.After(*ph.StartAt)) && now.Before(*ph.EndAt) {
				return string(ph.Key)
			}
		}
	}
	return ""
}

func buildPhasesList(dto *AdminElectionDTO) []ElectionPhaseDTO {
	phases := []ElectionPhaseDTO{}

	if dto.RegistrationStartAt != nil && dto.RegistrationEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "REGISTRATION",
			Label:   "Pendaftaran",
			StartAt: dto.RegistrationStartAt,
			EndAt:   dto.RegistrationEndAt,
		})
	}

	if dto.VerificationStartAt != nil && dto.VerificationEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "VERIFICATION",
			Label:   "Verifikasi Berkas",
			StartAt: dto.VerificationStartAt,
			EndAt:   dto.VerificationEndAt,
		})
	}

	if dto.CampaignStartAt != nil && dto.CampaignEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "CAMPAIGN",
			Label:   "Masa Kampanye",
			StartAt: dto.CampaignStartAt,
			EndAt:   dto.CampaignEndAt,
		})
	}

	if dto.QuietStartAt != nil && dto.QuietEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "QUIET_PERIOD",
			Label:   "Masa Tenang",
			StartAt: dto.QuietStartAt,
			EndAt:   dto.QuietEndAt,
		})
	}

	if dto.VotingStartAt != nil && dto.VotingEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "VOTING",
			Label:   "Voting",
			StartAt: dto.VotingStartAt,
			EndAt:   dto.VotingEndAt,
		})
	}

	if dto.RecapStartAt != nil && dto.RecapEndAt != nil {
		phases = append(phases, ElectionPhaseDTO{
			Key:     "RECAP",
			Label:   "Rekapitulasi",
			StartAt: dto.RecapStartAt,
			EndAt:   dto.RecapEndAt,
		})
	}

	return phases
}

var (
	ErrVoterMappingMissing = errors.New("voter mapping missing for user")
)

func (s *Service) GetMeStatus(
	ctx context.Context,
	authUser auth.AuthUser,
	electionID int64,
) (*MeStatusDTO, error) {
	if authUser.VoterID == nil {
		return nil, ErrVoterMappingMissing
	}

	if _, err := s.repo.GetByID(ctx, electionID); err != nil {
		return nil, err
	}

	row, err := s.repo.GetVoterStatus(ctx, electionID, *authUser.VoterID)
	if err != nil {
		if errors.Is(err, ErrVoterStatusNotFound) {
			return &MeStatusDTO{
				ElectionID:    electionID,
				VoterID:       *authUser.VoterID,
				Eligible:      false,
				HasVoted:      false,
				Method:        VoteMethodNone,
				TPSID:         nil,
				LastVoteAt:    nil,
				OnlineAllowed: false,
				TPSAllowed:    false,
			}, nil
		}
		return nil, err
	}

	method := VoteMethodNone
	if row.LastVoteChannel != nil {
		switch *row.LastVoteChannel {
		case string(VoteMethodOnline):
			method = VoteMethodOnline
		case string(VoteMethodTPS):
			method = VoteMethodTPS
		}
	} else if row.PreferredMethod != nil {
		if *row.PreferredMethod == string(VoteMethodOnline) {
			method = VoteMethodOnline
		} else if *row.PreferredMethod == string(VoteMethodTPS) {
			method = VoteMethodTPS
		}
	}

	dto := &MeStatusDTO{
		ElectionID:      row.ElectionID,
		VoterID:         row.VoterID,
		Eligible:        row.IsEligible,
		HasVoted:        row.HasVoted,
		Method:          method,
		TPSID:           row.LastTPSID,
		LastVoteAt:      row.LastVoteAt,
		PreferredMethod: row.PreferredMethod,
		OnlineAllowed:   row.OnlineAllowed,
		TPSAllowed:      row.TPSAllowed,
	}

	return dto, nil
}
