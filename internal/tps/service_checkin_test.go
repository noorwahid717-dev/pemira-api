package tps

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Connect to test database
	dbURL := "postgres://test:test@localhost:5433/pemira_test?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	// Clean up tables
	_, _ = pool.Exec(context.Background(), "TRUNCATE tps_checkins CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE tps_qr CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE tps CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE voter_status CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE voters CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE elections CASCADE")

	return pool
}

func createTestElection(t *testing.T, pool *pgxpool.Pool) int64 {
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO elections (name, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW() + INTERVAL '7 days', NOW(), NOW())
		RETURNING id
	`, "Test Election", "VOTING_OPEN").Scan(&id)
	require.NoError(t, err)
	return id
}

func createTestTPS(t *testing.T, pool *pgxpool.Pool, electionID int64) (int64, string) {
	var id int64
	code := "TPS01"
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tps (election_id, code, name, location, status, voting_date, open_time, close_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), '08:00', '16:00', NOW(), NOW())
		RETURNING id
	`, electionID, code, "TPS Gedung A", "Gedung A Lt.1", StatusActive).Scan(&id)
	require.NoError(t, err)
	return id, code
}

func createTestQR(t *testing.T, pool *pgxpool.Pool, tpsID int64) (int64, string) {
	var id int64
	secret := "abc123def456"
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tps_qr (tps_id, qr_secret_suffix, is_active, created_at)
		VALUES ($1, $2, true, NOW())
		RETURNING id
	`, tpsID, secret).Scan(&id)
	require.NoError(t, err)
	return id, secret
}

func createTestVoter(t *testing.T, pool *pgxpool.Pool, electionID int64) int64 {
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO voters (election_id, nim, name, faculty, is_eligible, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id
	`, electionID, "1234567890", "Test Voter", "Teknik").Scan(&id)
	require.NoError(t, err)

	// Create voter_status
	_, err = pool.Exec(context.Background(), `
		INSERT INTO voter_status (election_id, voter_id, has_voted, created_at, updated_at)
		VALUES ($1, $2, false, NOW(), NOW())
	`, electionID, id)
	require.NoError(t, err)

	return id
}

func createTestPanitia(t *testing.T, pool *pgxpool.Pool, tpsID, userID int64) {
	_, err := pool.Exec(context.Background(), `
		INSERT INTO tps_panitia (tps_id, user_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
	`, tpsID, userID, RoleOperatorPanel)
	require.NoError(t, err)
}

// Unit Tests

func TestCheckinScan_Success(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)

	// Create QR payload
	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret

	// Execute
	result, err := service.CheckinScan(ctx, voterID, qrPayload)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.CheckinID, int64(0))
	assert.Equal(t, CheckinStatusPending, result.Status)
	assert.Equal(t, tpsID, result.TPS.ID)
	assert.Equal(t, tpsCode, result.TPS.Code)
	assert.Contains(t, result.Message, "menunggu verifikasi")
}

func TestCheckinScan_InvalidQR(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Invalid payload
	result, err := service.CheckinScan(ctx, 1, "INVALID|FORMAT")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrQRInvalid, err)
}

func TestCheckinScan_AlreadyVoted(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)

	// Mark voter as already voted
	_, err := pool.Exec(ctx, `
		UPDATE voter_status
		SET has_voted = true, voted_at = NOW()
		WHERE election_id = $1 AND voter_id = $2
	`, electionID, voterID)
	require.NoError(t, err)

	// Create QR payload
	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret

	// Execute
	result, err := service.CheckinScan(ctx, voterID, qrPayload)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrAlreadyVoted, err)
}

func TestCheckinScan_ExistingPending(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)

	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret

	// First scan
	result1, err := service.CheckinScan(ctx, voterID, qrPayload)
	assert.NoError(t, err)

	// Second scan (should return existing pending)
	result2, err := service.CheckinScan(ctx, voterID, qrPayload)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, result1.CheckinID, result2.CheckinID)
	assert.Equal(t, CheckinStatusPending, result2.Status)
}

func TestApproveCheckin_Success(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)
	operatorID := int64(999)
	createTestPanitia(t, pool, tpsID, operatorID)

	// Create pending check-in
	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret
	scanResult, err := service.CheckinScan(ctx, voterID, qrPayload)
	require.NoError(t, err)

	// Execute approve
	result, err := service.ApproveCheckin(ctx, operatorID, tpsID, scanResult.CheckinID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scanResult.CheckinID, result.CheckinID)
	assert.Equal(t, CheckinStatusApproved, result.Status)
	assert.Equal(t, voterID, result.Voter.ID)
	assert.Equal(t, tpsID, result.TPS.ID)
	assert.WithinDuration(t, time.Now().UTC(), result.ApprovedAt, 5*time.Second)
}

