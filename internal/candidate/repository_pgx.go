package candidate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgCandidateRepository implements CandidateRepository using pgxpool
type PgCandidateRepository struct {
	db *pgxpool.Pool
}

// NewPgCandidateRepository creates a new PostgreSQL candidate repository
func NewPgCandidateRepository(db *pgxpool.Pool) *PgCandidateRepository {
	return &PgCandidateRepository{db: db}
}

const qListCandidatesBase = `
SELECT
id,
election_id,
number,
name,
photo_url,
photo_media_id,
short_bio,
long_bio,
tagline,
faculty_name,
study_program_name,
cohort_year,
vision,
missions,
main_programs,
media,
social_links,
status,
created_at,
updated_at
FROM candidates
WHERE election_id = $1
`

const qCountCandidatesBase = `
SELECT COUNT(*) FROM candidates WHERE election_id = $1
`

// ListByElection returns candidates for an election with filters and pagination
func (r *PgCandidateRepository) ListByElection(
	ctx context.Context,
	electionID int64,
	filter Filter,
) ([]Candidate, int64, error) {
	args := []any{electionID}
	where := ""

	// status filter
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}

	// simple search by name/tagline
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		where += fmt.Sprintf(" AND (name ILIKE $%d OR tagline ILIKE $%d)", len(args), len(args))
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// total count
	countSQL := qCountCandidatesBase + where
	var total int64
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// list query
	listSQL := qListCandidatesBase + where + `
ORDER BY number ASC
LIMIT $` + fmt.Sprint(len(args)+1) + `
OFFSET $` + fmt.Sprint(len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var candidates []Candidate
	for rows.Next() {
		c, err := scanCandidate(rows)
		if err != nil {
			return nil, 0, err
		}
		candidates = append(candidates, c)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return candidates, total, nil
}

const qGetCandidateByID = `
SELECT
id,
election_id,
number,
name,
photo_url,
photo_media_id,
short_bio,
long_bio,
tagline,
faculty_name,
study_program_name,
cohort_year,
vision,
missions,
main_programs,
media,
social_links,
status,
created_at,
updated_at
FROM candidates
WHERE election_id = $1 AND id = $2
`

const qGetCandidateByCandidateID = `
SELECT
id,
election_id,
number,
name,
photo_url,
photo_media_id,
short_bio,
long_bio,
tagline,
faculty_name,
study_program_name,
cohort_year,
vision,
missions,
main_programs,
media,
social_links,
status,
created_at,
updated_at
FROM candidates
WHERE id = $1
`

// GetByID returns a single candidate by election and candidate ID
func (r *PgCandidateRepository) GetByID(
	ctx context.Context,
	electionID, candidateID int64,
) (*Candidate, error) {
	row := r.db.QueryRow(ctx, qGetCandidateByID, electionID, candidateID)

	c, err := scanCandidateRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *PgCandidateRepository) GetByCandidateID(
	ctx context.Context,
	candidateID int64,
) (*Candidate, error) {
	row := r.db.QueryRow(ctx, qGetCandidateByCandidateID, candidateID)
	c, err := scanCandidateRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}
	return &c, nil
}

// scanCandidate scans a candidate from pgx.Rows
func scanCandidate(rows pgx.Rows) (Candidate, error) {
	var c Candidate
	var missionsRaw, mainProgramsRaw, mediaRaw, socialLinksRaw any

	if err := rows.Scan(
		&c.ID,
		&c.ElectionID,
		&c.Number,
		&c.Name,
		&c.PhotoURL,
		&c.PhotoMediaID,
		&c.ShortBio,
		&c.LongBio,
		&c.Tagline,
		&c.FacultyName,
		&c.StudyProgramName,
		&c.CohortYear,
		&c.Vision,
		&missionsRaw,
		&mainProgramsRaw,
		&mediaRaw,
		&socialLinksRaw,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return Candidate{}, err
	}

	if err := scanJSON(missionsRaw, &c.Missions); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(mainProgramsRaw, &c.MainPrograms); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(mediaRaw, &c.Media); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(socialLinksRaw, &c.SocialLinks); err != nil {
		return Candidate{}, err
	}

	return c, nil
}

