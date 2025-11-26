package voter

import "time"

type CreateVoterRequest struct {
	NIM          string    `json:"nim" validate:"required"`
	FullName     string    `json:"full_name" validate:"required"`
	Faculty      string    `json:"faculty" validate:"required"`
	StudyProgram string    `json:"study_program" validate:"required"`
	Batch        int       `json:"batch" validate:"required"`
	DateOfBirth  time.Time `json:"date_of_birth" validate:"required"`
}

type ImportDPTRequest struct {
	ElectionID int64  `json:"election_id" validate:"required"`
	FileData   []byte `json:"-"`
}

type VoterStatusResponse struct {
	Voter  *Voter               `json:"voter"`
	Status *VoterElectionStatus `json:"status"`
}

// Profile DTOs
type CompleteProfileResponse struct {
	PersonalInfo  PersonalInfo      `json:"personal_info"`
	VotingInfo    VotingInfo        `json:"voting_info"`
	Participation ParticipationInfo `json:"participation"`
	AccountInfo   AccountInfo       `json:"account_info"`
}

type PersonalInfo struct {
	VoterID         int64   `json:"voter_id"`
	Name            string  `json:"name"`
	Username        string  `json:"username"`
	Email           *string `json:"email"`
	Phone           *string `json:"phone"`
	FacultyName     *string `json:"faculty_name"`
	StudyProgramName *string `json:"study_program_name"`
	CohortYear      *int    `json:"cohort_year"`
	Semester        string  `json:"semester"`
	PhotoURL        *string `json:"photo_url"`
	VoterType       string  `json:"voter_type"`
}

type VotingInfo struct {
	PreferredMethod *string    `json:"preferred_method"`
	HasVoted        bool       `json:"has_voted"`
	VotedAt         *time.Time `json:"voted_at"`
	TPSName         *string    `json:"tps_name"`
	TPSLocation     *string    `json:"tps_location"`
}

type ParticipationInfo struct {
	TotalElections        int     `json:"total_elections"`
	ParticipatedElections int     `json:"participated_elections"`
	ParticipationRate     float64 `json:"participation_rate"`
	LastParticipation     *time.Time `json:"last_participation"`
}

type AccountInfo struct {
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login"`
	LoginCount    int        `json:"login_count"`
	AccountStatus string     `json:"account_status"`
}

type UpdateProfileRequest struct {
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	PhotoURL *string `json:"photo_url"`
}

type UpdateVotingMethodRequest struct {
	ElectionID      int64  `json:"election_id" validate:"required"`
	PreferredMethod string `json:"preferred_method" validate:"required,oneof=ONLINE TPS"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type ParticipationStatsResponse struct {
	Summary   ParticipationSummary   `json:"summary"`
	Elections []ElectionParticipation `json:"elections"`
}

type ParticipationSummary struct {
	TotalElections    int     `json:"total_elections"`
	Participated      int     `json:"participated"`
	NotParticipated   int     `json:"not_participated"`
	ParticipationRate float64 `json:"participation_rate"`
}

type ElectionParticipation struct {
	ElectionID   int64      `json:"election_id"`
	ElectionName string     `json:"election_name"`
	Year         int        `json:"year"`
	Voted        bool       `json:"voted"`
	VotedAt      *time.Time `json:"voted_at"`
	Method       string     `json:"method"`
}
