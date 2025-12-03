package electionvoter

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pemira-api/internal/shared"
)

type pgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) Repository {
	return &pgRepository{db: db}
}

const (
	identityColumnStudent  = "student_id"
	identityColumnLecturer = "lecturer_id"
	identityColumnStaff    = "staff_id"
)

type identityRefs struct {
	Column     string
	StudentID  *int64
	LecturerID *int64
	StaffID    *int64
}

func (r *pgRepository) LookupByNIM(ctx context.Context, electionID int64, nim string) (*LookupResult, error) {
	query := `
		SELECT
			v.id,
			v.nim,
			v.name,
			v.voter_type,
			v.email,
			v.faculty_code,
			v.study_program_code,
			v.cohort_year,
			v.academic_status,
			v.lecturer_id,
			v.staff_id,
			v.voting_method,
			(ua.id IS NOT NULL) AS has_account,
			ev.id,
			ev.status,
			ev.voting_method,
			ev.tps_id,
			ev.checked_in_at,
			ev.voted_at,
			ev.updated_at,
			v.name
		FROM voters v
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		LEFT JOIN election_voters ev ON ev.voter_id = v.id AND ev.election_id = $2
		WHERE v.nim = $1
		LIMIT 1;
	`

	var res LookupResult
	var (
		evID           *int64
		evStatus       *string
		evMethod       *string
		evTPSID        *int64
		evCheckedIn    *time.Time
		evVotedAt      *time.Time
		evUpdatedAt    *time.Time
		email          sql.NullString
		facultyCode    sql.NullString
		studyProgram   sql.NullString
		cohortYear     sql.NullInt32
		academicStatus sql.NullString
		lecturerID     sql.NullInt64
		staffID        sql.NullInt64
		votingMethod   sql.NullString
	)

	var voter VoterSummary
	err := r.db.QueryRow(ctx, query, nim, electionID).Scan(
		&voter.ID,
		&voter.NIM,
		&voter.Name,
		&voter.VoterType,
		&email,
		&facultyCode,
		&studyProgram,
		&cohortYear,
		&academicStatus,
		&lecturerID,
		&staffID,
		&votingMethod,
		&voter.HasAccount,
		&evID,
		&evStatus,
		&evMethod,
		&evTPSID,
		&evCheckedIn,
		&evVotedAt,
		&evUpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("lookup voter: %w", err)
	}

	res.Voter = &voter
	voter.Email = nullableStringPtr(email)
	voter.FacultyCode = nullableStringPtr(facultyCode)
	voter.StudyProgramCode = nullableStringPtr(studyProgram)
	voter.CohortYear = nullableIntPtr(cohortYear)
	voter.AcademicStatus = nullableStringPtr(academicStatus)
	voter.LecturerID = nullableInt64Ptr(lecturerID)
	voter.StaffID = nullableInt64Ptr(staffID)
	voter.VotingMethod = nullableStringPtr(votingMethod)

	if evID != nil && evStatus != nil && evMethod != nil && evUpdatedAt != nil {
		res.ElectionVoter = &ElectionVoter{
			ID:           *evID,
			ElectionID:   electionID,
			VoterID:      voter.ID,
			NIM:          voter.NIM,
			Status:       *evStatus,
			VotingMethod: *evMethod,
			TPSID:        evTPSID,
			CheckedInAt:  evCheckedIn,
			VotedAt:      evVotedAt,
			UpdatedAt:    *evUpdatedAt,
			VoterType:    voter.VoterType,
			Name:         voter.Name,
			Email:        voter.Email,
			FacultyCode:  voter.FacultyCode,
			StudyProgram: voter.StudyProgramCode,
			CohortYear:   voter.CohortYear,
		}
	}

	return &res, nil
}

