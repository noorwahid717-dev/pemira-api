package election

import "time"

type ElectionStatus string

const (
	ElectionStatusDraft            ElectionStatus = "DRAFT"
	ElectionStatusRegistration     ElectionStatus = "REGISTRATION"
	ElectionStatusRegistrationOpen ElectionStatus = "REGISTRATION_OPEN"
	ElectionStatusVerification     ElectionStatus = "VERIFICATION"
	ElectionStatusCampaign         ElectionStatus = "CAMPAIGN"
	ElectionStatusQuietPeriod      ElectionStatus = "QUIET_PERIOD"
	ElectionStatusVotingOpen       ElectionStatus = "VOTING_OPEN"
	ElectionStatusVotingClosed     ElectionStatus = "VOTING_CLOSED"
	ElectionStatusClosed           ElectionStatus = "CLOSED"
	ElectionStatusRecap            ElectionStatus = "RECAP"
	ElectionStatusArchived         ElectionStatus = "ARCHIVED"
)

type Election struct {
	ID            int64          `json:"id"`
	Year          int            `json:"year"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Status        ElectionStatus `json:"status"`
	VotingStartAt *time.Time     `json:"voting_start_at,omitempty"`
	VotingEndAt   *time.Time     `json:"voting_end_at,omitempty"`
	OnlineEnabled bool           `json:"online_enabled"`
	TPSEnabled    bool           `json:"tps_enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type CurrentElectionDTO struct {
	ID            int64          `json:"id"`
	Year          int            `json:"year"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Status        ElectionStatus `json:"status"`
	VotingStartAt *time.Time     `json:"voting_start_at,omitempty"`
	VotingEndAt   *time.Time     `json:"voting_end_at,omitempty"`
	OnlineEnabled bool           `json:"online_enabled"`
	TPSEnabled    bool           `json:"tps_enabled"`
}
