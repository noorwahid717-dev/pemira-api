package election

import "pemira-api/internal/shared/constants"

type CreateElectionRequest struct {
	Name        string                `json:"name" validate:"required"`
	Description string                `json:"description"`
	Year        int                   `json:"year" validate:"required"`
	VotingMode  constants.VotingMode  `json:"voting_mode" validate:"required"`
}

type UpdateElectionRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	VotingMode  constants.VotingMode  `json:"voting_mode"`
}
