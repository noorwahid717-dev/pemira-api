package candidate

import (
"context"
"encoding/json"
"fmt"

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
b, ok := src.([]byte)
if !ok {
return fmt.Errorf("invalid jsonb type %T", src)
}
if len(b) == 0 {
return nil
}
return json.Unmarshal(b, dest)
}
