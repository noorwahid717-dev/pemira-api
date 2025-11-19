package tps

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupCheckinSystem menginisialisasi sistem check-in TPS
func SetupCheckinSystem(ctx context.Context, dbURL string) (*CheckinHandler, error) {
	// 1. Setup database connection pool
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	// Configure pool
	poolConfig.MaxConns = 50
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 0 // connections live forever
	poolConfig.MaxConnIdleTime = 0 // connections don't close when idle

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("Database connection pool created successfully")

	// 2. Create service
	checkinService := NewCheckinService(db)

	// 3. Create handler
	handler := NewCheckinHandler(checkinService)

	return handler, nil
}

// Example: Full application setup
func ExampleFullSetup() {
	ctx := context.Background()

	// Setup check-in system
	dbURL := "postgres://user:pass@localhost:5432/pemira?sslmode=disable"
	checkinHandler, err := SetupCheckinSystem(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// Setup router
	r := mux.NewRouter()

	// Register routes
	checkinHandler.RegisterRoutes(r)

	// Add middlewares
	r.Use(loggingMiddleware)
	r.Use(authMiddleware)
	r.Use(corsMiddleware)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

// Example middlewares

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT token from header
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate token and extract user/voter ID
		// For voter endpoints: set voter_id
		// For operator endpoints: set user_id
		
		// Example:
		ctx := context.WithValue(r.Context(), "voter_id", int64(123))
		ctx = context.WithValue(ctx, "user_id", int64(456))
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Example: Integration test
func ExampleIntegrationTest() {
	ctx := context.Background()

	// Setup test database
	testDB, err := pgxpool.New(ctx, "postgres://test:test@localhost:5433/pemira_test")
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()

	// Create service
	service := NewCheckinService(testDB)

	// Test CheckinScan
	result, err := service.CheckinScan(ctx, 123, "PEMIRA|TPS01|abc123")
	if err != nil {
		log.Printf("CheckinScan failed: %v", err)
		return
	}

	log.Printf("CheckinScan success: checkin_id=%d, status=%s", result.CheckinID, result.Status)

	// Test ApproveCheckin
	approveResult, err := service.ApproveCheckin(ctx, 456, 1, result.CheckinID)
	if err != nil {
		log.Printf("ApproveCheckin failed: %v", err)
		return
	}

	log.Printf("ApproveCheckin success: status=%s, expires_at=%v",
		approveResult.Status, approveResult.ApprovedAt)
}

// Example: Usage in main.go
func ExampleMainIntegration() {
	// In your main.go:
	
	/*
	package main

	import (
		"context"
		"log"
		"net/http"
		"os"

		"github.com/gorilla/mux"
		"pemira-api/internal/tps"
		"pemira-api/internal/voting"
	)

	func main() {
		ctx := context.Background()

		// Get DB URL from env
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			log.Fatal("DATABASE_URL is required")
		}

		// Setup TPS check-in system
		checkinHandler, err := tps.SetupCheckinSystem(ctx, dbURL)
		if err != nil {
			log.Fatal("Failed to setup check-in system:", err)
		}

		// Setup voting system
		votingHandler, err := voting.SetupVotingSystem(ctx, dbURL)
		if err != nil {
			log.Fatal("Failed to setup voting system:", err)
		}

		// Setup router
		r := mux.NewRouter()

		// Register routes
		checkinHandler.RegisterRoutes(r)
		votingHandler.RegisterRoutes(r)

		// Health check
		r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}).Methods("GET")

		// Start server
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		log.Printf("Server starting on port %s", port)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatal(err)
		}
	}
	*/
}

// Example: Background job to expire check-ins
func ExampleExpireCheckinJob(db *pgxpool.Pool) {
	ctx := context.Background()

	// Run every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Expire check-ins that passed the expires_at time
		query := `
			UPDATE tps_checkins
			SET status = $1, updated_at = NOW()
			WHERE status = $2 AND expires_at < NOW()
		`

		result, err := db.Exec(ctx, query, CheckinStatusExpired, CheckinStatusApproved)
		if err != nil {
			log.Printf("Failed to expire check-ins: %v", err)
			continue
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("Expired %d check-ins", rowsAffected)
		}
	}
}

// Example: WebSocket integration
func ExampleWebSocketIntegration(wsHub interface{}) {
	// After ApproveCheckin success
	
	/*
	// Notify voter
	wsHub.PublishToVoter(voterID, map[string]interface{}{
		"type": "CHECKIN_APPROVED",
		"data": map[string]interface{}{
			"checkin_id": checkinID,
			"tps_id":     tpsID,
			"expires_at": expiresAt.Unix(),
		},
	})

	// Notify TPS panel
	wsHub.PublishToTPS(tpsID, map[string]interface{}{
		"type": "CHECKIN_UPDATED",
		"data": map[string]interface{}{
			"checkin_id": checkinID,
			"status":     "APPROVED",
		},
	})
	*/
}
