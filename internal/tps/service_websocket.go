package tps

import (
	"context"
)

// ServiceWithWebSocket extends Service with WebSocket broadcasting capabilities
type ServiceWithWebSocket struct {
	*Service
	wsHub *WSHub
}

func NewServiceWithWebSocket(repo Repository, wsHub *WSHub) *ServiceWithWebSocket {
	return &ServiceWithWebSocket{
		Service: NewService(repo),
		wsHub:   wsHub,
	}
}

// Override ScanQR to broadcast new check-in
func (s *ServiceWithWebSocket) ScanQR(ctx context.Context, voterID int64, req *ScanQRRequest) (*ScanQRResponse, error) {
	result, err := s.Service.ScanQR(ctx, voterID, req)
	if err != nil {
		return nil, err
	}
	
	// Broadcast new check-in to WebSocket clients
	if s.wsHub != nil && result.CheckinID > 0 {
		voterInfo, _ := s.repo.GetVoterInfo(ctx, voterID)
		if voterInfo != nil {
			s.wsHub.BroadcastCheckinNew(
				result.TPS.ID,
				result.CheckinID,
				voterInfo,
				result.ScanAt,
			)
		}
	}
	
	return result, nil
}

// Override ApproveCheckin to broadcast status update
func (s *ServiceWithWebSocket) ApproveCheckin(ctx context.Context, tpsID, checkinID, approverID int64) (*ApproveCheckinResponse, error) {
	result, err := s.Service.ApproveCheckin(ctx, tpsID, checkinID, approverID)
	if err != nil {
		return nil, err
	}
	
	// Broadcast approval to WebSocket clients
	if s.wsHub != nil {
		s.wsHub.BroadcastCheckinUpdated(tpsID, checkinID, CheckinStatusApproved)
	}
	
	return result, nil
}

// Override RejectCheckin to broadcast status update
func (s *ServiceWithWebSocket) RejectCheckin(ctx context.Context, tpsID, checkinID, approverID int64, reason string) (*RejectCheckinResponse, error) {
	result, err := s.Service.RejectCheckin(ctx, tpsID, checkinID, approverID, reason)
	if err != nil {
		return nil, err
	}
	
	// Broadcast rejection to WebSocket clients
	if s.wsHub != nil {
		s.wsHub.BroadcastCheckinUpdated(tpsID, checkinID, CheckinStatusRejected)
	}
	
	return result, nil
}

// MarkCheckinAsUsed is called by Voting module after successful vote
func (s *ServiceWithWebSocket) MarkCheckinAsUsed(ctx context.Context, checkinID int64) error {
	checkin, err := s.repo.GetCheckin(ctx, checkinID)
	if err != nil {
		return err
	}
	
	checkin.Status = CheckinStatusUsed
	err = s.repo.UpdateCheckin(ctx, checkin)
	if err != nil {
		return err
	}
	
	// Broadcast status update
	if s.wsHub != nil {
		s.wsHub.BroadcastCheckinUpdated(checkin.TPSID, checkinID, CheckinStatusUsed)
	}
	
	return nil
}

// GetWSHub returns the WebSocket hub for external use
func (s *ServiceWithWebSocket) GetWSHub() *WSHub {
	return s.wsHub
}
