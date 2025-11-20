package tps

import "time"

const (
	StatusDraft    = "DRAFT"
	StatusActive   = "ACTIVE"
	StatusClosed   = "CLOSED"
	
	CheckinStatusPending  = "PENDING"
	CheckinStatusApproved = "APPROVED"
	CheckinStatusRejected = "REJECTED"
	CheckinStatusUsed     = "USED"
	CheckinStatusExpired  = "EXPIRED"
	
	RoleKetuaTPS      = "KETUA_TPS"
	RoleOperatorPanel = "OPERATOR_PANEL"
)

type TPS struct {
	ID               int64      `json:"id"`
	ElectionID       int64      `json:"election_id"`
	Code             string     `json:"code"`
	Name             string     `json:"name"`
	Location         string     `json:"location"`
	Status           string     `json:"status"`
	VotingDate       *time.Time `json:"voting_date"`
	OpenTime         string     `json:"open_time"`
	CloseTime        string     `json:"close_time"`
	CapacityEstimate int        `json:"capacity_estimate"`
	AreaFacultyID    *int64     `json:"area_faculty_id"`
	PICName          *string    `json:"pic_name,omitempty"`
	PICPhone         *string    `json:"pic_phone,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TPSQR struct {
	ID         int64      `json:"id"`
	TPSID      int64      `json:"tps_id"`
	QRToken    string     `json:"qr_token"`
	IsActive   bool       `json:"is_active"`
	RotatedAt  *time.Time `json:"rotated_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TPSPanitia struct {
	ID        int64     `json:"id"`
	TPSID     int64     `json:"tps_id"`
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type TPSCheckin struct {
	ID              int64      `json:"id"`
	TPSID           int64      `json:"tps_id"`
	VoterID         int64      `json:"voter_id"`
	ElectionID      int64      `json:"election_id"`
	Status          string     `json:"status"`
	ScanAt          time.Time  `json:"scan_at"`
	ApprovedAt      *time.Time `json:"approved_at"`
	ApprovedByID    *int64     `json:"approved_by_id"`
	RejectionReason *string    `json:"rejection_reason"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TPSStats struct {
	TotalVotes        int `json:"total_votes"`
	TotalCheckins     int `json:"total_checkins"`
	PendingCheckins   int `json:"pending_checkins"`
	ApprovedCheckins  int `json:"approved_checkins"`
	RejectedCheckins  int `json:"rejected_checkins"`
}
