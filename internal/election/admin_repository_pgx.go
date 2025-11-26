package election

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgAdminRepository struct {
	db *pgxpool.Pool
}

func NewPgAdminRepository(db *pgxpool.Pool) *PgAdminRepository {
	return &PgAdminRepository{db: db}
}

const adminElectionColumns = `
    id,
    year,
    name,
    code,
    COALESCE(slug, '') as slug,
    description,
    academic_year,
    status,
    current_phase,
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
    online_login_url,
    online_max_sessions_per_voter,
    tps_require_checkin,
    tps_require_ballot_qr,
    tps_max,
    created_at,
    updated_at
`

var phaseColumnMap = map[ElectionPhaseKey]struct {
	startCol string
	endCol   string
}{
	PhaseKeyRegistration: {"registration_start_at", "registration_end_at"},
	PhaseKeyVerification: {"verification_start_at", "verification_end_at"},
	PhaseKeyCampaign:     {"campaign_start_at", "campaign_end_at"},
	PhaseKeyQuietPeriod:  {"quiet_start_at", "quiet_end_at"},
	PhaseKeyVoting:       {"voting_start_at", "voting_end_at"},
	PhaseKeyRecap:        {"recap_start_at", "recap_end_at"},
}

func nullableString[T any](body *T, getter func(*T) *string) *string {
	if body == nil {
		return nil
	}
	return getter(body)
}

func nullableInt[T any](body *T, getter func(*T) *int) *int {
	if body == nil {
		return nil
	}
	return getter(body)
}

