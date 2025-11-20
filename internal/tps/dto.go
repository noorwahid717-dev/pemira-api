package tps

import "time"

// Admin TPS Management DTOs
type CreateTPSRequest struct {
	ElectionID       int64   `json:"election_id" validate:"required"`
	Code             string  `json:"code" validate:"required,min=3,max=20"`
	Name             string  `json:"name" validate:"required"`
	Location         string  `json:"location" validate:"required"`
	VotingDate       string  `json:"voting_date" validate:"required"`
	OpenTime         string  `json:"open_time" validate:"required"`
	CloseTime        string  `json:"close_time" validate:"required"`
	CapacityEstimate int     `json:"capacity_estimate" validate:"min=0"`
	Status           string  `json:"status" validate:"required,oneof=DRAFT ACTIVE CLOSED"`
	PICName          *string `json:"pic_name,omitempty"`
	PICPhone         *string `json:"pic_phone,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

type UpdateTPSRequest struct {
	Name             string  `json:"name" validate:"required"`
	Location         string  `json:"location" validate:"required"`
	VotingDate       string  `json:"voting_date" validate:"required"`
	OpenTime         string  `json:"open_time" validate:"required"`
	CloseTime        string  `json:"close_time" validate:"required"`
	CapacityEstimate int     `json:"capacity_estimate" validate:"min=0"`
	Status           string  `json:"status" validate:"required,oneof=DRAFT ACTIVE CLOSED"`
	PICName          *string `json:"pic_name,omitempty"`
	PICPhone         *string `json:"pic_phone,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

type AssignPanitiaRequest struct {
	Members []PanitiaMember `json:"members" validate:"required,dive"`
}

type PanitiaMember struct {
	UserID int64  `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=KETUA_TPS OPERATOR_PANEL"`
}

type TPSListResponse struct {
	Items      []TPSListItem   `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}

type TPSListItem struct {
	ID            int64   `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Location      string  `json:"location"`
	Status        string  `json:"status"`
	VotingDate    string  `json:"voting_date"`
	OpenTime      string  `json:"open_time"`
	CloseTime     string  `json:"close_time"`
	PICName       *string `json:"pic_name,omitempty"`
	PICPhone      *string `json:"pic_phone,omitempty"`
	HasActiveQR   bool    `json:"has_active_qr"`
	TotalVotes    int     `json:"total_votes"`
	TotalCheckins int     `json:"total_checkins"`
}

type TPSDetailResponse struct {
	ID               int64              `json:"id"`
	ElectionID       int64              `json:"election_id"`
	Code             string             `json:"code"`
	Name             string             `json:"name"`
	Location         string             `json:"location"`
	Status           string             `json:"status"`
	VotingDate       string             `json:"voting_date"`
	OpenTime         string             `json:"open_time"`
	CloseTime        string             `json:"close_time"`
	CapacityEstimate int                `json:"capacity_estimate"`
	PICName          *string            `json:"pic_name,omitempty"`
	PICPhone         *string            `json:"pic_phone,omitempty"`
	Notes            *string            `json:"notes,omitempty"`
	AreaFaculty      *FacultyInfo       `json:"area_faculty"`
	QR               *QRInfo            `json:"qr"`
	Stats            TPSStats           `json:"stats"`
	Panitia          []PanitiaInfo      `json:"panitia"`
}

type FacultyInfo struct {
	ID   *int64  `json:"id"`
	Name *string `json:"name"`
}

type QRInfo struct {
	ID        int64  `json:"id"`
	QRToken   string `json:"qr_token"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type PanitiaInfo struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type RegenerateQRResponse struct {
	TPSID int64       `json:"tps_id"`
	QR    QRPayload   `json:"qr"`
}

type QRPayload struct {
	ID        int64  `json:"id"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

// Student Check-in DTOs
type ScanQRRequest struct {
	QRPayload string `json:"qr_payload" validate:"required"`
}

type ScanQRResponse struct {
	CheckinID int64      `json:"checkin_id"`
	TPS       TPSInfo    `json:"tps"`
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	ScanAt    time.Time  `json:"scan_at,omitempty"`
}

type TPSInfo struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CheckinStatusResponse struct {
	HasActiveCheckin bool       `json:"has_active_checkin"`
	Status           *string    `json:"status,omitempty"`
	TPS              *TPSInfo   `json:"tps,omitempty"`
	ScanAt           *time.Time `json:"scan_at,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

// TPS Panel DTOs
type CheckinQueueResponse struct {
	Items []CheckinQueueItem `json:"items"`
}

type CheckinQueueItem struct {
	ID       int64      `json:"id"`
	Voter    VoterInfo  `json:"voter"`
	Status   string     `json:"status"`
	ScanAt   time.Time  `json:"scan_at"`
	HasVoted bool       `json:"has_voted"`
}

type VoterInfo struct {
	ID             int64  `json:"id"`
	NIM            string `json:"nim"`
	Name           string `json:"name"`
	Faculty        string `json:"faculty"`
	StudyProgram   string `json:"study_program"`
	CohortYear     int    `json:"cohort_year"`
	AcademicStatus string `json:"academic_status"`
}

type RejectCheckinRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ApproveCheckinResponse struct {
	CheckinID  int64     `json:"checkin_id"`
	Status     string    `json:"status"`
	Voter      VoterInfo `json:"voter"`
	TPS        TPSInfo   `json:"tps"`
	ApprovedAt time.Time `json:"approved_at"`
}

type RejectCheckinResponse struct {
	CheckinID int64  `json:"checkin_id"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}

type TPSSummaryResponse struct {
	ID         int64    `json:"id"`
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Location   string   `json:"location"`
	Status     string   `json:"status"`
	VotingDate string   `json:"voting_date"`
	OpenTime   string   `json:"open_time"`
	CloseTime  string   `json:"close_time"`
	Stats      TPSStats `json:"stats"`
}

type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}
