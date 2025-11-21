package election

import (
	"context"
	"errors"

	"pemira-api/internal/auth"
	"pemira-api/internal/shared/constants"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCurrentElection(ctx context.Context) (*CurrentElectionDTO, error) {
	e, err := s.repo.GetCurrentElection(ctx)
	if err != nil {
		return nil, err
	}

	return &CurrentElectionDTO{
		ID:            e.ID,
		Year:          e.Year,
		Name:          e.Name,
		Slug:          e.Slug,
		Status:        e.Status,
		VotingStartAt: e.VotingStartAt,
		VotingEndAt:   e.VotingEndAt,
		OnlineEnabled: e.OnlineEnabled,
		TPSEnabled:    e.TPSEnabled,
	}, nil
}

var (
	ErrUnauthorizedRole    = errors.New("role not allowed")
	ErrVoterMappingMissing = errors.New("voter mapping missing for user")
)

func (s *Service) GetMeStatus(
	ctx context.Context,
	authUser auth.AuthUser,
	electionID int64,
) (*MeStatusDTO, error) {
	if authUser.Role != constants.RoleStudent {
		return nil, ErrUnauthorizedRole
	}

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
