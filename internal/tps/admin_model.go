package tps

import "time"

type TPSDTO struct {
	ID           int64     `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	Capacity     int       `json:"capacity"`
	IsActive     bool      `json:"is_active"`
	OpenTime     *string   `json:"open_time,omitempty"`
	CloseTime    *string   `json:"close_time,omitempty"`
	PICName      *string   `json:"pic_name,omitempty"`
	PICPhone     *string   `json:"pic_phone,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	HasActiveQR  bool      `json:"has_active_qr"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TPSCreateRequest struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Location  string  `json:"location"`
	Capacity  int     `json:"capacity"`
	OpenTime  *string `json:"open_time,omitempty"`
	CloseTime *string `json:"close_time,omitempty"`
	PICName   *string `json:"pic_name,omitempty"`
	PICPhone  *string `json:"pic_phone,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type TPSUpdateRequest struct {
	Code      *string `json:"code,omitempty"`
	Name      *string `json:"name,omitempty"`
	Location  *string `json:"location,omitempty"`
	Capacity  *int    `json:"capacity,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	OpenTime  *string `json:"open_time,omitempty"`
	CloseTime *string `json:"close_time,omitempty"`
	PICName   *string `json:"pic_name,omitempty"`
	PICPhone  *string `json:"pic_phone,omitempty"`
	Notes     *string `json:"notes,omitempty"`
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
	OpenTime         *string    `json:"open_time,omitempty"`
	CloseTime        *string    `json:"close_time,omitempty"`
	PICName          *string    `json:"pic_name,omitempty"`
	PICPhone         *string    `json:"pic_phone,omitempty"`
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

type TPSQRMetadataResponse struct {
	TPSID     int64        `json:"tps_id"`
	Code      string       `json:"code"`
	Name      string       `json:"name"`
	ActiveQR  *ActiveQRDTO `json:"active_qr,omitempty"`
}

type ActiveQRDTO struct {
	ID        int64     `json:"id"`
	QRToken   string    `json:"qr_token"`
	CreatedAt time.Time `json:"created_at"`
}

type TPSQRRotateResponse struct {
	TPSID     int64        `json:"tps_id"`
	Code      string       `json:"code"`
	Name      string       `json:"name"`
	ActiveQR  ActiveQRDTO  `json:"active_qr"`
}

type TPSQRPrintResponse struct {
	TPSID     int64  `json:"tps_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	QRPayload string `json:"qr_payload"`
}
