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
	UpdateGeneralInfo(ctx context.Context, id int64, req AdminElectionGeneralUpdateRequest) (*AdminElectionDTO, error)
	SetVotingStatus(ctx context.Context, id int64, status ElectionStatus, currentPhase *string, votingStartAt, votingEndAt *time.Time) (*AdminElectionDTO, error)
	GetPhases(ctx context.Context, id int64) (*AdminElectionDTO, error)
	UpdatePhases(ctx context.Context, id int64, phases []ElectionPhaseInput) (*AdminElectionDTO, error)
	GetModeSettings(ctx context.Context, id int64) (*ModeSettingsDTO, error)
	UpdateModeSettings(ctx context.Context, id int64, req ModeSettingsRequest) (*ModeSettingsDTO, error)
	GetSummary(ctx context.Context, id int64) (*ElectionSummaryDTO, error)
	GetBranding(ctx context.Context, electionID int64) (*BrandingSettings, error)
	GetBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot) (*BrandingFile, error)
	SaveBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot, file BrandingFileCreate) (*BrandingFile, error)
	DeleteBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot, adminID int64) (*BrandingSettings, error)
}
