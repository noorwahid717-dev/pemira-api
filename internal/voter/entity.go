package voter

import (
	"time"
	"pemira-api/internal/shared/constants"
)

type Voter struct {
	ID             int64       `json:"id"`
	NIM            string      `json:"nim"`
	FullName       string      `json:"full_name"`
	Faculty        string      `json:"faculty"`
	StudyProgram   string      `json:"study_program"`
	Batch          int         `json:"batch"`
	DateOfBirth    time.Time   `json:"date_of_birth"`
	IsActive       bool        `json:"is_active"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type VoterElectionStatus struct {
	ID         int64                `json:"id"`
	VoterID    int64                `json:"voter_id"`
	ElectionID int64                `json:"election_id"`
	Status     constants.VoterStatus `json:"status"`
	HasVoted   bool                 `json:"has_voted"`
	VotedAt    *time.Time           `json:"voted_at"`
	VotedVia   *string              `json:"voted_via"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}