func (r *pgRepository) UpsertAndEnroll(ctx context.Context, electionID int64, in UpsertAndEnrollInput) (*UpsertAndEnrollResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	nim := strings.TrimSpace(in.NIM)
	voterType := strings.ToUpper(strings.TrimSpace(in.VoterType))
	if voterType == "" {
		voterType = "STUDENT"
	}

	votingMethod := strings.ToUpper(strings.TrimSpace(in.VotingMethod))
	if votingMethod == "" {
		votingMethod = "ONLINE"
	}

	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status == "" {
		status = "PENDING"
	}

	identity := identityRefs{}
	switch voterType {
	case "STUDENT":
		studentID, err := r.upsertStudentIdentity(ctx, tx, nim, in)
		if err != nil {
			return nil, err
		}
		identity.Column = identityColumnStudent
		identity.StudentID = &studentID
	case "LECTURER":
		lecturerID := in.LecturerID
		if lecturerID == nil {
			foundID, err := r.lookupLecturerID(ctx, tx, nim)
			if err != nil {
				return nil, err
			}
			if foundID == nil {
				return nil, shared.ErrBadRequest
			}
			lecturerID = foundID
		}
		identity.Column = identityColumnLecturer
		identity.LecturerID = lecturerID
	case "STAFF":
		staffID := in.StaffID
		if staffID == nil {
			foundID, err := r.lookupStaffID(ctx, tx, nim)
			if err != nil {
				return nil, err
			}
			if foundID == nil {
				return nil, shared.ErrBadRequest
			}
			staffID = foundID
		}
		identity.Column = identityColumnStaff
		identity.StaffID = staffID
	default:
		identity.Column = ""
	}

	voterID, createdVoter, err := r.saveVoterRecord(ctx, tx, nim, voterType, votingMethod, in, identity)
	if err != nil {
		return nil, err
	}

	var ev ElectionVoter
	var createdEnrollment bool
	qEnroll := `
		INSERT INTO election_voters (
			election_id, voter_id, nim, status, voting_method, tps_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
		ON CONFLICT ON CONSTRAINT ux_election_voters_election_voter DO UPDATE SET
			status = EXCLUDED.status,
			voting_method = EXCLUDED.voting_method,
			tps_id = EXCLUDED.tps_id,
			nim = EXCLUDED.nim,
			updated_at = NOW()
		RETURNING id, (xmax = 0) AS is_insert, status, voting_method, tps_id, checked_in_at, voted_at, updated_at;
	`

	err = tx.QueryRow(ctx, qEnroll, electionID, voterID, nim, status, votingMethod, in.TPSID).
		Scan(&ev.ID, &createdEnrollment, &ev.Status, &ev.VotingMethod, &ev.TPSID, &ev.CheckedInAt, &ev.VotedAt, &ev.UpdatedAt)
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" && pgerr.ConstraintName == "ux_election_voters_election_nim" {
			return nil, shared.ErrDuplicateEntry
		}
		return nil, fmt.Errorf("upsert election_voters: %w", err)
	}

	ev.ElectionID = electionID
	ev.VoterID = voterID
	ev.NIM = nim
	ev.VoterType = voterType
	ev.Name = in.Name

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &UpsertAndEnrollResult{
		VoterID:             voterID,
		ElectionVoterID:     ev.ID,
		Status:              ev.Status,
		VotingMethod:        ev.VotingMethod,
		TPSID:               ev.TPSID,
		CreatedVoter:        createdVoter,
		CreatedEnrollment:   createdEnrollment,
		DuplicateInElection: false,
	}, nil
}

