package voting

type CastVoteRequest struct {
	ElectionID  int64 `json:"election_id" validate:"required"`
	CandidateID int64 `json:"candidate_id" validate:"required"`
}

type CastVoteResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type LiveCountResponse struct {
	ElectionID int64            `json:"election_id"`
	Counts     map[int64]int64  `json:"counts"`
}
