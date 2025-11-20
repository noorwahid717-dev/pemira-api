package election

import "time"

type AdminElectionDTO struct {
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
	Year          *int    `json:"year,omitempty"`
	Name          *string `json:"name,omitempty"`
	Slug          *string `json:"slug,omitempty"`
	OnlineEnabled *bool   `json:"online_enabled,omitempty"`
	TPSEnabled    *bool   `json:"tps_enabled,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}
