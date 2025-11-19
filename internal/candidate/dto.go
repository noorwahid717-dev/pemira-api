package candidate

type CreateCandidateRequest struct {
	ElectionID    int64  `json:"election_id" validate:"required"`
	OrderNumber   int    `json:"order_number" validate:"required"`
	Name          string `json:"name" validate:"required"`
	VisionMission string `json:"vision_mission"`
	PhotoURL      string `json:"photo_url"`
}

type UpdateCandidateRequest struct {
	Name          string `json:"name"`
	VisionMission string `json:"vision_mission"`
	PhotoURL      string `json:"photo_url"`
}
