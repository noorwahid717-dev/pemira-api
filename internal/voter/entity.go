package voter

import (
	"time"
	"pemira-api/internal/shared/constants"
)

type Voter struct {
	ID                      int64       `json:"id"`
	NIM                     string      `json:"nim"`
	Name                    string      `json:"name"`
	Email                   *string     `json:"email"`
	Phone                   *string     `json:"phone"`
	FacultyCode             *string     `json:"faculty_code"`
	FacultyName             *string     `json:"faculty_name"`
	StudyProgramCode        *string     `json:"study_program_code"`
	StudyProgramName        *string     `json:"study_program_name"`
	CohortYear              *int        `json:"cohort_year"`
	ClassLabel              *string     `json:"class_label"`
	PhotoURL                *string     `json:"photo_url"`
	Bio                     *string     `json:"bio"`
	VotingMethodPreference  *string     `json:"voting_method_preference"`
	AcademicStatus          string      `json:"academic_status"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
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
