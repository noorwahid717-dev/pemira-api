package tps

import (
	"context"
	"strings"
	"time"
)

type PanelService struct {
	repo Repository
}

func NewPanelService(repo Repository) *PanelService {
	return &PanelService{repo: repo}
}

type PanelDashboard struct {
	ElectionID int64        `json:"election_id"`
	TPS        PanelTPSInfo `json:"tps"`
	Status     string       `json:"status"`
	Stats      PanelStats   `json:"stats"`
	LastActive *time.Time   `json:"last_activity_at,omitempty"`
}

type PanelTPSInfo struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type PanelStats struct {
	TotalRegistered int `json:"total_registered_tps_voters"`
	TotalCheckedIn  int `json:"total_checked_in"`
	TotalVoted      int `json:"total_voted"`
	TotalNotVoted   int `json:"total_not_voted"`
}

type PanelCheckinItem struct {
	CheckinID int64      `json:"checkin_id"`
	VoterID   int64      `json:"voter_id"`
	Name      string     `json:"name"`
	NIM       string     `json:"nim"`
	Faculty   string     `json:"faculty"`
	Program   string     `json:"program"`
	Status    string     `json:"status"`
	CheckinAt time.Time  `json:"checkin_time"`
	VotedAt   *time.Time `json:"voted_time,omitempty"`
}

type PanelCheckinDetail struct {
	CheckinID  int64      `json:"checkin_id"`
	ElectionID int64      `json:"election_id"`
	TPSID      int64      `json:"tps_id"`
	Voter      PanelVoter `json:"voter"`
	Status     string     `json:"status"`
	CheckinAt  time.Time  `json:"checkin_time"`
	VotedAt    *time.Time `json:"voted_time,omitempty"`
}

type PanelVoter struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	NIM     string `json:"nim"`
	Faculty string `json:"faculty"`
	Program string `json:"program"`
}

type TimelinePoint struct {
	Hour      string `json:"hour"`
	CheckedIn int    `json:"checked_in"`
	Voted     int    `json:"voted"`
}

func (s *PanelService) derivePanelStatus(tpsRow *TPS) string {
	now := time.Now()
	// Default by status field
	if tpsRow.Status == StatusClosed {
		return "CLOSED"
	}
	if tpsRow.Status != StatusActive {
		return "NOT_STARTED"
	}

	if tpsRow.VotingDate != nil {
		start := parseDailyTime(*tpsRow.VotingDate, tpsRow.OpenTime)
		end := parseDailyTime(*tpsRow.VotingDate, tpsRow.CloseTime)
		if start != nil && now.Before(*start) {
			return "NOT_STARTED"
		}
		if end != nil && now.After(*end) {
			return "CLOSED"
		}
	}
	return "OPEN"
}

func parseDailyTime(date time.Time, hhmm string) *time.Time {
	if strings.TrimSpace(hhmm) == "" {
		return nil
	}
	layout := "15:04"
	t, err := time.Parse(layout, hhmm)
	if err != nil {
		return nil
	}
	combined := time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
	return &combined
}

func mapCheckinStatus(status string) string {
	switch status {
	case CheckinStatusApproved:
		return "CHECKED_IN"
	case CheckinStatusVoted, "USED":
		return "VOTED"
	default:
		return "CHECKED_IN"
	}
}

func (s *PanelService) Dashboard(ctx context.Context, tpsID int64) (*PanelDashboard, error) {
	tpsRow, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, err
	}

	stats, err := s.repo.PanelDashboardStats(ctx, tpsID, tpsRow.ElectionID)
	if err != nil {
		return nil, err
	}

	totalNotVoted := stats.TotalRegistered - stats.TotalVoted
	if totalNotVoted < 0 {
		totalNotVoted = 0
	}

	return &PanelDashboard{
		ElectionID: tpsRow.ElectionID,
		TPS: PanelTPSInfo{
			ID:   tpsRow.ID,
			Code: tpsRow.Code,
			Name: tpsRow.Name,
		},
		Status: s.derivePanelStatus(tpsRow),
		Stats: PanelStats{
			TotalRegistered: stats.TotalRegistered,
			TotalCheckedIn:  stats.TotalCheckedIn,
			TotalVoted:      stats.TotalVoted,
			TotalNotVoted:   totalNotVoted,
		},
		LastActive: stats.LastActivity,
	}, nil
}

