package election

import (
	"time"
	"pemira-api/internal/shared/constants"
)

type Election struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Year        int                    `json:"year"`
	VotingMode  constants.VotingMode   `json:"voting_mode"`
	StartDate   time.Time              `json:"start_date"`
	EndDate     time.Time              `json:"end_date"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ElectionPhaseSchedule struct {
	ID         int64                  `json:"id"`
	ElectionID int64                  `json:"election_id"`
	Phase      constants.ElectionPhase `json:"phase"`
	StartDate  time.Time              `json:"start_date"`
	EndDate    time.Time              `json:"end_date"`
	CreatedAt  time.Time              `json:"created_at"`
}