func TestApproveCheckin_AccessDenied(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)
	unauthorizedUserID := int64(888) // Not assigned to TPS

	// Create pending check-in
	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret
	scanResult, err := service.CheckinScan(ctx, voterID, qrPayload)
	require.NoError(t, err)

	// Execute approve with unauthorized user
	result, err := service.ApproveCheckin(ctx, unauthorizedUserID, tpsID, scanResult.CheckinID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrTPSAccessDenied, err)
}

func TestApproveCheckin_NotPending(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, _ := createTestTPS(t, pool, electionID)
	voterID := createTestVoter(t, pool, electionID)
	operatorID := int64(999)
	createTestPanitia(t, pool, tpsID, operatorID)

	// Create already approved check-in
	var checkinID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO tps_checkins (tps_id, voter_id, election_id, status, scan_at, approved_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW(), NOW())
		RETURNING id
	`, tpsID, voterID, electionID, CheckinStatusApproved).Scan(&checkinID)
	require.NoError(t, err)

	// Execute approve
	result, err := service.ApproveCheckin(ctx, operatorID, tpsID, checkinID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrCheckinNotPending, err)
}

func TestRejectCheckin_Success(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup test data
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)
	operatorID := int64(999)
	createTestPanitia(t, pool, tpsID, operatorID)

	// Create pending check-in
	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret
	scanResult, err := service.CheckinScan(ctx, voterID, qrPayload)
	require.NoError(t, err)

	// Execute reject
	reason := "Identitas tidak sesuai"
	result, err := service.RejectCheckin(ctx, operatorID, tpsID, scanResult.CheckinID, reason)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scanResult.CheckinID, result.CheckinID)
	assert.Equal(t, CheckinStatusRejected, result.Status)
	assert.Equal(t, reason, result.Reason)
}

func TestParseQRPayload_Valid(t *testing.T) {
	service := &CheckinService{}

	tpsCode, qrSecret, err := service.parseQRPayload("PEMIRA|TPS01|abc123")

	assert.NoError(t, err)
	assert.Equal(t, "TPS01", tpsCode)
	assert.Equal(t, "abc123", qrSecret)
}

func TestParseQRPayload_Invalid(t *testing.T) {
	service := &CheckinService{}

	tests := []struct {
		name    string
		payload string
	}{
		{"Wrong prefix", "WRONG|TPS01|abc123"},
		{"Missing parts", "PEMIRA|TPS01"},
		{"Empty", ""},
		{"Single pipe", "PEMIRA|"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.parseQRPayload(tt.payload)
			assert.Error(t, err)
			assert.Equal(t, ErrQRInvalid, err)
		})
	}
}

// Integration Tests

func TestFullCheckinFlow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)
	operatorID := int64(999)
	createTestPanitia(t, pool, tpsID, operatorID)

	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret

	// Step 1: Voter scans QR
	scanResult, err := service.CheckinScan(ctx, voterID, qrPayload)
	require.NoError(t, err)
	assert.Equal(t, CheckinStatusPending, scanResult.Status)

	// Step 2: Operator approves
	approveResult, err := service.ApproveCheckin(ctx, operatorID, tpsID, scanResult.CheckinID)
	require.NoError(t, err)
	assert.Equal(t, CheckinStatusApproved, approveResult.Status)

	// Step 3: Verify check-in is approved in database
	var status string
	var expiresAt time.Time
	err = pool.QueryRow(ctx, `
		SELECT status, expires_at
		FROM tps_checkins
		WHERE id = $1
	`, scanResult.CheckinID).Scan(&status, &expiresAt)
	require.NoError(t, err)
	assert.Equal(t, CheckinStatusApproved, status)
	assert.True(t, expiresAt.After(time.Now().UTC()))
}

func TestConcurrentCheckins_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	service := NewCheckinService(pool)
	ctx := context.Background()

	// Setup
	electionID := createTestElection(t, pool)
	tpsID, tpsCode := createTestTPS(t, pool, electionID)
	_, qrSecret := createTestQR(t, pool, tpsID)
	voterID := createTestVoter(t, pool, electionID)

	qrPayload := "PEMIRA|" + tpsCode + "|" + qrSecret

	// Concurrent scans
	done := make(chan bool, 5)
	results := make([]*ScanQRResponse, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			results[idx], errors[idx] = service.CheckinScan(ctx, voterID, qrPayload)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify: should only create one check-in (or all return same ID)
	successCount := 0
	var checkinID int64
	for i := 0; i < 5; i++ {
		if errors[i] == nil {
			successCount++
			if checkinID == 0 {
				checkinID = results[i].CheckinID
			} else {
				// All successful results should have same checkin ID
				assert.Equal(t, checkinID, results[i].CheckinID)
			}
		}
	}

	assert.Greater(t, successCount, 0, "At least one scan should succeed")

	// Verify only one check-in in database
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tps_checkins
		WHERE voter_id = $1 AND election_id = $2
	`, voterID, electionID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should only have one check-in record")
}
