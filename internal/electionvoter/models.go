package electionvoter

import "time"

type VoterSummary struct {
	ID               int64   `json:"id"`
	NIM              string  `json:"nim"`
	Name             string  `json:"name"`
	VoterType        string  `json:"voter_type"`
	Email            *string `json:"email,omitempty"`
	FacultyCode      *string `json:"faculty_code,omitempty"`
	StudyProgramCode *string `json:"study_program_code,omitempty"`
	CohortYear       *int    `json:"cohort_year,omitempty"`
	AcademicStatus   *string `json:"academic_status,omitempty"`
	HasAccount       bool    `json:"has_account"`
	LecturerID       *int64  `json:"lecturer_id,omitempty"`
	StaffID          *int64  `json:"staff_id,omitempty"`
	VotingMethod     *string `json:"voting_method,omitempty"`
}

type ElectionVoter struct {
	ID              int64      `json:"election_voter_id"`
	ElectionID      int64      `json:"election_id"`
	VoterID         int64      `json:"voter_id"`
	NIM             string     `json:"nim"`
	Status          string     `json:"status"`
	VotingMethod    string     `json:"voting_method"`
	TPSID           *int64     `json:"tps_id,omitempty"`
	CheckedInAt     *time.Time `json:"checked_in_at,omitempty"`
	VotedAt         *time.Time `json:"voted_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
	VoterType       string     `json:"voter_type,omitempty"`
	Name            string     `json:"name,omitempty"`
	Email           *string    `json:"email,omitempty"`
	FacultyCode     *string    `json:"faculty_code,omitempty"`
	FacultyName     *string    `json:"faculty_name,omitempty"`
	StudyProgram    *string    `json:"study_program_code,omitempty"`
	StudyProgramName *string   `json:"study_program_name,omitempty"`
	CohortYear      *int       `json:"cohort_year,omitempty"`
	Semester        *int       `json:"semester,omitempty"`
	AcademicStatus  *string    `json:"academic_status,omitempty"`
	HasVoted        *bool      `json:"has_voted,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

type LookupResult struct {
	Voter         *VoterSummary  `json:"voter"`
	ElectionVoter *ElectionVoter `json:"election_voter,omitempty"`
}

type UpsertAndEnrollInput struct {
	VoterType              string  `json:"voter_type"`
	NIM                    string  `json:"nim"`
	Name                   string  `json:"name"`
	Email                  *string `json:"email,omitempty"`
	Phone                  *string `json:"phone,omitempty"`
	FacultyCode            *string `json:"faculty_code,omitempty"`
	FacultyName            *string `json:"faculty_name,omitempty"`
	StudyProgramCode       *string `json:"study_program_code,omitempty"`
	StudyProgramName       *string `json:"study_program_name,omitempty"`
	CohortYear             *int    `json:"cohort_year,omitempty"`
	Semester               *int    `json:"semester,omitempty"`
	AcademicStatus         *string `json:"academic_status,omitempty"`
	LecturerID             *int64  `json:"lecturer_id,omitempty"`
	StaffID                *int64  `json:"staff_id,omitempty"`
	VotingMethod           string  `json:"voting_method"`
	Status                 string  `json:"status"`
	TPSID                  *int64  `json:"tps_id,omitempty"`
}

type UpsertAndEnrollResult struct {
	VoterID             int64  `json:"voter_id"`
	ElectionVoterID     int64  `json:"election_voter_id,omitempty"`
	Status              string `json:"status,omitempty"`
	VotingMethod        string `json:"voting_method,omitempty"`
	TPSID               *int64 `json:"tps_id,omitempty"`
	CreatedVoter        bool   `json:"created_voter"`
	CreatedEnrollment   bool   `json:"created_election_voter"`
	DuplicateInElection bool   `json:"duplicate_in_election"`
}

type ListFilter struct {
	Search           string
	VoterType        string
	Status           string
	VotingMethod     string
	FacultyCode      string
	StudyProgramCode string
	CohortYear       *int
	TPSID            *int64
}

type UpdateInput struct {
	Status       *string `json:"status,omitempty"`
	VotingMethod *string `json:"voting_method,omitempty"`
	TPSID        *int64  `json:"tps_id,omitempty"`
	Semester     *int    `json:"semester,omitempty"`
}

type SelfRegisterInput struct {
	VotingMethod string `json:"voting_method"`
	TPSID        *int64 `json:"tps_id,omitempty"`
}
