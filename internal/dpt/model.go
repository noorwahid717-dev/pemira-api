package dpt

import "time"

type VoterWithStatusDTO struct {
	VoterID          int64  `json:"voter_id"`
	NIM              string `json:"nim"`
	Name             string `json:"name"`
	FacultyName      string `json:"faculty_name"`
	StudyProgramName string `json:"study_program_name"`
	Semester         string `json:"semester,omitempty"`
	ClassLabel       string `json:"class_label,omitempty"`
	CohortYear       *int   `json:"cohort_year,omitempty"`
	Email            string `json:"email"`
	HasAccount       bool   `json:"has_account"`
	VoterType        string `json:"voter_type,omitempty"`

	Status VoterStatusDTO `json:"status"`
}

type VoterStatusDTO struct {
	IsEligible      bool       `json:"is_eligible"`
	HasVoted        bool       `json:"has_voted"`
	LastVoteAt      *time.Time `json:"last_vote_at,omitempty"`
	VotingMethod    *string    `json:"voting_method,omitempty"`
	LastVoteChannel *string    `json:"last_vote_channel,omitempty"`
	LastTPSID       *int64     `json:"last_tps_id,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

type ListFilter struct {
	Faculty      string
	StudyProgram string
	CohortYear   *int
	HasVoted     *bool
	Eligible     *bool
	Search       string
	Limit        int
	Offset       int
}

type ImportRow struct {
	NIM          string
	Name         string
	FacultyName  string
	StudyProgram string
	CohortYear   int
	Email        string
	Phone        string
}

type ImportResult struct {
	TotalRows      int `json:"total_rows"`
	InsertedVoters int `json:"inserted_voters"`
	UpdatedVoters  int `json:"updated_voters"`
	CreatedStatus  int `json:"created_status"`
	SkippedStatus  int `json:"skipped_status"`
}

type VoterUpdateDTO struct {
	Name         *string `json:"name,omitempty"`
	FacultyName  *string `json:"faculty_name,omitempty"`
	StudyProgram *string `json:"study_program_name,omitempty"`
	CohortYear   *int    `json:"cohort_year,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	IsEligible   *bool   `json:"is_eligible,omitempty"`
	VoterType    *string `json:"voter_type,omitempty"`
	VotingMethod *string `json:"voting_method,omitempty"`
}
