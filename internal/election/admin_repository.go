package election

import (
	"context"
	"time"
)

type AdminRepository interface {
	ListElections(ctx context.Context, filter AdminElectionListFilter) ([]AdminElectionDTO, int64, error)
	GetElectionByID(ctx context.Context, id int64) (*AdminElectionDTO, error)
	CreateElection(ctx context.Context, req AdminElectionCreateRequest) (*AdminElectionDTO, error)
	UpdateElection(ctx context.Context, id int64, req AdminElectionUpdateRequest) (*AdminElectionDTO, error)
	SetVotingStatus(ctx context.Context, id int64, status ElectionStatus, votingStartAt, votingEndAt *time.Time) (*AdminElectionDTO, error)
	GetBranding(ctx context.Context, electionID int64) (*BrandingSettings, error)
	GetBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot) (*BrandingFile, error)
	SaveBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot, file BrandingFileCreate) (*BrandingFile, error)
	DeleteBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot, adminID int64) (*BrandingSettings, error)
}