func (r *pgRepository) saveVoterRecord(
	ctx context.Context,
	tx pgx.Tx,
	nim string,
	voterType string,
	votingMethod string,
	in UpsertAndEnrollInput,
	identity identityRefs,
) (int64, bool, error) {
	existingID, err := r.findExistingVoter(ctx, tx, identity, nim)
	if err != nil {
		return 0, false, err
	}

	status := resolveAcademicStatus(in.AcademicStatus)

	if existingID != nil {
		const updateQuery = `
		UPDATE voters
		SET
			nim = $2,
			name = $3,
			email = $4,
			phone = $5,
			faculty_code = $6,
			faculty_name = $7,
			study_program_code = $8,
			study_program_name = $9,
			cohort_year = $10,
			semester = $11,
			academic_status = $12,
			voter_type = $13,
			student_id = $14,
			lecturer_id = $15,
			staff_id = $16,
			voting_method = $17,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id;
		`

		var voterID int64
		if err := tx.QueryRow(ctx, updateQuery,
			*existingID,
			nim,
			in.Name,
			in.Email,
			in.Phone,
			in.FacultyCode,
			in.FacultyName,
			in.StudyProgramCode,
			in.StudyProgramName,
			in.CohortYear,
			in.Semester,
			status,
			voterType,
			identity.StudentID,
			identity.LecturerID,
			identity.StaffID,
			votingMethod,
		).Scan(&voterID); err != nil {
			return 0, false, fmt.Errorf("update voter: %w", err)
		}
		return voterID, false, nil
	}

	const insertQuery = `
	INSERT INTO voters (
		nim, name, email, phone,
		faculty_code, faculty_name,
		study_program_code, study_program_name,
		cohort_year, semester, academic_status,
		voter_type, student_id, lecturer_id, staff_id,
		voting_method, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4,
		$5, $6,
		$7, $8,
		$9, $10, $11,
		$12, $13, $14, $15,
		$16, NOW(), NOW()
	)
	RETURNING id;
	`

	var voterID int64
	if err := tx.QueryRow(ctx, insertQuery,
		nim,
		in.Name,
		in.Email,
		in.Phone,
		in.FacultyCode,
		in.FacultyName,
		in.StudyProgramCode,
		in.StudyProgramName,
		in.CohortYear,
		in.Semester,
		status,
		voterType,
		identity.StudentID,
		identity.LecturerID,
		identity.StaffID,
		votingMethod,
	).Scan(&voterID); err != nil {
		return 0, false, fmt.Errorf("insert voter: %w", err)
	}
	return voterID, true, nil
}

func (r *pgRepository) findExistingVoter(ctx context.Context, tx pgx.Tx, identity identityRefs, nim string) (*int64, error) {
	if col, idPtr, ok := identity.columnAndID(); ok {
		query := fmt.Sprintf(`SELECT id FROM voters WHERE %s = $1 LIMIT 1`, col)
		var existing int64
		err := tx.QueryRow(ctx, query, *idPtr).Scan(&existing)
		if err == nil {
			return &existing, nil
		}
		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("find voter by %s: %w", col, err)
		}
	}

	var voterID int64
	err := tx.QueryRow(ctx, `SELECT id FROM voters WHERE nim = $1 ORDER BY id DESC LIMIT 1`, nim).Scan(&voterID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find voter by nim: %w", err)
	}

	if col, idPtr, ok := identity.columnAndID(); ok {
		if _, execErr := tx.Exec(ctx, fmt.Sprintf(`UPDATE voters SET %s = $1 WHERE id = $2`, col), *idPtr, voterID); execErr != nil {
			return nil, fmt.Errorf("attach %s to voter: %w", col, execErr)
		}
	}

	return &voterID, nil
}

func (r *pgRepository) upsertStudentIdentity(ctx context.Context, tx pgx.Tx, nim string, in UpsertAndEnrollInput) (int64, error) {
	const query = `
	INSERT INTO students (nim, name, faculty_code, program_code, cohort_year, class_label, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	ON CONFLICT (nim) DO UPDATE SET
		name = EXCLUDED.name,
		faculty_code = EXCLUDED.faculty_code,
		program_code = EXCLUDED.program_code,
		cohort_year = EXCLUDED.cohort_year,
		class_label = EXCLUDED.class_label,
		updated_at = NOW()
	RETURNING id;
	`

	var studentID int64
	if err := tx.QueryRow(ctx, query,
		nim,
		in.Name,
		in.FacultyCode,
		in.StudyProgramCode,
		in.CohortYear,
		nil,
	).Scan(&studentID); err != nil {
		return 0, fmt.Errorf("upsert student identity: %w", err)
	}
	return studentID, nil
}

func (r *pgRepository) lookupLecturerID(ctx context.Context, tx pgx.Tx, nidn string) (*int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `SELECT id FROM lecturers WHERE nidn = $1 LIMIT 1`, nidn).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find lecturer by nidn: %w", err)
	}
	return &id, nil
}

