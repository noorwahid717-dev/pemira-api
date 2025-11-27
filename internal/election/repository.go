package election

import "context"

type Repository interface {
	GetCurrentElection(ctx context.Context) (*Election, error)
	GetCurrentForRegistration(ctx context.Context) (*Election, error)
	GetActiveElection(ctx context.Context, settingsElectionID int) (*Election, error)
	ListPublicElections(ctx context.Context) ([]Election, error)
	GetByID(ctx context.Context, id int64) (*Election, error)
	GetVoterStatus(ctx context.Context, electionID, voterID int64) (*MeStatusRow, error)
	GetHistory(ctx context.Context, electionID, voterID, userID int64) (*MeHistoryDTO, error)
	IsRegistrationAllowed(ctx context.Context, election *Election) (bool, string)
}
