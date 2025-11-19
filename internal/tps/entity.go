package tps

import "time"

type TPS struct {
	ID          int64     `json:"id"`
	ElectionID  int64     `json:"election_id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	QRCode      string    `json:"qr_code"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TPSOperator struct {
	ID        int64     `json:"id"`
	TPSID     int64     `json:"tps_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TPSCheckin struct {
	ID         int64      `json:"id"`
	TPSID      int64      `json:"tps_id"`
	VoterID    int64      `json:"voter_id"`
	Status     string     `json:"status"`
	CheckedInAt time.Time `json:"checked_in_at"`
	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *int64     `json:"approved_by"`
}