func (r *pgRepository) lookupStaffID(ctx context.Context, tx pgx.Tx, nip string) (*int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `SELECT id FROM staff_members WHERE nip = $1 LIMIT 1`, nip).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find staff by nip: %w", err)
	}
	return &id, nil
}

func (i identityRefs) columnAndID() (string, *int64, bool) {
	switch i.Column {
	case identityColumnStudent:
		if i.StudentID == nil {
			return "", nil, false
		}
		return identityColumnStudent, i.StudentID, true
	case identityColumnLecturer:
		if i.LecturerID == nil {
			return "", nil, false
		}
		return identityColumnLecturer, i.LecturerID, true
	case identityColumnStaff:
		if i.StaffID == nil {
			return "", nil, false
		}
		return identityColumnStaff, i.StaffID, true
	default:
		return "", nil, false
	}
}

func resolveAcademicStatus(raw *string) string {
	if raw == nil {
		return defaultAcademicStatus
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return defaultAcademicStatus
	}
	return strings.ToUpper(trimmed)
}

func (r *pgRepository) List(ctx context.Context, electionID int64, filter ListFilter, pag shared.PaginationParams) ([]ElectionVoter, int64, error) {
	var args []interface{}
	where := []string{"ev.election_id = $1"}
	args = append(args, electionID)

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(ev.nim ILIKE $%d OR v.name ILIKE $%d)", len(args)+1, len(args)+2))
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.VoterType != "" {
		where = append(where, fmt.Sprintf("v.voter_type = $%d", len(args)+1))
		args = append(args, filter.VoterType)
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("ev.status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.VotingMethod != "" {
		where = append(where, fmt.Sprintf("ev.voting_method = $%d", len(args)+1))
		args = append(args, filter.VotingMethod)
	}
	if filter.FacultyCode != "" {
		where = append(where, fmt.Sprintf("v.faculty_code = $%d", len(args)+1))
		args = append(args, filter.FacultyCode)
	}
	if filter.StudyProgramCode != "" {
		where = append(where, fmt.Sprintf("v.study_program_code = $%d", len(args)+1))
		args = append(args, filter.StudyProgramCode)
	}
	if filter.CohortYear != nil {
		where = append(where, fmt.Sprintf("v.cohort_year = $%d", len(args)+1))
		args = append(args, *filter.CohortYear)
	}
	if filter.TPSID != nil {
		where = append(where, fmt.Sprintf("ev.tps_id = $%d", len(args)+1))
		args = append(args, *filter.TPSID)
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM election_voters ev
		JOIN voters v ON v.id = ev.voter_id
		%s
	`, whereClause)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count election_voters: %w", err)
	}

	args = append(args, pag.Limit(), pag.Offset())
	listQuery := fmt.Sprintf(`
		SELECT
			ev.id, ev.election_id, ev.voter_id, ev.nim,
			ev.status, ev.voting_method, ev.tps_id,
			ev.checked_in_at, ev.voted_at, ev.updated_at,
			v.voter_type, v.name, v.email, 
			v.faculty_code, v.faculty_name,
			v.study_program_code, v.study_program_name,
			v.cohort_year,
			COALESCE(v.semester, 
				CASE 
					WHEN v.cohort_year IS NOT NULL AND v.voter_type = 'STUDENT' 
					THEN (EXTRACT(YEAR FROM CURRENT_DATE)::int - v.cohort_year) * 2 + 1
					ELSE NULL
				END
			) AS semester,
			v.academic_status,
			vs.has_voted,
			ua.last_login_at
		FROM election_voters ev
		JOIN voters v ON v.id = ev.voter_id
		LEFT JOIN voter_status vs ON vs.election_id = ev.election_id AND vs.voter_id = ev.voter_id
		LEFT JOIN user_accounts ua ON ua.voter_id = v.id
		%s
		ORDER BY ev.updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)-1, len(args))

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list election_voters: %w", err)
	}
	defer rows.Close()

	var items []ElectionVoter
	for rows.Next() {
		var item ElectionVoter
		var email sql.NullString
		var facultyCode sql.NullString
		var facultyName sql.NullString
		var studyProgram sql.NullString
		var studyProgramName sql.NullString
		var cohortYear sql.NullInt32
		var semester sql.NullInt32
		var academicStatus sql.NullString
		var hasVoted sql.NullBool
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.ElectionID,
			&item.VoterID,
			&item.NIM,
			&item.Status,
			&item.VotingMethod,
			&item.TPSID,
			&item.CheckedInAt,
			&item.VotedAt,
			&item.UpdatedAt,
			&item.VoterType,
			&item.Name,
			&email,
			&facultyCode,
			&facultyName,
			&studyProgram,
			&studyProgramName,
			&cohortYear,
			&semester,
			&academicStatus,
			&hasVoted,
			&lastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan election_voters: %w", err)
		}
		item.Email = nullableStringPtr(email)
		item.FacultyCode = nullableStringPtr(facultyCode)
		item.FacultyName = nullableStringPtr(facultyName)
		item.StudyProgram = nullableStringPtr(studyProgram)
		item.StudyProgramName = nullableStringPtr(studyProgramName)
		item.CohortYear = nullableIntPtr(cohortYear)
		item.Semester = nullableIntPtr(semester)
		item.AcademicStatus = nullableStringPtr(academicStatus)
		item.HasVoted = nullableBoolPtr(hasVoted)
		item.LastLoginAt = nullableTimePtr(lastLoginAt)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return items, total, nil
}

