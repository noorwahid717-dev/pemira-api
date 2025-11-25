package election

import "time"

type AdminElectionDTO struct {
	ID                  int64          `json:"id"`
	Year                int            `json:"year"`
	Name                string         `json:"name"`
	Slug                string         `json:"slug"`
	Description         string         `json:"description"`
	AcademicYear        *string        `json:"academic_year,omitempty"`
	Status              ElectionStatus `json:"status"`
	CurrentPhase        *string        `json:"current_phase,omitempty"`
	RegistrationStartAt *time.Time     `json:"registration_start_at,omitempty"`
	RegistrationEndAt   *time.Time     `json:"registration_end_at,omitempty"`
	VerificationStartAt *time.Time     `json:"verification_start_at,omitempty"`
	VerificationEndAt   *time.Time     `json:"verification_end_at,omitempty"`
	CampaignStartAt     *time.Time     `json:"campaign_start_at,omitempty"`
	CampaignEndAt       *time.Time     `json:"campaign_end_at,omitempty"`
	QuietStartAt        *time.Time     `json:"quiet_start_at,omitempty"`
	QuietEndAt          *time.Time     `json:"quiet_end_at,omitempty"`
	VotingStartAt       *time.Time     `json:"voting_start_at,omitempty"`
	VotingEndAt         *time.Time     `json:"voting_end_at,omitempty"`
	RecapStartAt        *time.Time     `json:"recap_start_at,omitempty"`
	RecapEndAt          *time.Time     `json:"recap_end_at,omitempty"`
	AnnouncementAt      *time.Time     `json:"announcement_at,omitempty"`
	FinishedAt          *time.Time     `json:"finished_at,omitempty"`
	OnlineEnabled       bool           `json:"online_enabled"`
	TPSEnabled          bool           `json:"tps_enabled"`
	OnlineLoginURL      *string        `json:"online_login_url,omitempty"`
	OnlineMaxSessions   *int           `json:"online_max_sessions_per_voter,omitempty"`
	TPSRequireCheckin   *bool          `json:"tps_require_checkin,omitempty"`
	TPSRequireBallotQR  *bool          `json:"tps_require_ballot_qr,omitempty"`
	TPSMax              *int           `json:"tps_max,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type AdminElectionListFilter struct {
	Year   *int
	Status *ElectionStatus
	Search string
	Limit  int
	Offset int
}

type AdminElectionCreateRequest struct {
	Year          int    `json:"year"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	OnlineEnabled bool   `json:"online_enabled"`
	TPSEnabled    bool   `json:"tps_enabled"`
}

type AdminElectionUpdateRequest struct {
	Year                *int       `json:"year,omitempty"`
	Name                *string    `json:"name,omitempty"`
	Slug                *string    `json:"slug,omitempty"`
	OnlineEnabled       *bool      `json:"online_enabled,omitempty"`
	TPSEnabled          *bool      `json:"tps_enabled,omitempty"`
	RegistrationStartAt *time.Time `json:"registration_start_at,omitempty"`
	RegistrationEndAt   *time.Time `json:"registration_end_at,omitempty"`
	VerificationStartAt *time.Time `json:"verification_start_at,omitempty"`
	VerificationEndAt   *time.Time `json:"verification_end_at,omitempty"`
	CampaignStartAt     *time.Time `json:"campaign_start_at,omitempty"`
	CampaignEndAt       *time.Time `json:"campaign_end_at,omitempty"`
	QuietStartAt        *time.Time `json:"quiet_start_at,omitempty"`
	QuietEndAt          *time.Time `json:"quiet_end_at,omitempty"`
	VotingStartAt       *time.Time `json:"voting_start_at,omitempty"`
	VotingEndAt         *time.Time `json:"voting_end_at,omitempty"`
	RecapStartAt        *time.Time `json:"recap_start_at,omitempty"`
	RecapEndAt          *time.Time `json:"recap_end_at,omitempty"`
	AnnouncementAt      *time.Time `json:"announcement_at,omitempty"`
	FinishedAt          *time.Time `json:"finished_at,omitempty"`
}

type AdminElectionGeneralUpdateRequest struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	AcademicYear *string `json:"academic_year,omitempty"`
}

