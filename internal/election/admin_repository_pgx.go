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

type PgAdminRepository struct {
	db *pgxpool.Pool
}

func NewPgAdminRepository(db *pgxpool.Pool) *PgAdminRepository {
	return &PgAdminRepository{db: db}
}

func (r *PgAdminRepository) ListElections(ctx context.Context, filter AdminElectionListFilter) ([]AdminElectionDTO, int64, error) {
	conditions := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Year != nil {
		conditions = append(conditions, fmt.Sprintf("year = $%d", argPos))
		args = append(args, *filter.Year)
		argPos++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM elections %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT
    id,
    year,
    name,
    code,
    status,
    registration_start_at,
    registration_end_at,
    verification_start_at,
    verification_end_at,
    campaign_start_at,
    campaign_end_at,
    quiet_start_at,
    quiet_end_at,
    voting_start_at,
    voting_end_at,
    recap_start_at,
    recap_end_at,
    announcement_at,
    finished_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM elections
%s
ORDER BY year DESC, id DESC
LIMIT $%d OFFSET $%d
`, whereClause, argPos, argPos+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []AdminElectionDTO{}
	for rows.Next() {
		var dto AdminElectionDTO
		err := rows.Scan(
			&dto.ID,
			&dto.Year,
			&dto.Name,
			&dto.Slug,
			&dto.Status,
			&dto.RegistrationStartAt,
			&dto.RegistrationEndAt,
			&dto.VerificationStartAt,
			&dto.VerificationEndAt,
			&dto.CampaignStartAt,
			&dto.CampaignEndAt,
			&dto.QuietStartAt,
			&dto.QuietEndAt,
			&dto.VotingStartAt,
			&dto.VotingEndAt,
			&dto.RecapStartAt,
			&dto.RecapEndAt,
			&dto.AnnouncementAt,
			&dto.FinishedAt,
			&dto.OnlineEnabled,
			&dto.TPSEnabled,
			&dto.CreatedAt,
			&dto.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, dto)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *PgAdminRepository) GetElectionByID(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	const q = `
SELECT
    id,
    year,
    name,
    code,
    status,
    registration_start_at,
    registration_end_at,
    verification_start_at,
    verification_end_at,
    campaign_start_at,
    campaign_end_at,
    quiet_start_at,
    quiet_end_at,
    voting_start_at,
    voting_end_at,
    recap_start_at,
    recap_end_at,
    announcement_at,
    finished_at,
    online_enabled,
    tps_enabled,
    created_at,
    updated_at
FROM elections
WHERE id = $1
`
	var dto AdminElectionDTO
	err := r.db.QueryRow(ctx, q, id).Scan(
		&dto.ID,
		&dto.Year,
		&dto.Name,
		&dto.Slug,
		&dto.Status,
		&dto.RegistrationStartAt,
		&dto.RegistrationEndAt,
		&dto.VerificationStartAt,
		&dto.VerificationEndAt,
		&dto.CampaignStartAt,
		&dto.CampaignEndAt,
		&dto.QuietStartAt,
		&dto.QuietEndAt,
		&dto.VotingStartAt,
		&dto.VotingEndAt,
		&dto.RecapStartAt,
		&dto.RecapEndAt,
		&dto.AnnouncementAt,
		&dto.FinishedAt,
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &dto, nil
}

func (r *PgAdminRepository) CreateElection(ctx context.Context, req AdminElectionCreateRequest) (*AdminElectionDTO, error) {
	const q = `
INSERT INTO elections (
    year,
    name,
    code,
    status,
    online_enabled,
    tps_enabled
) VALUES ($1, $2, $3, 'DRAFT', $4, $5)
RETURNING
    id, year, name, code, status,
    registration_start_at, registration_end_at,
    verification_start_at, verification_end_at,
    campaign_start_at, campaign_end_at,
    quiet_start_at, quiet_end_at,
    voting_start_at, voting_end_at,
    recap_start_at, recap_end_at,
    announcement_at, finished_at,
    online_enabled, tps_enabled,
    created_at, updated_at
`
	var dto AdminElectionDTO
	err := r.db.QueryRow(ctx, q,
		req.Year,
		req.Name,
		req.Slug,
		req.OnlineEnabled,
		req.TPSEnabled,
	).Scan(
		&dto.ID,
		&dto.Year,
		&dto.Name,
		&dto.Slug,
		&dto.Status,
		&dto.RegistrationStartAt,
		&dto.RegistrationEndAt,
		&dto.VerificationStartAt,
		&dto.VerificationEndAt,
		&dto.CampaignStartAt,
		&dto.CampaignEndAt,
		&dto.QuietStartAt,
		&dto.QuietEndAt,
		&dto.VotingStartAt,
		&dto.VotingEndAt,
		&dto.RecapStartAt,
		&dto.RecapEndAt,
		&dto.AnnouncementAt,
		&dto.FinishedAt,
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (r *PgAdminRepository) UpdateElection(ctx context.Context, id int64, req AdminElectionUpdateRequest) (*AdminElectionDTO, error) {
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Year != nil {
		updates = append(updates, fmt.Sprintf("year = $%d", argPos))
		args = append(args, *req.Year)
		argPos++
	}

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *req.Name)
		argPos++
	}

	if req.Slug != nil {
		updates = append(updates, fmt.Sprintf("code = $%d", argPos))
		args = append(args, *req.Slug)
		argPos++
	}

	if req.OnlineEnabled != nil {
		updates = append(updates, fmt.Sprintf("online_enabled = $%d", argPos))
		args = append(args, *req.OnlineEnabled)
		argPos++
	}

	if req.TPSEnabled != nil {
		updates = append(updates, fmt.Sprintf("tps_enabled = $%d", argPos))
		args = append(args, *req.TPSEnabled)
		argPos++
	}

	if req.RegistrationStartAt != nil {
		updates = append(updates, fmt.Sprintf("registration_start_at = $%d", argPos))
		args = append(args, *req.RegistrationStartAt)
		argPos++
	}
	if req.RegistrationEndAt != nil {
		updates = append(updates, fmt.Sprintf("registration_end_at = $%d", argPos))
		args = append(args, *req.RegistrationEndAt)
		argPos++
	}
	if req.VerificationStartAt != nil {
		updates = append(updates, fmt.Sprintf("verification_start_at = $%d", argPos))
		args = append(args, *req.VerificationStartAt)
		argPos++
	}
	if req.VerificationEndAt != nil {
		updates = append(updates, fmt.Sprintf("verification_end_at = $%d", argPos))
		args = append(args, *req.VerificationEndAt)
		argPos++
	}
	if req.CampaignStartAt != nil {
		updates = append(updates, fmt.Sprintf("campaign_start_at = $%d", argPos))
		args = append(args, *req.CampaignStartAt)
		argPos++
	}
	if req.CampaignEndAt != nil {
		updates = append(updates, fmt.Sprintf("campaign_end_at = $%d", argPos))
		args = append(args, *req.CampaignEndAt)
		argPos++
	}
	if req.QuietStartAt != nil {
		updates = append(updates, fmt.Sprintf("quiet_start_at = $%d", argPos))
		args = append(args, *req.QuietStartAt)
		argPos++
	}
	if req.QuietEndAt != nil {
		updates = append(updates, fmt.Sprintf("quiet_end_at = $%d", argPos))
		args = append(args, *req.QuietEndAt)
		argPos++
	}
	if req.VotingStartAt != nil {
		updates = append(updates, fmt.Sprintf("voting_start_at = $%d", argPos))
		args = append(args, *req.VotingStartAt)
		argPos++
	}
	if req.VotingEndAt != nil {
		updates = append(updates, fmt.Sprintf("voting_end_at = $%d", argPos))
		args = append(args, *req.VotingEndAt)
		argPos++
	}
	if req.RecapStartAt != nil {
		updates = append(updates, fmt.Sprintf("recap_start_at = $%d", argPos))
		args = append(args, *req.RecapStartAt)
		argPos++
	}
	if req.RecapEndAt != nil {
		updates = append(updates, fmt.Sprintf("recap_end_at = $%d", argPos))
		args = append(args, *req.RecapEndAt)
		argPos++
	}
	if req.AnnouncementAt != nil {
		updates = append(updates, fmt.Sprintf("announcement_at = $%d", argPos))
		args = append(args, *req.AnnouncementAt)
		argPos++
	}
	if req.FinishedAt != nil {
		updates = append(updates, fmt.Sprintf("finished_at = $%d", argPos))
		args = append(args, *req.FinishedAt)
		argPos++
	}

	if len(updates) == 0 {
		return r.GetElectionByID(ctx, id)
	}

	updates = append(updates, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf(`
UPDATE elections
SET %s
WHERE id = $%d
RETURNING
    id, year, name, code, status,
    registration_start_at, registration_end_at,
    verification_start_at, verification_end_at,
    campaign_start_at, campaign_end_at,
    quiet_start_at, quiet_end_at,
    voting_start_at, voting_end_at,
    recap_start_at, recap_end_at,
    announcement_at, finished_at,
    online_enabled, tps_enabled,
    created_at, updated_at
`, strings.Join(updates, ", "), argPos)

	args = append(args, id)

	var dto AdminElectionDTO
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&dto.ID,
		&dto.Year,
		&dto.Name,
		&dto.Slug,
		&dto.Status,
		&dto.RegistrationStartAt,
		&dto.RegistrationEndAt,
		&dto.VerificationStartAt,
		&dto.VerificationEndAt,
		&dto.CampaignStartAt,
		&dto.CampaignEndAt,
		&dto.QuietStartAt,
		&dto.QuietEndAt,
		&dto.VotingStartAt,
		&dto.VotingEndAt,
		&dto.RecapStartAt,
		&dto.RecapEndAt,
		&dto.AnnouncementAt,
		&dto.FinishedAt,
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &dto, nil
}

func (r *PgAdminRepository) SetVotingStatus(
	ctx context.Context,
	id int64,
	status ElectionStatus,
	votingStartAt, votingEndAt *time.Time,
) (*AdminElectionDTO, error) {
	const q = `
UPDATE elections
SET
    status = $2,
    voting_start_at = COALESCE($3, voting_start_at),
    voting_end_at   = COALESCE($4, voting_end_at),
    updated_at      = NOW()
WHERE id = $1
RETURNING
    id, year, name, code, status,
    registration_start_at, registration_end_at,
    verification_start_at, verification_end_at,
    campaign_start_at, campaign_end_at,
    quiet_start_at, quiet_end_at,
    voting_start_at, voting_end_at,
    recap_start_at, recap_end_at,
    announcement_at, finished_at,
    online_enabled, tps_enabled,
    created_at, updated_at
`
	var dto AdminElectionDTO
	err := r.db.QueryRow(ctx, q,
		id,
		status,
		votingStartAt,
		votingEndAt,
	).Scan(
		&dto.ID,
		&dto.Year,
		&dto.Name,
		&dto.Slug,
		&dto.Status,
		&dto.RegistrationStartAt,
		&dto.RegistrationEndAt,
		&dto.VerificationStartAt,
		&dto.VerificationEndAt,
		&dto.CampaignStartAt,
		&dto.CampaignEndAt,
		&dto.QuietStartAt,
		&dto.QuietEndAt,
		&dto.VotingStartAt,
		&dto.VotingEndAt,
		&dto.RecapStartAt,
		&dto.RecapEndAt,
		&dto.AnnouncementAt,
		&dto.FinishedAt,
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}
	return &dto, nil
}
