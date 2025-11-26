package voter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pemira-api/internal/shared"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

func (r *PgRepository) GetByID(ctx context.Context, id int64) (*Voter, error) {
	query := `
		SELECT id, nim, name, email, phone, faculty_code, faculty_name, 
		       study_program_code, study_program_name, cohort_year, class_label,
		       photo_url, bio, voting_method_preference, academic_status, 
		       created_at, updated_at
		FROM voters
		WHERE id = $1
	`

	var v Voter
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.NIM, &v.Name, &v.Email, &v.Phone,
		&v.FacultyCode, &v.FacultyName, &v.StudyProgramCode, &v.StudyProgramName,
		&v.CohortYear, &v.ClassLabel, &v.PhotoURL, &v.Bio,
		&v.VotingMethodPreference, &v.AcademicStatus, &v.CreatedAt, &v.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVoterNotFound
		}
		return nil, err
	}

	return &v, nil
}

func (r *PgRepository) GetByNIM(ctx context.Context, nim string) (*Voter, error) {
	query := `
		SELECT id, nim, name, email, phone, faculty_code, faculty_name, 
		       study_program_code, study_program_name, cohort_year, class_label,
		       photo_url, bio, voting_method_preference, academic_status, 
		       created_at, updated_at
		FROM voters
		WHERE nim = $1
	`

	var v Voter
	err := r.db.QueryRow(ctx, query, nim).Scan(
		&v.ID, &v.NIM, &v.Name, &v.Email, &v.Phone,
		&v.FacultyCode, &v.FacultyName, &v.StudyProgramCode, &v.StudyProgramName,
		&v.CohortYear, &v.ClassLabel, &v.PhotoURL, &v.Bio,
		&v.VotingMethodPreference, &v.AcademicStatus, &v.CreatedAt, &v.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVoterNotFound
		}
		return nil, err
	}

	return &v, nil
}

func (r *PgRepository) List(ctx context.Context, params shared.PaginationParams) ([]*Voter, int64, error) {
	countQuery := `SELECT COUNT(*) FROM voters`
	var total int64
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, nim, name, email, phone, faculty_code, faculty_name, 
		       study_program_code, study_program_name, cohort_year, class_label,
		       photo_url, bio, voting_method_preference, academic_status, 
		       created_at, updated_at
		FROM voters
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var voters []*Voter
	for rows.Next() {
		var v Voter
		err := rows.Scan(
			&v.ID, &v.NIM, &v.Name, &v.Email, &v.Phone,
			&v.FacultyCode, &v.FacultyName, &v.StudyProgramCode, &v.StudyProgramName,
			&v.CohortYear, &v.ClassLabel, &v.PhotoURL, &v.Bio,
			&v.VotingMethodPreference, &v.AcademicStatus, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		voters = append(voters, &v)
	}

	return voters, total, nil
}

func (r *PgRepository) Create(ctx context.Context, voter *Voter) error {
	query := `
		INSERT INTO voters (nim, name, email, phone, faculty_code, faculty_name, 
		                    study_program_code, study_program_name, cohort_year, 
		                    class_label, academic_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		voter.NIM, voter.Name, voter.Email, voter.Phone,
		voter.FacultyCode, voter.FacultyName, voter.StudyProgramCode,
		voter.StudyProgramName, voter.CohortYear, voter.ClassLabel,
		voter.AcademicStatus,
	).Scan(&voter.ID, &voter.CreatedAt, &voter.UpdatedAt)
}

func (r *PgRepository) Update(ctx context.Context, voter *Voter) error {
	query := `
		UPDATE voters
		SET email = $2, phone = $3, photo_url = $4, bio = $5, 
		    voting_method_preference = $6, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		voter.ID, voter.Email, voter.Phone, voter.PhotoURL,
		voter.Bio, voter.VotingMethodPreference,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrVoterNotFound
	}

	return nil
}

func (r *PgRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM voters WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrVoterNotFound
	}

	return nil
}

func (r *PgRepository) GetElectionStatus(ctx context.Context, voterID, electionID int64) (*VoterElectionStatus, error) {
	query := `
		SELECT id, voter_id, election_id, status, has_voted, voted_at, voted_via, created_at, updated_at
		FROM voter_status
		WHERE voter_id = $1 AND election_id = $2
	`

	var status VoterElectionStatus
	err := r.db.QueryRow(ctx, query, voterID, electionID).Scan(
		&status.ID, &status.VoterID, &status.ElectionID,
		&status.Status, &status.HasVoted, &status.VotedAt,
		&status.VotedVia, &status.CreatedAt, &status.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}

	return &status, nil
}