// scanCandidateRow scans a candidate from pgx.Row
func scanCandidateRow(row pgx.Row) (Candidate, error) {
	var c Candidate
	var missionsRaw, mainProgramsRaw, mediaRaw, socialLinksRaw any

	err := row.Scan(
		&c.ID,
		&c.ElectionID,
		&c.Number,
		&c.Name,
		&c.PhotoURL,
		&c.PhotoMediaID,
		&c.ShortBio,
		&c.LongBio,
		&c.Tagline,
		&c.FacultyName,
		&c.StudyProgramName,
		&c.CohortYear,
		&c.Vision,
		&missionsRaw,
		&mainProgramsRaw,
		&mediaRaw,
		&socialLinksRaw,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return Candidate{}, err
	}

	if err := scanJSON(missionsRaw, &c.Missions); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(mainProgramsRaw, &c.MainPrograms); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(mediaRaw, &c.Media); err != nil {
		return Candidate{}, err
	}
	if err := scanJSON(socialLinksRaw, &c.SocialLinks); err != nil {
		return Candidate{}, err
	}

	return c, nil
}

// scanJSON scans JSONB data into a Go type
func scanJSON[T any](src any, dest *T) error {
	if src == nil {
		return nil
	}

	var b []byte

	switch v := src.(type) {
	case []byte:
		if len(v) == 0 {
			return nil
		}
		b = v
	case string:
		if v == "" {
			return nil
		}
		b = []byte(v)
	default:
		var err error
		b, err = json.Marshal(v)
		if err != nil {
			logJSONError(fmt.Errorf("marshal fallback failed for type %T: %w", src, err))
			return err
		}
	}

	err := json.Unmarshal(b, dest)
	if err != nil {
		logJSONError(fmt.Errorf("scanJSON unmarshal error: %w, data: %s", err, string(b)))
		return err
	}
	return nil
}

func logJSONError(err error) {
	if f, ferr := os.OpenFile("/tmp/scanJSON_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); ferr == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("%v\n", err))
	}
}

const qCreateCandidate = `
INSERT INTO candidates (
election_id, number, name, photo_url, photo_media_id, short_bio, long_bio, tagline,
faculty_name, study_program_name, cohort_year, vision, missions,
main_programs, media, social_links, status, created_at, updated_at, updated_by_admin_id
) VALUES (
$1, $2, $3, $4, NULL, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW(), NULL
)
RETURNING id, election_id, number, name, photo_url, photo_media_id, short_bio, long_bio, tagline,
faculty_name, study_program_name, cohort_year, vision, missions, main_programs,
media, social_links, status, created_at, updated_at
`

// Create creates a new candidate
func (r *PgCandidateRepository) Create(ctx context.Context, candidate *Candidate) (*Candidate, error) {
	missionsJSON, _ := json.Marshal(candidate.Missions)
	mainProgramsJSON, _ := json.Marshal(candidate.MainPrograms)
	mediaJSON, _ := json.Marshal(candidate.Media)
	socialLinksJSON, _ := json.Marshal(candidate.SocialLinks)

	row := r.db.QueryRow(ctx, qCreateCandidate,
		candidate.ElectionID,
		candidate.Number,
		candidate.Name,
		candidate.PhotoURL,
		candidate.ShortBio,
		candidate.LongBio,
		candidate.Tagline,
		candidate.FacultyName,
		candidate.StudyProgramName,
		candidate.CohortYear,
		candidate.Vision,
		missionsJSON,
		mainProgramsJSON,
		mediaJSON,
		socialLinksJSON,
		candidate.Status,
	)

	c, err := scanCandidateRow(row)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

const qUpdateCandidate = `
UPDATE candidates SET
number = $3,
name = $4,
photo_url = $5,
short_bio = $6,
long_bio = $7,
tagline = $8,
faculty_name = $9,
study_program_name = $10,
cohort_year = $11,
vision = $12,
missions = $13,
main_programs = $14,
media = $15,
social_links = $16,
status = $17,
updated_at = NOW()
WHERE election_id = $1 AND id = $2
RETURNING id, election_id, number, name, photo_url, photo_media_id, short_bio, long_bio, tagline,
faculty_name, study_program_name, cohort_year, vision, missions, main_programs,
media, social_links, status, created_at, updated_at
`

// Update updates an existing candidate
func (r *PgCandidateRepository) Update(ctx context.Context, electionID, candidateID int64, candidate *Candidate) (*Candidate, error) {
	missionsJSON, _ := json.Marshal(candidate.Missions)
	mainProgramsJSON, _ := json.Marshal(candidate.MainPrograms)
	mediaJSON, _ := json.Marshal(candidate.Media)
	socialLinksJSON, _ := json.Marshal(candidate.SocialLinks)

	row := r.db.QueryRow(ctx, qUpdateCandidate,
		electionID,
		candidateID,
		candidate.Number,
		candidate.Name,
		candidate.PhotoURL,
		candidate.ShortBio,
		candidate.LongBio,
		candidate.Tagline,
		candidate.FacultyName,
		candidate.StudyProgramName,
		candidate.CohortYear,
		candidate.Vision,
		missionsJSON,
		mainProgramsJSON,
		mediaJSON,
		socialLinksJSON,
		candidate.Status,
	)

	c, err := scanCandidateRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}

	return &c, nil
}

