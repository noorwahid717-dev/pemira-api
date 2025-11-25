package election

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

type AdminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

var (
	ErrElectionAlreadyOpen      = errors.New("election already open for voting")
	ErrElectionAlreadyOpened    = errors.New("election already opened")
	ErrElectionAlreadyClosed    = errors.New("election already closed")
	ErrElectionNotInOpenState   = errors.New("election is not in voting-open state")
	ErrInvalidStatusChange      = errors.New("invalid election status change")
	ErrElectionArchived         = errors.New("election archived")
	ErrVotingPhaseLocked        = errors.New("voting phase locked")
	ErrInvalidPhaseKey          = errors.New("invalid phase key")
	ErrPhaseTimeConflict        = errors.New("phase time conflict")
	ErrInvalidModeCombination   = errors.New("invalid mode combination")
	ErrElectionAlreadyStarted   = errors.New("election already started")
	ErrElectionNotInVotingPhase = errors.New("election not in voting phase")
	ErrElectionNotClosable      = errors.New("election not closable")
)

func (s *AdminService) List(
	ctx context.Context,
	filter AdminElectionListFilter,
	page, limit int,
) ([]AdminElectionDTO, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	items, total, err := s.repo.ListElections(ctx, filter)
	if err != nil {
		return nil, Pagination{}, err
	}

	for i := range items {
		s.enrichElection(&items[i])
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	p := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}
	return items, p, nil
}

func (s *AdminService) Create(ctx context.Context, req AdminElectionCreateRequest) (*AdminElectionDTO, error) {
	dto, err := s.repo.CreateElection(ctx, req)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) Get(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	dto, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) Update(ctx context.Context, id int64, req AdminElectionUpdateRequest) (*AdminElectionDTO, error) {
	dto, err := s.repo.UpdateElection(ctx, id, req)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) PatchGeneralInfo(ctx context.Context, id int64, req AdminElectionGeneralUpdateRequest) (*AdminElectionDTO, error) {
	current, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Status == ElectionStatusArchived {
		return nil, ErrElectionArchived
	}

	updated, err := s.repo.UpdateGeneralInfo(ctx, id, req)
	if err != nil {
		return nil, err
	}
	s.enrichElection(updated)
	return updated, nil
}

func (s *AdminService) GetPhases(ctx context.Context, id int64) (*ElectionPhasesResponse, error) {
	dto, err := s.repo.GetPhases(ctx, id)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	resp := s.buildPhasesResponse(dto)
	return resp, nil
}

func (s *AdminService) UpdatePhases(ctx context.Context, id int64, req UpdateElectionPhasesRequest) (*ElectionPhasesResponse, error) {
	if err := validatePhaseInputs(req.Phases); err != nil {
		return nil, err
	}

	current, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.Status == ElectionStatusVotingOpen {
		for _, ph := range req.Phases {
			if ph.Key == PhaseKeyVoting {
				return nil, ErrVotingPhaseLocked
			}
		}
	}

	updated, err := s.repo.UpdatePhases(ctx, id, req.Phases)
	if err != nil {
		return nil, err
	}
	s.enrichElection(updated)
	return s.buildPhasesResponse(updated), nil
}

