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