const qDeleteCandidate = `
DELETE FROM candidates WHERE election_id = $1 AND id = $2
`

// Delete deletes a candidate
func (r *PgCandidateRepository) Delete(ctx context.Context, electionID, candidateID int64) error {
	result, err := r.db.Exec(ctx, qDeleteCandidate, electionID, candidateID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCandidateNotFound
	}

	return nil
}

const qUpdateStatus = `
UPDATE candidates SET status = $3, updated_at = NOW()
WHERE election_id = $1 AND id = $2
`

// UpdateStatus updates candidate status
func (r *PgCandidateRepository) UpdateStatus(ctx context.Context, electionID, candidateID int64, status CandidateStatus) error {
	result, err := r.db.Exec(ctx, qUpdateStatus, electionID, candidateID, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCandidateNotFound
	}

	return nil
}

const qCheckNumberExists = `
SELECT EXISTS(
SELECT 1 FROM candidates
WHERE election_id = $1 AND number = $2 AND ($3::bigint IS NULL OR id != $3)
)
`

// CheckNumberExists checks if candidate number is already taken in an election
func (r *PgCandidateRepository) CheckNumberExists(ctx context.Context, electionID int64, number int, excludeCandidateID *int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, qCheckNumberExists, electionID, number, excludeCandidateID).Scan(&exists)
	return exists, err
}

func scanCandidateMedia(row pgx.Row) (*CandidateMedia, error) {
	var media CandidateMedia
	err := row.Scan(
		&media.ID,
		&media.CandidateID,
		&media.Slot,
		&media.FileName,
		&media.ContentType,
		&media.SizeBytes,
		&media.Data,
		&media.CreatedAt,
		&media.CreatedByID,
	)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// SaveProfileMedia stores/replace profile media and updates photo_media_id
func (r *PgCandidateRepository) SaveProfileMedia(
	ctx context.Context,
	candidateID int64,
	media CandidateMediaCreate,
) (*CandidateMedia, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var existingProfileID *string
	err = tx.QueryRow(ctx, `
SELECT photo_media_id FROM candidates WHERE id = $1 FOR UPDATE
`, candidateID).Scan(&existingProfileID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}

	if existingProfileID != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM candidate_media WHERE id = $1`, *existingProfileID); err != nil {
			return nil, err
		}
	}

	var createdAt time.Time
	err = tx.QueryRow(ctx, `
INSERT INTO candidate_media (id, candidate_id, slot, file_name, content_type, size_bytes, data, created_by_admin_id)
VALUES ($1, $2, 'profile', $3, $4, $5, $6, $7)
RETURNING created_at
`, media.ID, candidateID, media.FileName, media.ContentType, media.SizeBytes, media.Data, media.CreatedByID).Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
UPDATE candidates
SET photo_media_id = $1,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE id = $3
`, media.ID, media.CreatedByID, candidateID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CandidateMedia{
		ID:          media.ID,
		CandidateID: candidateID,
		Slot:        CandidateMediaSlotProfile,
		FileName:    media.FileName,
		ContentType: media.ContentType,
		SizeBytes:   media.SizeBytes,
		Data:        media.Data,
		CreatedAt:   createdAt,
		CreatedByID: &media.CreatedByID,
	}, nil
}

// GetProfileMedia retrieves profile media; 404 if missing
func (r *PgCandidateRepository) GetProfileMedia(ctx context.Context, candidateID int64) (*CandidateMedia, error) {
	var photoURL *string
	err := r.db.QueryRow(ctx, `
SELECT photo_url FROM candidates WHERE id = $1
`, candidateID).Scan(&photoURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateNotFound
		}
		return nil, err
	}
	if photoURL == nil || *photoURL == "" {
		return nil, ErrCandidateMediaNotFound
	}

	// Return media with URL from Supabase (photo_url field)
	return &CandidateMedia{
		CandidateID: candidateID,
		Slot:        CandidateMediaSlotProfile,
		URL:         *photoURL,
		ContentType: "image/jpeg",
	}, nil
}

