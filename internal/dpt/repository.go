package dpt

import "context"

type Repository interface {
	ImportVotersForElection(ctx context.Context, electionID int64, rows []ImportRow) (*ImportResult, error)
	ListVotersForElection(ctx context.Context, electionID int64, filter ListFilter) ([]VoterWithStatusDTO, int64, error)
	StreamVotersForElection(ctx context.Context, electionID int64, filter ListFilter, fn func(VoterWithStatusDTO) error) error
}