func nullableBool[T any](body *T, getter func(*T) *bool) *bool {
	if body == nil {
		return nil
	}
	return getter(body)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAdminElection(row rowScanner) (*AdminElectionDTO, error) {
	var dto AdminElectionDTO
	err := row.Scan(
		&dto.ID,
		&dto.Year,
		&dto.Name,
		&dto.Code,
		&dto.Slug,
		&dto.Description,
		&dto.AcademicYear,
		&dto.Status,
		&dto.CurrentPhase,
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
		&dto.OnlineLoginURL,
		&dto.OnlineMaxSessions,
		&dto.TPSRequireCheckin,
		&dto.TPSRequireBallotQR,
		&dto.TPSMax,
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
SELECT %s
FROM elections
%s
ORDER BY year DESC, id DESC
LIMIT $%d OFFSET $%d
`, adminElectionColumns, whereClause, argPos, argPos+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []AdminElectionDTO{}
	for rows.Next() {
		dto, err := scanAdminElection(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *dto)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *PgAdminRepository) GetElectionByID(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	const base = `
SELECT %s
FROM elections
WHERE id = $1
`
	q := fmt.Sprintf(base, adminElectionColumns)
	return scanAdminElection(r.db.QueryRow(ctx, q, id))
}

func (r *PgAdminRepository) CreateElection(ctx context.Context, req AdminElectionCreateRequest) (*AdminElectionDTO, error) {
	q := fmt.Sprintf(`
INSERT INTO elections (
    year,
    name,
    code,
    status,
    online_enabled,
    tps_enabled
) VALUES ($1, $2, $3, 'DRAFT', $4, $5)
RETURNING %s
`, adminElectionColumns)

	return scanAdminElection(r.db.QueryRow(ctx, q,
		req.Year,
		req.Name,
		req.Slug,
		req.OnlineEnabled,
		req.TPSEnabled,
	))
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
RETURNING %s
`, strings.Join(updates, ", "), argPos, adminElectionColumns)

	args = append(args, id)

	return scanAdminElection(r.db.QueryRow(ctx, query, args...))
}

func (r *PgAdminRepository) SetVotingStatus(
	ctx context.Context,
	id int64,
	status ElectionStatus,
	currentPhase *string,
	votingStartAt, votingEndAt *time.Time,
) (*AdminElectionDTO, error) {
	const q = `
UPDATE elections
SET
    status = $2,
    current_phase = COALESCE($3, current_phase),
    voting_start_at = COALESCE($4, voting_start_at),
    voting_end_at   = COALESCE($5, voting_end_at),
    updated_at      = NOW()
WHERE id = $1
RETURNING %s
`
	return scanAdminElection(r.db.QueryRow(ctx, fmt.Sprintf(q, adminElectionColumns),
		id,
		status,
		currentPhase,
		votingStartAt,
		votingEndAt,
	))
}

func (r *PgAdminRepository) UpdateGeneralInfo(ctx context.Context, id int64, req AdminElectionGeneralUpdateRequest) (*AdminElectionDTO, error) {
	updates := []string{}
	args := []any{}
	pos := 1

	if req.Year != nil {
		updates = append(updates, fmt.Sprintf("year = $%d", pos))
		args = append(args, *req.Year)
		pos++
	}
	if req.Code != nil {
		updates = append(updates, fmt.Sprintf("code = $%d", pos))
		args = append(args, *req.Code)
		pos++
	}
	if req.Slug != nil {
		updates = append(updates, fmt.Sprintf("slug = $%d", pos))
		args = append(args, *req.Slug)
		pos++
	}
	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", pos))
		args = append(args, *req.Name)
		pos++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", pos))
		args = append(args, *req.Description)
		pos++
	}
	if req.AcademicYear != nil {
		updates = append(updates, fmt.Sprintf("academic_year = $%d", pos))
		args = append(args, *req.AcademicYear)
		pos++
	}

	if len(updates) == 0 {
		return r.GetElectionByID(ctx, id)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf(`
UPDATE elections
SET %s
WHERE id = $%d
RETURNING %s
`, strings.Join(updates, ", "), pos, adminElectionColumns)

	args = append(args, id)
	return scanAdminElection(r.db.QueryRow(ctx, query, args...))
}

func (r *PgAdminRepository) GetPhases(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	return r.GetElectionByID(ctx, id)
}

func (r *PgAdminRepository) UpdatePhases(ctx context.Context, id int64, phases []ElectionPhaseInput) (*AdminElectionDTO, error) {
	updates := []string{}
	args := []any{}
	pos := 1

	for _, ph := range phases {
		cols, ok := phaseColumnMap[ph.Key]
		if !ok {
			continue
		}
		updates = append(updates, fmt.Sprintf("%s = $%d", cols.startCol, pos))
		args = append(args, ph.StartAt)
		pos++

		updates = append(updates, fmt.Sprintf("%s = $%d", cols.endCol, pos))
		args = append(args, ph.EndAt)
		pos++
	}

	if len(updates) == 0 {
		return r.GetElectionByID(ctx, id)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf(`
UPDATE elections
SET %s
WHERE id = $%d
RETURNING %s
`, strings.Join(updates, ", "), pos, adminElectionColumns)

	args = append(args, id)
	return scanAdminElection(r.db.QueryRow(ctx, query, args...))
}

func (r *PgAdminRepository) GetModeSettings(ctx context.Context, id int64) (*ModeSettingsDTO, error) {
	const q = `
SELECT
    online_enabled,
    tps_enabled,
    online_login_url,
    online_max_sessions_per_voter,
    tps_require_checkin,
    tps_require_ballot_qr,
    tps_max,
    updated_at
FROM elections
WHERE id = $1
`

	var dto ModeSettingsDTO
	dto.ElectionID = id
	err := r.db.QueryRow(ctx, q, id).Scan(
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.OnlineSettings.LoginURL,
		&dto.OnlineSettings.MaxSessionsPerVoter,
		&dto.TPSSettings.RequireCheckin,
		&dto.TPSSettings.RequireBallotQR,
		&dto.TPSSettings.MaxTPS,
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

func (r *PgAdminRepository) UpdateModeSettings(ctx context.Context, id int64, req ModeSettingsRequest) (*ModeSettingsDTO, error) {
	const q = `
UPDATE elections
SET
    online_enabled = COALESCE($2, online_enabled),
    tps_enabled = COALESCE($3, tps_enabled),
    online_login_url = COALESCE($4, online_login_url),
    online_max_sessions_per_voter = COALESCE($5, online_max_sessions_per_voter),
    tps_require_checkin = COALESCE($6, tps_require_checkin),
    tps_require_ballot_qr = COALESCE($7, tps_require_ballot_qr),
    tps_max = COALESCE($8, tps_max),
    updated_at = NOW()
WHERE id = $1
RETURNING
    online_enabled,
    tps_enabled,
    online_login_url,
    online_max_sessions_per_voter,
    tps_require_checkin,
    tps_require_ballot_qr,
    tps_max,
    updated_at
`

	var dto ModeSettingsDTO
	dto.ElectionID = id

	var onlineSettings *OnlineSettingsBody
	if req.OnlineSettings != nil {
		onlineSettings = req.OnlineSettings
	}
	var tpsSettings *TPSSettingsBody
	if req.TPSSettings != nil {
		tpsSettings = req.TPSSettings
	}

	err := r.db.QueryRow(ctx, q,
		id,
		req.OnlineEnabled,
		req.TPSEnabled,
		nullableString(onlineSettings, func(s *OnlineSettingsBody) *string { return s.LoginURL }),
		nullableInt(onlineSettings, func(s *OnlineSettingsBody) *int { return s.MaxSessionsPerVoter }),
		nullableBool(tpsSettings, func(s *TPSSettingsBody) *bool { return s.RequireCheckin }),
		nullableBool(tpsSettings, func(s *TPSSettingsBody) *bool { return s.RequireBallotQR }),
		nullableInt(tpsSettings, func(s *TPSSettingsBody) *int { return s.MaxTPS }),
	).Scan(
		&dto.OnlineEnabled,
		&dto.TPSEnabled,
		&dto.OnlineSettings.LoginURL,
		&dto.OnlineSettings.MaxSessionsPerVoter,
		&dto.TPSSettings.RequireCheckin,
		&dto.TPSSettings.RequireBallotQR,
		&dto.TPSSettings.MaxTPS,
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

func (r *PgAdminRepository) GetSummary(ctx context.Context, id int64) (*ElectionSummaryDTO, error) {
	const q = `
SELECT
    e.status,
    e.current_phase,
    (SELECT COUNT(*) FROM candidates c WHERE c.election_id = $1) AS total_candidates,
    (SELECT COUNT(*) FROM candidates c WHERE c.election_id = $1 AND c.status = 'APPROVED') AS published_candidates,
    (SELECT COUNT(*) FROM voter_status vs WHERE vs.election_id = $1) AS total_voters,
    (SELECT COUNT(*) FROM voter_status vs WHERE vs.election_id = $1 AND vs.preferred_method = 'ONLINE') AS online_voters,
    (SELECT COUNT(*) FROM voter_status vs WHERE vs.election_id = $1 AND vs.preferred_method = 'TPS') AS tps_voters,
    (SELECT COUNT(*) FROM tps t WHERE t.election_id = $1) AS total_tps,
    (SELECT COUNT(*) FROM tps t WHERE t.election_id = $1 AND t.status = 'ACTIVE') AS active_tps,
    (SELECT COUNT(*) FROM votes v WHERE v.election_id = $1) AS total_votes,
    (SELECT COUNT(*) FROM votes v WHERE v.election_id = $1 AND v.channel = 'ONLINE') AS online_votes,
    (SELECT COUNT(*) FROM votes v WHERE v.election_id = $1 AND v.channel = 'TPS') AS tps_votes
FROM elections e
WHERE e.id = $1
`

	var dto ElectionSummaryDTO
	dto.ElectionID = id
	var currentPhase *string

	err := r.db.QueryRow(ctx, q, id).Scan(
		&dto.Status,
		&currentPhase,
		&dto.Candidates.Total,
		&dto.Candidates.Published,
		&dto.DPT.TotalVoters,
		&dto.DPT.OnlineVoters,
		&dto.DPT.TPSVoters,
		&dto.TPS.TotalTPS,
		&dto.TPS.ActiveTPS,
		&dto.Votes.TotalCast,
		&dto.Votes.OnlineCast,
		&dto.Votes.TPSCast,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}

	if currentPhase != nil {
		dto.CurrentPhase = *currentPhase
	}

	return &dto, nil
}

func brandingColumn(slot BrandingSlot) (string, error) {
	switch slot {
	case BrandingSlotPrimary:
		return "primary_logo_id", nil
	case BrandingSlotSecondary:
		return "secondary_logo_id", nil
	default:
		return "", ErrInvalidBrandingSlot
	}
}

func (r *PgAdminRepository) ensureBrandingSettings(ctx context.Context, electionID int64) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO branding_settings (election_id)
VALUES ($1)
ON CONFLICT (election_id) DO NOTHING
`, electionID)
	return err
}

func ensureBrandingSettingsTx(ctx context.Context, tx pgx.Tx, electionID int64) error {
	_, err := tx.Exec(ctx, `
INSERT INTO branding_settings (election_id)
VALUES ($1)
ON CONFLICT (election_id) DO NOTHING
`, electionID)
	return err
}

// Supabase helpers for branding storage
type supabaseStorage struct {
	client *storage_go.Client
	url    string
	bucket string
}

func newBrandingStorage() (*supabaseStorage, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SECRET_KEY")
	bucket := os.Getenv("SUPABASE_BRANDING_BUCKET")
	if bucket == "" {
		bucket = "branding"
	}
	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SECRET_KEY required")
	}
	headers := map[string]string{"apikey": key}
	client := storage_go.NewClient(url+"/storage/v1", key, headers)
	return &supabaseStorage{client: client, url: url, bucket: bucket}, nil
}

func (s *supabaseStorage) Upload(ctx context.Context, path string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	if _, err := s.client.UploadFile(s.bucket, path, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucket, path), nil
}

func getBrandingExtension(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ".bin"
	}
}

func (r *PgAdminRepository) GetBranding(ctx context.Context, electionID int64) (*BrandingSettings, error) {
	if err := r.ensureBrandingSettings(ctx, electionID); err != nil {
		return nil, err
	}

	const q = `
SELECT
    bs.primary_logo_id,
    bs.secondary_logo_id,
    bs.updated_at,
    ua.id,
    ua.username
FROM branding_settings bs
LEFT JOIN user_accounts ua ON ua.id = bs.updated_by_admin_id
WHERE bs.election_id = $1
`

	var primaryID, secondaryID *string
	var updatedAt time.Time
	var updatedByID *int64
	var updatedByUsername *string

	err := r.db.QueryRow(ctx, q, electionID).Scan(
		&primaryID,
		&secondaryID,
		&updatedAt,
		&updatedByID,
		&updatedByUsername,
	)
	if err != nil {
		return nil, err
	}

	branding := &BrandingSettings{
		PrimaryLogoID:   primaryID,
		SecondaryLogoID: secondaryID,
		UpdatedAt:       updatedAt,
	}
	if updatedByID != nil && updatedByUsername != nil {
		branding.UpdatedBy = &BrandingUser{
			ID:       *updatedByID,
			Username: *updatedByUsername,
		}
	}

	return branding, nil
}

func (r *PgAdminRepository) GetBrandingFile(ctx context.Context, electionID int64, slot BrandingSlot) (*BrandingFile, error) {
	column, err := brandingColumn(slot)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
SELECT
    bf.id,
    bf.election_id,
    bf.slot,
    bf.content_type,
    bf.size_bytes,
    bf.storage_path,
    bf.created_at,
    bf.created_by_admin_id
FROM branding_settings bs
JOIN branding_files bf ON bf.id = bs.%s
WHERE bs.election_id = $1
`, column)

	var file BrandingFile
	var storagePath string
	err = r.db.QueryRow(ctx, query, electionID).Scan(
		&file.ID,
		&file.ElectionID,
		&file.Slot,
		&file.ContentType,
		&file.SizeBytes,
		&storagePath,
		&file.CreatedAt,
		&file.CreatedByID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBrandingFileNotFound
		}
		return nil, err
	}

	if len(storagePath) > 0 && strings.HasPrefix(storagePath, "http") {
		file.URL = &storagePath
	} else {
		file.Data = []byte(storagePath)
	}

	return &file, nil
}

func (r *PgAdminRepository) SaveBrandingFile(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	file BrandingFileCreate,
) (*BrandingFile, error) {
	// Try Supabase upload first
	dataToStore := file.Data
	var uploadedURL *string
	if storage, _ := newBrandingStorage(); storage != nil {
		path := fmt.Sprintf("branding/%d/%s/%s%s", electionID, slot, file.ID, getBrandingExtension(file.ContentType))
		if url, err := storage.Upload(ctx, path, file.Data, file.ContentType); err == nil {
			dataToStore = []byte(url)
			uploadedURL = &url
		}
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := ensureBrandingSettingsTx(ctx, tx, electionID); err != nil {
		return nil, err
	}

	var currentPrimaryID, currentSecondaryID *string
	err = tx.QueryRow(ctx, `
SELECT primary_logo_id, secondary_logo_id
FROM branding_settings
WHERE election_id = $1
FOR UPDATE
`, electionID).Scan(&currentPrimaryID, &currentSecondaryID)
	if err != nil {
		return nil, err
	}

	var createdAt time.Time
	err = tx.QueryRow(ctx, `
INSERT INTO branding_files (id, election_id, slot, content_type, size_bytes, storage_path, created_by_admin_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING created_at
`, file.ID, electionID, slot, file.ContentType, file.SizeBytes, string(dataToStore), file.CreatedByID).Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	var oldID *string
	switch slot {
	case BrandingSlotPrimary:
		oldID = currentPrimaryID
		_, err = tx.Exec(ctx, `
UPDATE branding_settings
SET primary_logo_id = $1,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE election_id = $3
`, file.ID, file.CreatedByID, electionID)
	case BrandingSlotSecondary:
		oldID = currentSecondaryID
		_, err = tx.Exec(ctx, `
UPDATE branding_settings
SET secondary_logo_id = $1,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE election_id = $3
`, file.ID, file.CreatedByID, electionID)
	default:
		err = ErrInvalidBrandingSlot
	}
	if err != nil {
		return nil, err
	}

	if oldID != nil && *oldID != file.ID {
		if _, err := tx.Exec(ctx, `DELETE FROM branding_files WHERE id = $1`, *oldID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &BrandingFile{
		ID:          file.ID,
		ElectionID:  electionID,
		Slot:        slot,
		ContentType: file.ContentType,
		SizeBytes:   file.SizeBytes,
		Data:        file.Data,
		URL:         uploadedURL,
		CreatedAt:   createdAt,
		CreatedByID: &file.CreatedByID,
	}, nil
}

func (r *PgAdminRepository) DeleteBrandingFile(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	adminID int64,
) (*BrandingSettings, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := ensureBrandingSettingsTx(ctx, tx, electionID); err != nil {
		return nil, err
	}

	var primaryID, secondaryID *string
	err = tx.QueryRow(ctx, `
SELECT primary_logo_id, secondary_logo_id
FROM branding_settings
WHERE election_id = $1
FOR UPDATE
`, electionID).Scan(&primaryID, &secondaryID)
	if err != nil {
		return nil, err
	}

	var targetID *string
	switch slot {
	case BrandingSlotPrimary:
		targetID = primaryID
	case BrandingSlotSecondary:
		targetID = secondaryID
	default:
		return nil, ErrInvalidBrandingSlot
	}

	var updatedPrimaryID, updatedSecondaryID *string
	var updatedAt time.Time
	var updatedByID *int64
	var updatedByUsername *string

	switch slot {
	case BrandingSlotPrimary:
		err = tx.QueryRow(ctx, `
UPDATE branding_settings
SET primary_logo_id = NULL,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE election_id = $1
RETURNING primary_logo_id, secondary_logo_id, updated_at, updated_by_admin_id,
    (SELECT username FROM user_accounts WHERE id = updated_by_admin_id)
`, electionID, adminID).Scan(&updatedPrimaryID, &updatedSecondaryID, &updatedAt, &updatedByID, &updatedByUsername)
	case BrandingSlotSecondary:
		err = tx.QueryRow(ctx, `
UPDATE branding_settings
SET secondary_logo_id = NULL,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE election_id = $1
RETURNING primary_logo_id, secondary_logo_id, updated_at, updated_by_admin_id,
    (SELECT username FROM user_accounts WHERE id = updated_by_admin_id)
`, electionID, adminID).Scan(&updatedPrimaryID, &updatedSecondaryID, &updatedAt, &updatedByID, &updatedByUsername)
	}
	if err != nil {
		return nil, err
	}

	if targetID != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM branding_files WHERE id = $1`, *targetID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	branding := &BrandingSettings{
		PrimaryLogoID:   updatedPrimaryID,
		SecondaryLogoID: updatedSecondaryID,
		UpdatedAt:       updatedAt,
	}
	if updatedByID != nil && updatedByUsername != nil {
		branding.UpdatedBy = &BrandingUser{
			ID:       *updatedByID,
			Username: *updatedByUsername,
		}
	}

	return branding, nil
}
