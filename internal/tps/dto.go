package tps

type CreateTPSRequest struct {
	ElectionID int64  `json:"election_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Location   string `json:"location" validate:"required"`
}

type CheckinRequest struct {
	TPSID   int64 `json:"tps_id" validate:"required"`
	VoterID int64 `json:"voter_id" validate:"required"`
}

type ApproveCheckinRequest struct {
	CheckinID int64 `json:"checkin_id" validate:"required"`
}
