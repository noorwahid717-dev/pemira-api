package voter

import (
	"context"
	"pemira-api/internal/shared"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Voter, error)
	GetByNIM(ctx context.Context, nim string) (*Voter, error)
	List(ctx context.Context, params shared.PaginationParams) ([]*Voter, int64, error)
	Create(ctx context.Context, voter *Voter) error
	Update(ctx context.Context, voter *Voter) error
	Delete(ctx context.Context, id int64) error
	
	GetElectionStatus(ctx context.Context, voterID, electionID int64) (*VoterElectionStatus, error)
	CreateElectionStatus(ctx context.Context, status *VoterElectionStatus) error
	UpdateElectionStatus(ctx context.Context, status *VoterElectionStatus) error
	
	// Profile methods
	GetActiveElectionID(ctx context.Context) (int64, error)
	GetCompleteProfile(ctx context.Context, voterID int64, userID int64, electionID int64) (*CompleteProfileResponse, error)
	UpdateProfile(ctx context.Context, voterID int64, req *UpdateProfileRequest) error
	UpdateVotingMethod(ctx context.Context, voterID, electionID int64, method string) error
	GetParticipationStats(ctx context.Context, voterID int64) (*ParticipationStatsResponse, error)
	DeletePhoto(ctx context.Context, voterID int64) error
}

type AuthRepository interface {
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error
}

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}
