package tps

import (
	"database/sql"
	"log"

	"github.com/go-chi/chi/v5"
)

// SetupTPSModule initializes TPS module with all dependencies
// This is an example - adjust to your application structure
func SetupTPSModule(db *sql.DB, router chi.Router) (*ServiceWithWebSocket, *WSHandler) {
	// Create repository
	repo := NewPostgresRepository(db)
	
	// Create WebSocket hub
	wsHub := NewWSHub()
	go wsHub.Run() // Start hub in background
	
	// Create service with WebSocket support
	service := NewServiceWithWebSocket(repo, wsHub)
	
	// Create HTTP handlers
	httpHandler := NewHandler(service.Service)
	wsHandler := NewWSHandler(wsHub, service.Service)
	
	// Register routes
	httpHandler.RegisterRoutes(router)
	wsHandler.RegisterRoutes(router)
	
	log.Println("TPS module initialized successfully")
	
	return service, wsHandler
}

// Example usage in main.go or router setup:
/*
func main() {
	// ... setup database, router, etc

	// Setup TPS module
	tpsService, tpsWSHandler := tps.SetupTPSModule(db, router)
	
	// TPS service is now available for use by other modules
	// Example: Voting module can use tpsService.MarkCheckinAsUsed()
	
	// Start HTTP server
	http.ListenAndServe(":8080", router)
}
*/

// Example integration with Voting module:
/*
// In voting module service
type VotingService struct {
	tpsService *tps.ServiceWithWebSocket
}

func (v *VotingService) CastTPSVote(ctx context.Context, voterID, electionID int64, ballot Ballot) error {
	// 1. Validate check-in
	checkin, err := v.tpsService.repo.GetCheckinByVoter(ctx, voterID, electionID)
	if err != nil || checkin == nil {
		return errors.New("no active check-in found")
	}
	
	if checkin.Status != tps.CheckinStatusApproved {
		return errors.New("check-in not approved")
	}
	
	if time.Now().After(*checkin.ExpiresAt) {
		return tps.ErrCheckinExpired
	}
	
	// 2. Process vote
	err = v.processVote(ctx, voterID, electionID, ballot)
	if err != nil {
		return err
	}
	
	// 3. Mark check-in as used
	err = v.tpsService.MarkCheckinAsUsed(ctx, checkin.ID)
	if err != nil {
		log.Printf("Failed to mark check-in as used: %v", err)
	}
	
	return nil
}
*/
