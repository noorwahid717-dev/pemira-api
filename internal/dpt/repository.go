package dpt

import "context"

type Repository interface {
	ImportVotersForElection(ctx context.Context, electionID int64, rows []ImportRow) (*ImportResult, error)
	ListAllVoters(ctx context.Context, filter ListFilter) ([]VoterWithStatusDTO, int64, error)
	ListVotersForElection(ctx context.Context, electionID int64, filter ListFilter) ([]VoterWithStatusDTO, int64, error)
	StreamVotersForElection(ctx context.Context, electionID int64, filter ListFilter, fn func(VoterWithStatusDTO) error) error
	GetVoterByID(ctx context.Context, electionID int64, voterID int64) (*VoterWithStatusDTO, error)
	UpdateVoter(ctx context.Context, electionID int64, voterID int64, updates VoterUpdateDTO) error
	DeleteVoter(ctx context.Context, electionID int64, voterID int64) error
}
