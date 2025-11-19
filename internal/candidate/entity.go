package candidate

import "time"

type Candidate struct {
	ID          int64     `json:"id"`
	ElectionID  int64     `json:"election_id"`
	OrderNumber int       `json:"order_number"`
	Name        string    `json:"name"`
	VisionMission string  `json:"vision_mission"`
	PhotoURL    string    `json:"photo_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CandidateMember struct {
	ID          int64  `json:"id"`
	CandidateID int64  `json:"candidate_id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	PhotoURL    string `json:"photo_url"`
}
