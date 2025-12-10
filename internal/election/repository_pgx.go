package election

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

func NewRepository(db *pgxpool.Pool) Repository {
	return NewPgRepository(db)
}

var (
	ErrElectionNotFound    = fmt.Errorf("election not found")
	ErrVoterStatusNotFound = fmt.Errorf("voter status not found")
)

// GetCurrentElection returns the election with VOTING_OPEN status
// Priority changed: REGISTRATION → CAMPAIGN → VOTING_OPEN
func (r *PgRepository) GetCurrentElection(ctx context.Context) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    slug,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM myschema.elections
WHERE status IN ('REGISTRATION', 'CAMPAIGN', 'VOTING_OPEN')
ORDER BY 
    CASE status
        WHEN 'REGISTRATION' THEN 1
        WHEN 'CAMPAIGN' THEN 2
        WHEN 'VOTING_OPEN' THEN 3
    END,
    id DESC
LIMIT 1
`
	var e Election
	err := r.db.QueryRow(ctx, q).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

// GetCurrentForRegistration returns the election currently accepting registrations
// Priority: REGISTRATION → CAMPAIGN
func (r *PgRepository) GetCurrentForRegistration(ctx context.Context) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    slug,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM myschema.elections
WHERE status IN ('REGISTRATION', 'CAMPAIGN')
ORDER BY 
    CASE status
        WHEN 'REGISTRATION' THEN 1
        WHEN 'CAMPAIGN' THEN 2
    END,
    id DESC
LIMIT 1
`
	var e Election
	err := r.db.QueryRow(ctx, q).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