func (s *PanelService) ListCheckins(ctx context.Context, tpsID int64, status, search string, limit, offset int) ([]PanelCheckinItem, int, error) {
	items, total, err := s.repo.PanelListCheckins(ctx, tpsID, status, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]PanelCheckinItem, 0, len(items))
	for _, row := range items {
		resp = append(resp, PanelCheckinItem{
			CheckinID: row.ID,
			VoterID:   row.VoterID,
			Name:      row.VoterName,
			NIM:       row.VoterNIM,
			Faculty:   row.Faculty,
			Program:   row.Program,
			Status:    mapCheckinStatus(row.Status),
			CheckinAt: row.ScanAt,
			VotedAt:   row.VotedAt,
		})
	}
	return resp, total, nil
}

func (s *PanelService) GetCheckin(ctx context.Context, id int64) (*PanelCheckinDetail, error) {
	row, err := s.repo.PanelGetCheckin(ctx, id)
	if err != nil {
		return nil, err
	}

	return &PanelCheckinDetail{
		CheckinID:  row.ID,
		ElectionID: row.ElectionID,
		TPSID:      row.TPSID,
		Voter: PanelVoter{
			ID:      row.VoterID,
			Name:    row.VoterName,
			NIM:     row.VoterNIM,
			Faculty: row.Faculty,
			Program: row.Program,
		},
		Status:    mapCheckinStatus(row.Status),
		CheckinAt: row.ScanAt,
		VotedAt:   row.VotedAt,
	}, nil
}

func (s *PanelService) Status(ctx context.Context, tpsID int64) (*TPSStatusResponse, error) {
	tpsRow, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, err
	}
	status := s.derivePanelStatus(tpsRow)

	resp := &TPSStatusResponse{
		ElectionID: tpsRow.ElectionID,
		TPSID:      tpsRow.ID,
		Status:     status,
		Now:        time.Now().UTC(),
		VotingWindow: VotingWindow{
			StartAt: nil,
			EndAt:   nil,
		},
	}

	if tpsRow.VotingDate != nil {
		start := parseDailyTime(*tpsRow.VotingDate, tpsRow.OpenTime)
		end := parseDailyTime(*tpsRow.VotingDate, tpsRow.CloseTime)
		resp.VotingWindow.StartAt = start
		resp.VotingWindow.EndAt = end
	}

	return resp, nil
}

type TPSStatusResponse struct {
	ElectionID   int64        `json:"election_id"`
	TPSID        int64        `json:"tps_id"`
	Status       string       `json:"status"`
	Now          time.Time    `json:"now"`
	VotingWindow VotingWindow `json:"voting_window"`
}

func (s *PanelService) Timeline(ctx context.Context, tpsID int64) ([]TimelinePoint, error) {
	points, err := s.repo.PanelTimeline(ctx, tpsID)
	if err != nil {
		return nil, err
	}
	resp := make([]TimelinePoint, 0, len(points))
	for _, p := range points {
		resp = append(resp, TimelinePoint{
			Hour:      p.BucketStart.Format(time.RFC3339),
			CheckedIn: p.CheckedIn,
			Voted:     p.Voted,
		})
	}
	return resp, nil
}

func (s *PanelService) ListTPSByElection(ctx context.Context, electionID int64) ([]PanelTPSListItem, error) {
	return s.repo.PanelListTPSByElection(ctx, electionID)
}

func (s *PanelService) EnsureAccess(ctx context.Context, electionID, tpsID int64, role string, tokenTPS *int64) (*TPS, error) {
	tpsRow, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, err
	}
	if tpsRow.ElectionID != electionID {
		return nil, ErrTPSMismatch
	}
	if role == "TPS_OPERATOR" {
		if tokenTPS == nil || *tokenTPS != tpsID {
			return nil, ErrTPSAccessDenied
		}
	}
	return tpsRow, nil
}

// CreateCheckinViaQR creates a check-in using registration QR payload, deriving election from TPS.
func (s *PanelService) CreateCheckinViaQR(ctx context.Context, tpsID int64, raw string) (*PanelCheckinRow, error) {
	tpsRow, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, err
	}

	reg, err := s.repo.ParseRegistrationCode(ctx, raw)
	if err != nil {
		return nil, ErrQRInvalid
	}

	if reg.ElectionID != tpsRow.ElectionID {
		return nil, ErrTPSMismatch
	}

	reg.TPSID = &tpsID
	return s.repo.CreatePanelCheckin(ctx, *reg)
}
