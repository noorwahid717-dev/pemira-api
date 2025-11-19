package monitoring

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetLiveCountSnapshot(ctx context.Context, electionID int64) (*LiveCountSnapshot, error) {
	// Get all stats in parallel
	voteStats, err := s.repo.GetVoteStats(ctx, electionID)
	if err != nil {
		return nil, err
	}

	participation, err := s.repo.GetParticipationStats(ctx, electionID)
	if err != nil {
		return nil, err
	}

	tpsStats, err := s.repo.GetTPSStats(ctx, electionID)
	if err != nil {
		return nil, err
	}

	// Build candidate votes map
	candidateVotes := make(map[int64]int64)
	var totalVotes int64
	for _, stat := range voteStats {
		candidateVotes[stat.CandidateID] = stat.TotalVotes
		totalVotes += stat.TotalVotes
	}

	return &LiveCountSnapshot{
		ElectionID:     electionID,
		Timestamp:      time.Now(),
		TotalVotes:     totalVotes,
		Participation:  *participation,
		CandidateVotes: candidateVotes,
		TPSStats:       tpsStats,
	}, nil
}

func (s *Service) GetDashboardSummary(ctx context.Context, electionID int64) (map[string]interface{}, error) {
	snapshot, err := s.GetLiveCountSnapshot(ctx, electionID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_votes":      snapshot.TotalVotes,
		"total_eligible":   snapshot.Participation.TotalEligible,
		"participation_pct": snapshot.Participation.ParticipationPct,
		"candidate_votes":  snapshot.CandidateVotes,
		"tps_count":        len(snapshot.TPSStats),
		"last_updated":     snapshot.Timestamp,
	}, nil
}