// DeleteProfileMedia removes profile media and clears reference
func (r *PgCandidateRepository) DeleteProfileMedia(ctx context.Context, candidateID int64, adminID int64) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var mediaID *string
	err = tx.QueryRow(ctx, `
SELECT photo_media_id FROM candidates WHERE id = $1 FOR UPDATE
`, candidateID).Scan(&mediaID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrCandidateNotFound
		}
		return err
	}
	if mediaID == nil {
		return ErrCandidateMediaNotFound
	}

	if _, err := tx.Exec(ctx, `
UPDATE candidates
SET photo_media_id = NULL,
    updated_by_admin_id = $2,
    updated_at = NOW()
WHERE id = $1
`, candidateID, adminID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM candidate_media WHERE id = $1`, *mediaID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// AddMedia stores non-profile media
func (r *PgCandidateRepository) AddMedia(ctx context.Context, candidateID int64, media CandidateMediaCreate) (*CandidateMedia, error) {
	if media.Slot == CandidateMediaSlotProfile {
		return nil, ErrInvalidCandidateMediaSlot
	}

	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM candidates WHERE id = $1)`, candidateID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCandidateNotFound
	}

	var createdAt time.Time
	err := r.db.QueryRow(ctx, `
INSERT INTO candidate_media (id, candidate_id, slot, file_name, content_type, size_bytes, data, created_by_admin_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING created_at
`, media.ID, candidateID, media.Slot, media.FileName, media.ContentType, media.SizeBytes, media.Data, media.CreatedByID).Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	return &CandidateMedia{
		ID:          media.ID,
		CandidateID: candidateID,
		Slot:        media.Slot,
		FileName:    media.FileName,
		ContentType: media.ContentType,
		SizeBytes:   media.SizeBytes,
		Data:        media.Data,
		CreatedAt:   createdAt,
		CreatedByID: &media.CreatedByID,
	}, nil
}

// GetMedia fetches any media by id scoped to candidate
func (r *PgCandidateRepository) GetMedia(ctx context.Context, candidateID int64, mediaID string) (*CandidateMedia, error) {
	row := r.db.QueryRow(ctx, `
SELECT id, candidate_id, slot, file_name, content_type, size_bytes, data, created_at, created_by_admin_id
FROM candidate_media
WHERE candidate_id = $1 AND id = $2
`, candidateID, mediaID)
	media, err := scanCandidateMedia(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCandidateMediaNotFound
		}
		return nil, err
	}
	return media, nil
}

// DeleteMedia deletes a media asset and clears profile reference if needed
func (r *PgCandidateRepository) DeleteMedia(ctx context.Context, candidateID int64, mediaID string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var slot CandidateMediaSlot
	err = tx.QueryRow(ctx, `
SELECT slot FROM candidate_media WHERE candidate_id = $1 AND id = $2 FOR UPDATE
`, candidateID, mediaID).Scan(&slot)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrCandidateMediaNotFound
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM candidate_media WHERE id = $1`, mediaID); err != nil {
		return err
	}

	// Clear photo reference if the deleted media was profile
	if slot == CandidateMediaSlotProfile {
		if _, err := tx.Exec(ctx, `
UPDATE candidates SET photo_media_id = NULL, updated_at = NOW()
WHERE id = $1 AND photo_media_id = $2
`, candidateID, mediaID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ListMediaMeta returns lightweight metadata for a candidate
func (r *PgCandidateRepository) ListMediaMeta(ctx context.Context, candidateID int64) ([]CandidateMediaMeta, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, slot, file_name, content_type, size_bytes, created_at
FROM candidate_media
WHERE candidate_id = $1
ORDER BY created_at DESC
`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CandidateMediaMeta
	for rows.Next() {
		var meta CandidateMediaMeta
		if err := rows.Scan(&meta.ID, &meta.Slot, &meta.Label, &meta.ContentType, &meta.SizeBytes, &meta.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, meta)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}