func (s *AdminService) GetModeSettings(ctx context.Context, id int64) (*ModeSettingsDTO, error) {
	if _, err := s.repo.GetElectionByID(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.GetModeSettings(ctx, id)
}

func (s *AdminService) UpdateModeSettings(ctx context.Context, id int64, req ModeSettingsRequest) (*ModeSettingsDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	currentOnline := e.OnlineEnabled
	currentTPS := e.TPSEnabled

	if req.OnlineEnabled != nil {
		currentOnline = *req.OnlineEnabled
	}
	if req.TPSEnabled != nil {
		currentTPS = *req.TPSEnabled
	}

	if !currentOnline && !currentTPS {
		return nil, ErrInvalidModeCombination
	}

	switch e.Status {
	case ElectionStatusVotingOpen, ElectionStatusVotingClosed, ElectionStatusClosed, ElectionStatusArchived:
		return nil, ErrElectionAlreadyStarted
	}

	return s.repo.UpdateModeSettings(ctx, id, req)
}

func (s *AdminService) GetSummary(ctx context.Context, id int64) (*ElectionSummaryDTO, error) {
	election, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	summary, err := s.repo.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	summary.Status = election.Status
	summary.CurrentPhase = s.computeCurrentPhase(election, now)
	return summary, nil
}

func (s *AdminService) OpenVoting(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	if e.Status == ElectionStatusVotingOpen {
		return nil, ErrElectionAlreadyOpened
	}
	if e.Status == ElectionStatusArchived {
		return nil, ErrInvalidStatusChange
	}

	if e.VotingStartAt != nil && now.Before(*e.VotingStartAt) {
		return nil, ErrElectionNotInVotingPhase
	}
	if e.VotingEndAt != nil && now.After(*e.VotingEndAt) {
		return nil, ErrElectionNotInVotingPhase
	}

	startAt := e.VotingStartAt
	if startAt == nil {
		startAt = &now
	}

	currentPhase := string(PhaseKeyVoting)

	dto, err := s.repo.SetVotingStatus(ctx, id, ElectionStatusVotingOpen, &currentPhase, startAt, e.VotingEndAt)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) CloseVoting(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Status == ElectionStatusVotingClosed || e.Status == ElectionStatusClosed {
		return nil, ErrElectionAlreadyClosed
	}

	if e.Status != ElectionStatusVotingOpen {
		return nil, ErrElectionNotInOpenState
	}

	now := time.Now().UTC()
	currentPhase := string(PhaseKeyRecap)
	dto, err := s.repo.SetVotingStatus(ctx, id, ElectionStatusVotingClosed, &currentPhase, nil, &now)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) Archive(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Status == ElectionStatusArchived {
		return nil, ErrElectionArchived
	}

	if e.Status != ElectionStatusVotingClosed && e.Status != ElectionStatusRecap && e.Status != ElectionStatusClosed {
		return nil, ErrElectionNotClosable
	}

	currentPhase := string(PhaseKeyRecap)
	dto, err := s.repo.SetVotingStatus(ctx, id, ElectionStatusArchived, &currentPhase, nil, nil)
	if err != nil {
		return nil, err
	}
	s.enrichElection(dto)
	return dto, nil
}

func (s *AdminService) GetBranding(ctx context.Context, electionID int64) (*BrandingSettings, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.GetBranding(ctx, electionID)
}

func (s *AdminService) GetBrandingLogo(ctx context.Context, electionID int64, slot BrandingSlot) (*BrandingFile, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.GetBrandingFile(ctx, electionID, slot)
}

func (s *AdminService) UploadBrandingLogo(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	file BrandingFileCreate,
) (*BrandingFile, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.SaveBrandingFile(ctx, electionID, slot, file)
}

func (s *AdminService) DeleteBrandingLogo(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	adminID int64,
) (*BrandingSettings, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.DeleteBrandingFile(ctx, electionID, slot, adminID)
}

func (s *AdminService) enrichElection(dto *AdminElectionDTO) {
	if dto == nil {
		return
	}

	if dto.AcademicYear == nil {
		ay := fmt.Sprintf("%d/%d", dto.Year-1, dto.Year)
		dto.AcademicYear = &ay
	}

	now := time.Now().UTC()
	current := s.computeCurrentPhase(dto, now)
	dto.CurrentPhase = &current
}

func (s *AdminService) computeCurrentPhase(dto *AdminElectionDTO, now time.Time) string {
	for _, ph := range phaseDTOsFromElection(dto) {
		if ph.StartAt != nil && ph.EndAt != nil {
			if !now.Before(*ph.StartAt) && !now.After(*ph.EndAt) {
				return string(ph.Key)
			}
		}
	}

	switch dto.Status {
	case ElectionStatusRegistration, ElectionStatusRegistrationOpen:
		return string(PhaseKeyRegistration)
	case ElectionStatusVerification:
		return string(PhaseKeyVerification)
	case ElectionStatusCampaign:
		return string(PhaseKeyCampaign)
	case ElectionStatusQuietPeriod:
		return string(PhaseKeyQuietPeriod)
	case ElectionStatusVotingOpen:
		return string(PhaseKeyVoting)
	case ElectionStatusVotingClosed, ElectionStatusRecap, ElectionStatusClosed:
		return string(PhaseKeyRecap)
	case ElectionStatusArchived:
		return string(PhaseKeyRecap)
	default:
		return string(dto.Status)
	}
}

var phaseLabels = map[ElectionPhaseKey]string{
	PhaseKeyRegistration: "Pendaftaran",
	PhaseKeyVerification: "Verifikasi Berkas",
	PhaseKeyCampaign:     "Masa Kampanye",
	PhaseKeyQuietPeriod:  "Masa Tenang",
	PhaseKeyVoting:       "Voting",
	PhaseKeyRecap:        "Rekapitulasi",
}

func phaseDTOsFromElection(dto *AdminElectionDTO) []ElectionPhaseDTO {
	return []ElectionPhaseDTO{
		{Key: PhaseKeyRegistration, Label: phaseLabels[PhaseKeyRegistration], StartAt: dto.RegistrationStartAt, EndAt: dto.RegistrationEndAt},
		{Key: PhaseKeyVerification, Label: phaseLabels[PhaseKeyVerification], StartAt: dto.VerificationStartAt, EndAt: dto.VerificationEndAt},
		{Key: PhaseKeyCampaign, Label: phaseLabels[PhaseKeyCampaign], StartAt: dto.CampaignStartAt, EndAt: dto.CampaignEndAt},
		{Key: PhaseKeyQuietPeriod, Label: phaseLabels[PhaseKeyQuietPeriod], StartAt: dto.QuietStartAt, EndAt: dto.QuietEndAt},
		{Key: PhaseKeyVoting, Label: phaseLabels[PhaseKeyVoting], StartAt: dto.VotingStartAt, EndAt: dto.VotingEndAt},
		{Key: PhaseKeyRecap, Label: phaseLabels[PhaseKeyRecap], StartAt: dto.RecapStartAt, EndAt: dto.RecapEndAt},
	}
}

func (s *AdminService) buildPhasesResponse(dto *AdminElectionDTO) *ElectionPhasesResponse {
	return &ElectionPhasesResponse{
		ElectionID: dto.ID,
		Phases:     phaseDTOsFromElection(dto),
	}
}

func validatePhaseInputs(phases []ElectionPhaseInput) error {
	if len(phases) == 0 {
		return ErrInvalidPhaseKey
	}

	seen := map[ElectionPhaseKey]bool{}

	for _, ph := range phases {
		if _, ok := phaseColumnMap[ph.Key]; !ok {
			return ErrInvalidPhaseKey
		}
		if seen[ph.Key] {
			return ErrInvalidPhaseKey
		}
		seen[ph.Key] = true

		if ph.StartAt == nil || ph.EndAt == nil {
			return ErrInvalidPhaseKey
		}
		if !ph.StartAt.Before(*ph.EndAt) {
			return ErrPhaseTimeConflict
		}
	}

	if len(seen) != len(phaseColumnMap) {
		return ErrInvalidPhaseKey
	}

	type window struct {
		start time.Time
		end   time.Time
	}
	windows := []window{}
	for _, ph := range phases {
		if ph.StartAt != nil && ph.EndAt != nil {
			windows = append(windows, window{start: *ph.StartAt, end: *ph.EndAt})
		}
	}

	if len(windows) > 1 {
		sort.Slice(windows, func(i, j int) bool {
			return windows[i].start.Before(windows[j].start)
		})
		for i := 1; i < len(windows); i++ {
			if windows[i-1].end.After(windows[i].start) {
				return ErrPhaseTimeConflict
			}
		}
	}

	return nil
}