type ElectionPhaseKey string

const (
	PhaseKeyRegistration ElectionPhaseKey = "REGISTRATION"
	PhaseKeyVerification ElectionPhaseKey = "VERIFICATION"
	PhaseKeyCampaign     ElectionPhaseKey = "CAMPAIGN"
	PhaseKeyQuietPeriod  ElectionPhaseKey = "QUIET_PERIOD"
	PhaseKeyVoting       ElectionPhaseKey = "VOTING"
	PhaseKeyRecap        ElectionPhaseKey = "RECAP"
)

type ElectionPhaseDTO struct {
	Key     ElectionPhaseKey `json:"key"`
	Label   string           `json:"label"`
	StartAt *time.Time       `json:"start_at,omitempty"`
	EndAt   *time.Time       `json:"end_at,omitempty"`
}

type UpdateElectionPhasesRequest struct {
	Phases []ElectionPhaseInput `json:"phases"`
}

type ElectionPhaseInput struct {
	Key     ElectionPhaseKey `json:"key"`
	StartAt *time.Time       `json:"start_at,omitempty"`
	EndAt   *time.Time       `json:"end_at,omitempty"`
}

type ElectionPhasesResponse struct {
	ElectionID int64              `json:"election_id"`
	Phases     []ElectionPhaseDTO `json:"phases"`
}

type VotingWindow struct {
	StartAt *time.Time `json:"start_at,omitempty"`
	EndAt   *time.Time `json:"end_at,omitempty"`
}

type ModeSettingsDTO struct {
	ElectionID     int64             `json:"election_id"`
	OnlineEnabled  bool              `json:"online_enabled"`
	TPSEnabled     bool              `json:"tps_enabled"`
	OnlineSettings OnlineSettingsDTO `json:"online_settings"`
	TPSSettings    TPSSettingsDTO    `json:"tps_settings"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ModeSettingsRequest struct {
	OnlineEnabled  *bool               `json:"online_enabled,omitempty"`
	TPSEnabled     *bool               `json:"tps_enabled,omitempty"`
	OnlineSettings *OnlineSettingsBody `json:"online_settings,omitempty"`
	TPSSettings    *TPSSettingsBody    `json:"tps_settings,omitempty"`
}

type OnlineSettingsDTO struct {
	LoginURL            *string `json:"login_url,omitempty"`
	MaxSessionsPerVoter *int    `json:"max_sessions_per_voter,omitempty"`
}

type TPSSettingsDTO struct {
	RequireCheckin  *bool `json:"require_checkin,omitempty"`
	RequireBallotQR *bool `json:"require_ballot_qr,omitempty"`
	MaxTPS          *int  `json:"max_tps,omitempty"`
}

type OnlineSettingsBody struct {
	LoginURL            *string `json:"login_url,omitempty"`
	MaxSessionsPerVoter *int    `json:"max_sessions_per_voter,omitempty"`
}

type TPSSettingsBody struct {
	RequireCheckin  *bool `json:"require_checkin,omitempty"`
	RequireBallotQR *bool `json:"require_ballot_qr,omitempty"`
	MaxTPS          *int  `json:"max_tps,omitempty"`
}

type ElectionSummaryDTO struct {
	ElectionID   int64          `json:"election_id"`
	Status       ElectionStatus `json:"status"`
	CurrentPhase string         `json:"current_phase"`
	Candidates   CandidateStats `json:"candidates"`
	DPT          DPTStats       `json:"dpt"`
	TPS          TPSStats       `json:"tps"`
	Votes        VoteStats      `json:"votes"`
}

type CandidateStats struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
}

type DPTStats struct {
	TotalVoters  int64 `json:"total_voters"`
	OnlineVoters int64 `json:"online_voters"`
	TPSVoters    int64 `json:"tps_voters"`
}

type TPSStats struct {
	TotalTPS  int64 `json:"total_tps"`
	ActiveTPS int64 `json:"active_tps"`
}

type VoteStats struct {
	TotalCast  int64 `json:"total_cast"`
	OnlineCast int64 `json:"online_cast"`
	TPSCast    int64 `json:"tps_cast"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}