func (r *pgRepository) UpdateEnrollment(ctx context.Context, electionID int64, enrollmentID int64, in UpdateInput) (*ElectionVoter, error) {
	// First, update election_voters table
	setParts := []string{}
	args := []interface{}{enrollmentID, electionID}

	if in.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *in.Status)
	}
	if in.VotingMethod != nil {
		setParts = append(setParts, fmt.Sprintf("voting_method = $%d", len(args)+1))
		args = append(args, *in.VotingMethod)
	}
	if in.TPSID != nil {
		setParts = append(setParts, fmt.Sprintf("tps_id = $%d", len(args)+1))
		args = append(args, *in.TPSID)
	}

	if len(setParts) == 0 {
		// If no fields to update, just touch updated_at
		query := `
			UPDATE election_voters
			SET updated_at = NOW()
			WHERE id = $1 AND election_id = $2
			RETURNING id, election_id, voter_id, nim, status, voting_method, tps_id, checked_in_at, voted_at, updated_at;
		`
		
		var ev ElectionVoter
		err := r.db.QueryRow(ctx, query, enrollmentID, electionID).Scan(
			&ev.ID,
			&ev.ElectionID,
			&ev.VoterID,
			&ev.NIM,
			&ev.Status,
			&ev.VotingMethod,
			&ev.TPSID,
			&ev.CheckedInAt,
			&ev.VotedAt,
			&ev.UpdatedAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, shared.ErrNotFound
			}
			return nil, fmt.Errorf("update election_voter: %w", err)
		}

		// If semester is provided, update voters table
		if in.Semester != nil {
			_, err := r.db.Exec(ctx, `
				UPDATE voters 
				SET semester = $1, updated_at = NOW()
				WHERE id = $2
			`, *in.Semester, ev.VoterID)
			if err != nil {
				return nil, fmt.Errorf("update voter semester: %w", err)
			}
		}

		return &ev, nil
	}

	query := fmt.Sprintf(`
		UPDATE election_voters
		SET %s, updated_at = NOW()
		WHERE id = $1 AND election_id = $2
		RETURNING id, election_id, voter_id, nim, status, voting_method, tps_id, checked_in_at, voted_at, updated_at;
	`, strings.Join(setParts, ", "))

	var ev ElectionVoter
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&ev.ID,
		&ev.ElectionID,
		&ev.VoterID,
		&ev.NIM,
		&ev.Status,
		&ev.VotingMethod,
		&ev.TPSID,
		&ev.CheckedInAt,
		&ev.VotedAt,
		&ev.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("update election_voter: %w", err)
	}

	// If semester is provided, update voters table
	if in.Semester != nil {
		_, err := r.db.Exec(ctx, `
			UPDATE voters 
			SET semester = $1, updated_at = NOW()
			WHERE id = $2
		`, *in.Semester, ev.VoterID)
		if err != nil {
			return nil, fmt.Errorf("update voter semester: %w", err)
		}
	}

	return &ev, nil
}