// GetActiveElection returns election by ID from settings.active_election_id
// This is the single source of truth for "active election"
func (r *PgRepository) GetActiveElection(ctx context.Context, settingsElectionID int) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    slug,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM myschema.elections
WHERE id = $1
`
	var e Election
	err := r.db.QueryRow(ctx, q, settingsElectionID).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

// IsRegistrationAllowed validates if election is accepting new registrations
func (r *PgRepository) IsRegistrationAllowed(ctx context.Context, election *Election) (bool, string) {
	if election == nil {
		return false, "Election not found"
	}

	// Check status - allow REGISTRATION, CAMPAIGN
	// Note: REGISTRATION_OPEN is in Go constants but not in DB enum yet
	allowedStatuses := map[ElectionStatus]bool{
		ElectionStatusRegistration: true,
		ElectionStatusCampaign:     true,
	}

	if !allowedStatuses[election.Status] {
		return false, "Masa pendaftaran sudah ditutup"
	}

	// TODO: Add date-based validation if registration_start_at and registration_end_at exist
	// if election.RegistrationStartAt != nil && election.RegistrationEndAt != nil {
	//     now := time.Now()
	//     if now.Before(*election.RegistrationStartAt) {
	//         return false, "Pendaftaran belum dibuka"
	//     }
	//     if now.After(*election.RegistrationEndAt) {
	//         return false, "Masa pendaftaran sudah ditutup"
	//     }
	// }

	return true, ""
}

func (r *PgRepository) ListPublicElections(ctx context.Context) ([]Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    slug,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM myschema.elections
WHERE status NOT IN ('ARCHIVED')
ORDER BY year DESC, id DESC
LIMIT 10
`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elections []Election
	for rows.Next() {
		var e Election
		err := rows.Scan(
			&e.ID,
			&e.Year,
			&e.Name,
			&e.Slug,
			&e.Status,
			&e.VotingStartAt,
			&e.VotingEndAt,
			&e.OnlineEnabled,
			&e.TPSEnabled,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		elections = append(elections, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return elections, nil
}

func (r *PgRepository) GetByID(ctx context.Context, id int64) (*Election, error) {
	const q = `
SELECT
    id,
    year,
    name,
    slug,
    status,
    voting_start_at,
    voting_end_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM myschema.elections
WHERE id = $1
`
	var e Election
	err := r.db.QueryRow(ctx, q, id).Scan(
		&e.ID,
		&e.Year,
		&e.Name,
		&e.Slug,
		&e.Status,
		&e.VotingStartAt,
		&e.VotingEndAt,
		&e.OnlineEnabled,
		&e.TPSEnabled,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *PgRepository) GetVoterStatus(
	ctx context.Context,
	electionID, voterID int64,
) (*MeStatusRow, error) {
	const q = `
SELECT
    vs.election_id,
    vs.voter_id,
    vs.is_eligible,
    vs.has_voted,
    vs.voted_at,
    vs.voting_method,
    vs.tps_id,
    e.online_enabled,
    e.tps_enabled,
    vs.preferred_method,
    vs.online_allowed,
    vs.tps_allowed
FROM voter_status vs
JOIN myschema.elections e
  ON e.id = vs.election_id
WHERE vs.election_id = $1
  AND vs.voter_id = $2
`
	var row MeStatusRow
	var method *string
	var preferred *string
	var onlineAllowed, tpsAllowed bool

	err := r.db.QueryRow(ctx, q, electionID, voterID).Scan(
		&row.ElectionID,
		&row.VoterID,
		&row.IsEligible,
		&row.HasVoted,
		&row.LastVoteAt,
		&method,
		&row.LastTPSID,
		&row.OnlineEnabled,
		&row.TPSEnabled,
		&preferred,
		&onlineAllowed,
		&tpsAllowed,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVoterStatusNotFound
		}
		return nil, err
	}

	row.LastVoteChannel = method
	row.PreferredMethod = preferred
	row.OnlineAllowed = row.OnlineEnabled && onlineAllowed
	row.TPSAllowed = row.TPSEnabled && tpsAllowed
	return &row, nil
}

func (r *PgRepository) GetHistory(ctx context.Context, electionID, voterID, userID int64) (*MeHistoryDTO, error) {
	h := &MeHistoryDTO{
		Voting:       []HistoryItem{},
		Checkins:     []HistoryItem{},
		Registration: []HistoryItem{},
		QR:           []HistoryItem{},
		Activities:   []HistoryItem{},
	}

	// Registration & voting info from voter_status
	var regCreatedAt, regUpdatedAt time.Time
	var votedAt *time.Time
	var method *string
	err := r.db.QueryRow(ctx, `
		SELECT created_at, updated_at, voted_at, voting_method
		FROM voter_status
		WHERE election_id = $1 AND voter_id = $2
	`, electionID, voterID).Scan(&regCreatedAt, &regUpdatedAt, &votedAt, &method)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return h, nil
		}
		return nil, err
	}
	h.Registration = append(h.Registration, HistoryItem{
		Type:      "REGISTRATION",
		Timestamp: regCreatedAt,
		Details:   "Terdaftar sebagai pemilih",
	})
	if votedAt != nil {
		detail := "Voting"
		if method != nil {
			detail = fmt.Sprintf("Voting via %s", strings.ToUpper(*method))
		}
		h.Voting = append(h.Voting, HistoryItem{
			Type:      "VOTING",
			Timestamp: *votedAt,
			Details:   detail,
		})
	}

	// Check-ins
	rows, err := r.db.Query(ctx, `
		SELECT status, scan_at, voted_at, tps_id
		FROM tps_checkins
		WHERE election_id = $1 AND voter_id = $2
		ORDER BY scan_at DESC
		LIMIT 20
	`, electionID, voterID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var scanAt time.Time
			var votedAtCheckin *time.Time
			var tpsID int64
			if err := rows.Scan(&status, &scanAt, &votedAtCheckin, &tpsID); err == nil {
				detail := fmt.Sprintf("Status %s (TPS %d)", status, tpsID)
				h.Checkins = append(h.Checkins, HistoryItem{
					Type:      "CHECKIN",
					Timestamp: scanAt,
					Details:   detail,
				})
				if votedAtCheckin != nil {
					h.Voting = append(h.Voting, HistoryItem{
						Type:      "VOTING",
						Timestamp: *votedAtCheckin,
						Details:   fmt.Sprintf("Voting via TPS %d", tpsID),
					})
				}
			}
		}
	}

	// QR history
	qrRows, err := r.db.Query(ctx, `
		SELECT qr_token, created_at, rotated_at, is_active
		FROM voter_tps_qr
		WHERE voter_id = $1 AND election_id = $2
		ORDER BY created_at DESC
		LIMIT 10
	`, voterID, electionID)
	if err == nil {
		defer qrRows.Close()
		for qrRows.Next() {
			var token string
			var createdAt time.Time
			var rotatedAt *time.Time
			var isActive bool
			if err := qrRows.Scan(&token, &createdAt, &rotatedAt, &isActive); err == nil {
				detail := "QR generated"
				if rotatedAt != nil {
					detail = "QR rotated"
					createdAt = *rotatedAt
				}
				if !isActive {
					detail += " (inactive)"
				}
				h.QR = append(h.QR, HistoryItem{
					Type:      "QR",
					Timestamp: createdAt,
					Details:   detail,
				})
			}
		}
	}

	// User activities (login/logout) from sessions
	sessRows, err := r.db.Query(ctx, `
		SELECT created_at, revoked_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err == nil {
		defer sessRows.Close()
		for sessRows.Next() {
			var createdAt time.Time
			var revokedAt *time.Time
			if err := sessRows.Scan(&createdAt, &revokedAt); err == nil {
				h.Activities = append(h.Activities, HistoryItem{
					Type:      "LOGIN",
					Timestamp: createdAt,
					Details:   "Login",
				})
				if revokedAt != nil {
					h.Activities = append(h.Activities, HistoryItem{
						Type:      "LOGOUT",
						Timestamp: *revokedAt,
						Details:   "Logout",
					})
				}
			}
		}
	}

	return h, nil
}
