package dpt

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) ImportVotersForElection(ctx context.Context, electionID int64, rows []ImportRow) (*ImportResult, error) {
	result := &ImportResult{
		TotalRows: len(rows),
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, row := range rows {
		// Upsert voter
		var voterID int64
		var isInsert bool

		qUpsert := `
			INSERT INTO voters (nim, name, faculty_name, study_program_name, cohort_year, email, phone)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (nim) DO UPDATE
			SET name = EXCLUDED.name,
			    faculty_name = EXCLUDED.faculty_name,
			    study_program_name = EXCLUDED.study_program_name,
			    cohort_year = EXCLUDED.cohort_year,
			    email = EXCLUDED.email,
			    phone = EXCLUDED.phone,
			    updated_at = NOW()
			RETURNING id, (xmax = 0) AS is_insert
		`

		err := tx.QueryRow(ctx, qUpsert,
			row.NIM,
			row.Name,
			row.FacultyName,
			row.StudyProgram,
			row.CohortYear,
			row.Email,
			row.Phone,
		).Scan(&voterID, &isInsert)

		if err != nil {
			return nil, fmt.Errorf("upsert voter %s: %w", row.NIM, err)
		}

		if isInsert {
			result.InsertedVoters++
		} else {
			result.UpdatedVoters++
		}

		// Insert voter_status if not exists
		qStatus := `
			INSERT INTO voter_status (election_id, voter_id, is_eligible, has_voted)
			VALUES ($1, $2, TRUE, FALSE)
			ON CONFLICT (election_id, voter_id) DO NOTHING
		`

		tag, err := tx.Exec(ctx, qStatus, electionID, voterID)
		if err != nil {
			return nil, fmt.Errorf("insert voter_status: %w", err)
		}

		if tag.RowsAffected() > 0 {
			result.CreatedStatus++
		} else {
			result.SkippedStatus++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil
}

func (r *pgxRepository) ListVotersForElection(ctx context.Context, electionID int64, filter ListFilter) ([]VoterWithStatusDTO, int64, error) {
	whereClause, args := buildWhereClause(electionID, filter)

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM voters v
		INNER JOIN voter_status vs ON vs.voter_id = v.id
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count voters: %w", err)
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT 
			v.id,
			v.nim,
			v.name,
			v.faculty_name,
			v.study_program_name,
			COALESCE(v.class_label, '') AS semester,
			v.cohort_year,
			COALESCE(v.email, ''),
			(ua.id IS NOT NULL) AS has_account,
			COALESCE(ua.role::TEXT, v.voter_type) AS voter_type,
			vs.is_eligible,
			vs.has_voted,
			vs.voted_at,
			vs.voting_method,
			vs.tps_id
		FROM voters v
		INNER JOIN voter_status vs ON vs.voter_id = v.id
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		%s
		ORDER BY v.nim
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query voters: %w", err)
	}
	defer rows.Close()

	var items []VoterWithStatusDTO
	for rows.Next() {
		var item VoterWithStatusDTO
		var semester string
		var voterType string
		err := rows.Scan(
			&item.VoterID,
			&item.NIM,
			&item.Name,
			&item.FacultyName,
			&item.StudyProgramName,
			&semester,
			&item.CohortYear,
			&item.Email,
			&item.HasAccount,
			&voterType,
			&item.Status.IsEligible,
			&item.Status.HasVoted,
			&item.Status.LastVoteAt,
			&item.Status.LastVoteChannel,
			&item.Status.LastTPSID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan voter: %w", err)
		}
		item.Semester = strings.TrimSpace(semester)
		item.VoterType = voterType // Always set voter_type
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return items, total, nil
}

func (r *pgxRepository) StreamVotersForElection(ctx context.Context, electionID int64, filter ListFilter, fn func(VoterWithStatusDTO) error) error {
	whereClause, args := buildWhereClause(electionID, filter)

	query := fmt.Sprintf(`
		SELECT 
			v.id,
			v.nim,
			v.name,
			v.faculty_name,
			v.study_program_name,
			COALESCE(v.class_label, '') AS semester,
			v.cohort_year,
			COALESCE(v.email, ''),
			(ua.id IS NOT NULL) AS has_account,
			COALESCE(ua.role::TEXT, v.voter_type) AS voter_type,
			vs.is_eligible,
			vs.has_voted,
			vs.voted_at,
			vs.voting_method,
			vs.tps_id
		FROM voters v
		INNER JOIN voter_status vs ON vs.voter_id = v.id
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		%s
		ORDER BY v.nim
	`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query voters: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item VoterWithStatusDTO
		var semester string
		var voterType string
		err := rows.Scan(
			&item.VoterID,
			&item.NIM,
			&item.Name,
			&item.FacultyName,
			&item.StudyProgramName,
			&semester,
			&item.CohortYear,
			&item.Email,
			&item.HasAccount,
			&voterType,
			&item.Status.IsEligible,
			&item.Status.HasVoted,
			&item.Status.LastVoteAt,
			&item.Status.LastVoteChannel,
			&item.Status.LastTPSID,
		)
		if err != nil {
			return fmt.Errorf("scan voter: %w", err)
		}
		item.Semester = strings.TrimSpace(semester)
		item.VoterType = voterType // Always set voter_type

		if err := fn(item); err != nil {
			return err
		}
	}

	return rows.Err()
}

func buildWhereClause(electionID int64, filter ListFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Always filter by election
	conditions = append(conditions, fmt.Sprintf("vs.election_id = $%d", argIdx))
	args = append(args, electionID)
	argIdx++

	if filter.Faculty != "" {
		conditions = append(conditions, fmt.Sprintf("v.faculty_name = $%d", argIdx))
		args = append(args, filter.Faculty)
		argIdx++
	}

	if filter.StudyProgram != "" {
		conditions = append(conditions, fmt.Sprintf("v.study_program_name = $%d", argIdx))
		args = append(args, filter.StudyProgram)
		argIdx++
	}

	if filter.CohortYear != nil {
		conditions = append(conditions, fmt.Sprintf("v.cohort_year = $%d", argIdx))
		args = append(args, *filter.CohortYear)
		argIdx++
	}

	if filter.HasVoted != nil {
		conditions = append(conditions, fmt.Sprintf("vs.has_voted = $%d", argIdx))
		args = append(args, *filter.HasVoted)
		argIdx++
	}

	if filter.Eligible != nil {
		conditions = append(conditions, fmt.Sprintf("vs.is_eligible = $%d", argIdx))
		args = append(args, *filter.Eligible)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(v.nim ILIKE $%d OR v.name ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *pgxRepository) GetVoterByID(ctx context.Context, electionID int64, voterID int64) (*VoterWithStatusDTO, error) {
	query := `
		SELECT 
			v.id,
			v.nim,
			v.name,
			v.faculty_name,
			v.study_program_name,
			COALESCE(v.class_label, '') AS semester,
			v.cohort_year,
			COALESCE(v.email, ''),
			(ua.id IS NOT NULL) AS has_account,
			COALESCE(ua.role::TEXT, v.voter_type) AS voter_type,
			vs.is_eligible,
			vs.has_voted,
			vs.voted_at,
			vs.voting_method,
			vs.tps_id
		FROM voters v
		INNER JOIN voter_status vs ON vs.voter_id = v.id
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		WHERE v.id = $1 AND vs.election_id = $2
	`

	var item VoterWithStatusDTO
	var semester string
	var voterType string // Changed from *string to string
	
	err := r.db.QueryRow(ctx, query, voterID, electionID).Scan(
		&item.VoterID,
		&item.NIM,
		&item.Name,
		&item.FacultyName,
		&item.StudyProgramName,
		&semester,
		&item.CohortYear,
		&item.Email,
		&item.HasAccount,
		&voterType,
		&item.Status.IsEligible,
		&item.Status.HasVoted,
		&item.Status.LastVoteAt,
		&item.Status.LastVoteChannel,
		&item.Status.LastTPSID,
	)
	
	if err != nil {
		return nil, fmt.Errorf("get voter: %w", err)
	}

	item.Semester = strings.TrimSpace(semester)
	item.VoterType = voterType // Always set voter_type

	return &item, nil
}

func (r *pgxRepository) UpdateVoter(ctx context.Context, electionID int64, voterID int64, updates VoterUpdateDTO) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if voter exists for this election
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM voter_status WHERE voter_id = $1 AND election_id = $2)`
	if err := tx.QueryRow(ctx, checkQuery, voterID, electionID).Scan(&exists); err != nil {
		return fmt.Errorf("check voter exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("voter not found in this election")
	}

	// Check if voter has voted
	var hasVoted bool
	checkVotedQuery := `SELECT has_voted FROM voter_status WHERE voter_id = $1 AND election_id = $2`
	if err := tx.QueryRow(ctx, checkVotedQuery, voterID, electionID).Scan(&hasVoted); err != nil {
		return fmt.Errorf("check voter status: %w", err)
	}
	
	// If voter has voted, only allow voter_type update
	if hasVoted {
		// Check if trying to update fields other than voter_type
		if updates.Name != nil || updates.FacultyName != nil || updates.StudyProgram != nil ||
			updates.CohortYear != nil || updates.Email != nil || updates.Phone != nil || updates.IsEligible != nil {
			return fmt.Errorf("cannot update voter who has already voted")
		}
		// Only voter_type can be updated, continue...
	}

	// Update voter table
	var voterUpdates []string
	var voterArgs []interface{}
	argIdx := 1

	if updates.Name != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("name = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.Name)
		argIdx++
	}
	if updates.FacultyName != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("faculty_name = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.FacultyName)
		argIdx++
	}
	if updates.StudyProgram != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("study_program_name = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.StudyProgram)
		argIdx++
	}
	if updates.CohortYear != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("cohort_year = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.CohortYear)
		argIdx++
	}
	if updates.Email != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("email = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.Email)
		argIdx++
	}
	if updates.Phone != nil {
		voterUpdates = append(voterUpdates, fmt.Sprintf("phone = $%d", argIdx))
		voterArgs = append(voterArgs, *updates.Phone)
		argIdx++
	}

	if len(voterUpdates) > 0 {
		voterUpdates = append(voterUpdates, "updated_at = NOW()")
		voterArgs = append(voterArgs, voterID)
		updateVoterQuery := fmt.Sprintf(
			"UPDATE voters SET %s WHERE id = $%d",
			strings.Join(voterUpdates, ", "),
			argIdx,
		)
		if _, err := tx.Exec(ctx, updateVoterQuery, voterArgs...); err != nil {
			return fmt.Errorf("update voter: %w", err)
		}
	}

	// Update user_accounts role (voter_type) if provided
	if updates.VoterType != nil {
		// Validate voter_type
		validTypes := map[string]bool{"STUDENT": true, "LECTURER": true, "STAFF": true}
		if !validTypes[*updates.VoterType] {
			return fmt.Errorf("invalid voter_type: must be STUDENT, LECTURER, or STAFF")
		}
		
		// Update voters.voter_type (primary source)
		updateVoterTypeQuery := `
			UPDATE voters 
			SET voter_type = $1, updated_at = NOW()
			WHERE id = $2
		`
		if _, err := tx.Exec(ctx, updateVoterTypeQuery, *updates.VoterType, voterID); err != nil {
			return fmt.Errorf("update voter type in voters: %w", err)
		}
		
		// Update user_accounts.role if account exists (keep in sync)
		updateRoleQuery := `
			UPDATE user_accounts 
			SET role = $1, updated_at = NOW()
			WHERE voter_id = $2
		`
		// Ignore error if account doesn't exist
		_, _ = tx.Exec(ctx, updateRoleQuery, *updates.VoterType, voterID)
	}

	// Update voter_status if is_eligible is provided
	if updates.IsEligible != nil {
		updateStatusQuery := `
			UPDATE voter_status 
			SET is_eligible = $1, updated_at = NOW()
			WHERE voter_id = $2 AND election_id = $3
		`
		if _, err := tx.Exec(ctx, updateStatusQuery, *updates.IsEligible, voterID, electionID); err != nil {
			return fmt.Errorf("update voter status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *pgxRepository) DeleteVoter(ctx context.Context, electionID int64, voterID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if voter has voted
	var hasVoted bool
	checkQuery := `SELECT has_voted FROM voter_status WHERE voter_id = $1 AND election_id = $2`
	if err := tx.QueryRow(ctx, checkQuery, voterID, electionID).Scan(&hasVoted); err != nil {
		return fmt.Errorf("voter not found in this election")
	}
	
	if hasVoted {
		return fmt.Errorf("cannot delete voter who has already voted")
	}

	// Delete voter_status first (FK constraint)
	deleteStatusQuery := `DELETE FROM voter_status WHERE voter_id = $1 AND election_id = $2`
	if _, err := tx.Exec(ctx, deleteStatusQuery, voterID, electionID); err != nil {
		return fmt.Errorf("delete voter status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