func (r *pgRepository) SelfRegister(ctx context.Context, electionID int64, voterID int64, in SelfRegisterInput) (*ElectionVoter, error) {
	var nim string
	err := r.db.QueryRow(ctx, `SELECT nim FROM voters WHERE id = $1`, voterID).Scan(&nim)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get voter nim: %w", err)
	}
	if strings.TrimSpace(nim) == "" {
		return nil, shared.ErrBadRequest
	}

	query := `
		INSERT INTO election_voters (election_id, voter_id, nim, status, voting_method, tps_id, created_at, updated_at)
		VALUES ($1, $2, $3, 'PENDING', $4, $5, NOW(), NOW())
		ON CONFLICT ON CONSTRAINT ux_election_voters_election_voter DO UPDATE SET
			voting_method = EXCLUDED.voting_method,
			tps_id = EXCLUDED.tps_id,
			updated_at = NOW()
		RETURNING id, election_id, voter_id, nim, status, voting_method, tps_id, checked_in_at, voted_at, updated_at;
	`

	var ev ElectionVoter
	err = r.db.QueryRow(ctx, query, electionID, voterID, nim, in.VotingMethod, in.TPSID).Scan(
		&ev.ID,
		&ev.ElectionID,
		&ev.VoterID,
		&ev.NIM,
		&ev.Status,
		&ev.VotingMethod,
		&ev.TPSID,
		&ev.CheckedInAt,
		&ev.VotedAt,
		&ev.UpdatedAt,
	)
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == "23505" && pgerr.ConstraintName == "ux_election_voters_election_nim" {
			return nil, shared.ErrDuplicateEntry
		}
		return nil, fmt.Errorf("self register election_voter: %w", err)
	}

	return &ev, nil
}

func (r *pgRepository) GetStatus(ctx context.Context, electionID int64, voterID int64) (*ElectionVoter, error) {
	query := `
		SELECT
			ev.id, ev.election_id, ev.voter_id, ev.nim,
			ev.status, ev.voting_method, ev.tps_id,
			ev.checked_in_at, ev.voted_at, ev.updated_at,
			v.voter_type, v.name, v.email, v.faculty_code, v.study_program_code, v.cohort_year
		FROM election_voters ev
		JOIN voters v ON v.id = ev.voter_id
		WHERE ev.election_id = $1 AND ev.voter_id = $2
		LIMIT 1;
	`

	var ev ElectionVoter
	var email sql.NullString
	var facultyCode sql.NullString
	var studyProgram sql.NullString
	var cohortYear sql.NullInt32

	err := r.db.QueryRow(ctx, query, electionID, voterID).Scan(
		&ev.ID,
		&ev.ElectionID,
		&ev.VoterID,
		&ev.NIM,
		&ev.Status,
		&ev.VotingMethod,
		&ev.TPSID,
		&ev.CheckedInAt,
		&ev.VotedAt,
		&ev.UpdatedAt,
		&ev.VoterType,
		&ev.Name,
		&email,
		&facultyCode,
		&studyProgram,
		&cohortYear,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get election_voter status: %w", err)
	}

	ev.Email = nullableStringPtr(email)
	ev.FacultyCode = nullableStringPtr(facultyCode)
	ev.StudyProgram = nullableStringPtr(studyProgram)
	ev.CohortYear = nullableIntPtr(cohortYear)

	return &ev, nil
}

func nullableStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		val := ns.String
		return &val
	}
	return nil
}

func nullableIntPtr(ns sql.NullInt32) *int {
	if ns.Valid {
		val := int(ns.Int32)
		return &val
	}
	return nil
}

func nullableInt64Ptr(ns sql.NullInt64) *int64 {
	if ns.Valid {
		val := ns.Int64
		return &val
	}
	return nil
}

func nullableBoolPtr(nb sql.NullBool) *bool {
	if nb.Valid {
		val := nb.Bool
		return &val
	}
	return nil
}

func nullableTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		val := nt.Time
		return &val
	}
	return nil
}