func (r *PgRepository) CreateElectionStatus(ctx context.Context, status *VoterElectionStatus) error {
	query := `
		INSERT INTO voter_status (voter_id, election_id, status, has_voted)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		status.VoterID, status.ElectionID, status.Status, status.HasVoted,
	).Scan(&status.ID, &status.CreatedAt, &status.UpdatedAt)
}

func (r *PgRepository) UpdateElectionStatus(ctx context.Context, status *VoterElectionStatus) error {
	query := `
		UPDATE voter_status
		SET status = $3, has_voted = $4, voted_at = $5, voted_via = $6, updated_at = NOW()
		WHERE voter_id = $1 AND election_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		status.VoterID, status.ElectionID, status.Status,
		status.HasVoted, status.VotedAt, status.VotedVia,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrStatusNotFound
	}

	return nil
}

// Profile-specific methods

func (r *PgRepository) GetCompleteProfile(ctx context.Context, voterID int64, userID int64) (*CompleteProfileResponse, error) {
	query := `
		WITH voter_info AS (
			SELECT 
				v.id as voter_id,
				v.name,
				ua.username,
				v.email,
				v.phone,
				v.faculty_name,
				v.study_program_name,
				v.cohort_year,
				v.photo_url,
				v.voter_type,
				v.voting_method_preference,
				ua.created_at,
				ua.last_login_at,
				COALESCE(ua.login_count, 0) as login_count,
				CASE WHEN ua.is_active THEN 'active' ELSE 'inactive' END as account_status
			FROM voters v
			INNER JOIN user_accounts ua ON ua.voter_id = v.id
			WHERE v.id = $1 AND ua.id = $2
		),
		voting_info AS (
			SELECT 
				vs.has_voted,
				vs.voted_at,
				vs.voting_method::text as method,
				t.name as tps_name,
				t.location as tps_location
			FROM voter_status vs
			LEFT JOIN tps t ON vs.tps_id = t.id
			WHERE vs.voter_id = $1
			ORDER BY vs.election_id DESC
			LIMIT 1
		),
		participation AS (
			SELECT 
				COUNT(*) as total_elections,
				COUNT(CASE WHEN vs.has_voted THEN 1 END) as participated,
				MAX(vs.voted_at) as last_participation
			FROM voter_status vs
			WHERE vs.voter_id = $1
		)
		SELECT 
			vi.voter_id, vi.name, vi.username, vi.email, vi.phone,
			vi.faculty_name, vi.study_program_name, vi.cohort_year, vi.photo_url, vi.voter_type,
			COALESCE(voti.method, vi.voting_method_preference) as preferred_method,
			COALESCE(voti.has_voted, false) as has_voted,
			voti.voted_at,
			voti.tps_name,
			voti.tps_location,
			COALESCE(p.total_elections, 0) as total_elections,
			COALESCE(p.participated, 0) as participated,
			p.last_participation,
			vi.created_at,
			vi.last_login_at,
			vi.login_count,
			vi.account_status
		FROM voter_info vi
		LEFT JOIN voting_info voti ON true
		LEFT JOIN participation p ON true
	`

	var response CompleteProfileResponse
	var semester string

	err := r.db.QueryRow(ctx, query, voterID, userID).Scan(
		&response.PersonalInfo.VoterID,
		&response.PersonalInfo.Name,
		&response.PersonalInfo.Username,
		&response.PersonalInfo.Email,
		&response.PersonalInfo.Phone,
		&response.PersonalInfo.FacultyName,
		&response.PersonalInfo.StudyProgramName,
		&response.PersonalInfo.CohortYear,
		&response.PersonalInfo.PhotoURL,
		&response.PersonalInfo.VoterType,
		&response.VotingInfo.PreferredMethod,
		&response.VotingInfo.HasVoted,
		&response.VotingInfo.VotedAt,
		&response.VotingInfo.TPSName,
		&response.VotingInfo.TPSLocation,
		&response.Participation.TotalElections,
		&response.Participation.ParticipatedElections,
		&response.Participation.LastParticipation,
		&response.AccountInfo.CreatedAt,
		&response.AccountInfo.LastLogin,
		&response.AccountInfo.LoginCount,
		&response.AccountInfo.AccountStatus,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVoterNotFound
		}
		return nil, err
	}

	// Calculate semester
	if response.PersonalInfo.CohortYear != nil && *response.PersonalInfo.CohortYear > 0 {
		currentYear := time.Now().Year()
		yearsEnrolled := currentYear - *response.PersonalInfo.CohortYear
		semester = fmt.Sprintf("%d", yearsEnrolled*2+1)
	} else {
		semester = "-"
	}
	response.PersonalInfo.Semester = semester

	// Calculate participation rate
	if response.Participation.TotalElections > 0 {
		response.Participation.ParticipationRate = float64(response.Participation.ParticipatedElections) / float64(response.Participation.TotalElections) * 100
	}

	return &response, nil
}

