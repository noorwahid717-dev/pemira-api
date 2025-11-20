package tps

import "time"

type TPSDTO struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Capacity  int       `json:"capacity"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TPSCreateRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Capacity int    `json:"capacity"`
}

type TPSUpdateRequest struct {
	Code     *string `json:"code,omitempty"`
	Name     *string `json:"name,omitempty"`
	Location *string `json:"location,omitempty"`
	Capacity *int    `json:"capacity,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type TPSOperatorDTO struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

type TPSMonitorDTO struct {
	TPSID            int64      `json:"tps_id"`
	Code             string     `json:"code"`
	Name             string     `json:"name"`
	Location         string     `json:"location"`
	TotalCheckins    int64      `json:"total_checkins"`
	ApprovedCheckins int64      `json:"approved_checkins"`
	TotalVotes       int64      `json:"total_votes"`
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty"`
}

type CreateOperatorRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}
