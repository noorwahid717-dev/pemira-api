package voting

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
	
	"pemira-api/internal/shared"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CastVote(ctx context.Context, electionID, voterID, candidateID int64, votedVia string) error {
	token, err := s.generateToken()
	if err != nil {
		return err
	}

	voteToken := &VoteToken{
		VoterID:    voterID,
		ElectionID: electionID,
		Token:      token,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateToken(ctx, voteToken); err != nil {
		return err
	}

	vote := &Vote{
		ElectionID:  electionID,
		CandidateID: candidateID,
		TokenHash:   token,
		VotedAt:     time.Now(),
		VotedVia:    votedVia,
	}

	return s.repo.CreateVote(ctx, vote)
}

func (s *Service) GetLiveCount(ctx context.Context, electionID int64) (map[int64]int64, error) {
	return s.repo.GetVoteCount(ctx, electionID)
}

func (s *Service) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", shared.ErrInternalServer
	}
	return hex.EncodeToString(bytes), nil
}