func (r *PgRepository) UpdateProfile(ctx context.Context, voterID int64, req *UpdateProfileRequest) error {
	query := `
		UPDATE voters
		SET email = COALESCE($2, email),
		    phone = COALESCE($3, phone),
		    photo_url = COALESCE($4, photo_url),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, voterID, req.Email, req.Phone, req.PhotoURL)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrVoterNotFound
	}

	return nil
}

func (r *PgRepository) UpdateVotingMethod(ctx context.Context, voterID, electionID int64, method string) error {
	// Check if voter has already voted
	checkQuery := `
		SELECT has_voted, voting_method
		FROM voter_status
		WHERE voter_id = $1 AND election_id = $2
	`

	var hasVoted bool
	var votingMethod *string

	err := r.db.QueryRow(ctx, checkQuery, voterID, electionID).Scan(&hasVoted, &votingMethod)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if hasVoted {
		return ErrAlreadyVoted
	}

	// Check if already checked in at TPS (voting_method is set to TPS)
	if votingMethod != nil && *votingMethod == "TPS" && method == "ONLINE" {
		return ErrAlreadyCheckedIn
	}

	// Update voting method preference
	updateQuery := `
		UPDATE voters
		SET voting_method_preference = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, updateQuery, voterID, method)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrVoterNotFound
	}

	return nil
}

func (r *PgRepository) GetParticipationStats(ctx context.Context, voterID int64) (*ParticipationStatsResponse, error) {
	// Get summary
	summaryQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN vs.has_voted THEN 1 END) as participated
		FROM elections e
		LEFT JOIN voter_status vs ON vs.election_id = e.id AND vs.voter_id = $1
		WHERE e.status IN ('CLOSED', 'ARCHIVED', 'VOTING_OPEN')
	`

	var total, participated int
	err := r.db.QueryRow(ctx, summaryQuery, voterID).Scan(&total, &participated)
	if err != nil {
		return nil, err
	}

	// Get elections list
	electionsQuery := `
		SELECT 
			e.id,
			e.name,
			e.year,
			COALESCE(vs.has_voted, false),
			vs.voted_at,
			COALESCE(vs.voting_method::text, 'NONE')
		FROM elections e
		LEFT JOIN voter_status vs ON vs.election_id = e.id AND vs.voter_id = $1
		WHERE e.status IN ('CLOSED', 'ARCHIVED', 'VOTING_OPEN')
		ORDER BY e.year DESC
	`

	rows, err := r.db.Query(ctx, electionsQuery, voterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elections []ElectionParticipation
	for rows.Next() {
		var ep ElectionParticipation
		err := rows.Scan(&ep.ElectionID, &ep.ElectionName, &ep.Year, &ep.Voted, &ep.VotedAt, &ep.Method)
		if err != nil {
			return nil, err
		}
		elections = append(elections, ep)
	}

	response := &ParticipationStatsResponse{
		Summary: ParticipationSummary{
			TotalElections:  total,
			Participated:    participated,
			NotParticipated: total - participated,
		},
		Elections: elections,
	}

	if total > 0 {
		response.Summary.ParticipationRate = float64(participated) / float64(total) * 100
	}

	return response, nil
}

func (r *PgRepository) DeletePhoto(ctx context.Context, voterID int64) error {
	query := `
		UPDATE voters
		SET photo_url = NULL, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, voterID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrVoterNotFound
	}

	return nil
}

var (
	ErrVoterNotFound    = errors.New("voter not found")
	ErrStatusNotFound   = errors.New("voter status not found")
	ErrAlreadyVoted     = errors.New("voter has already voted")
	ErrAlreadyCheckedIn = errors.New("voter has already checked in at TPS")
)
