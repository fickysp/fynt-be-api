package submissions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type submissionsRepo struct {
	db *sqlx.DB
}

type SubmissionRepository interface {
	CreateSubmissionTx(ctx context.Context, req ApplyJobRequest, applicantID uuid.UUID, accountID string, cost int64) error
	GetSeekerSubmissions(ctx context.Context, applicantID string, statusFilter string, keyword string) ([]DTOSeekerSubmissionResponse, error)
	GetSubmissionDetail(ctx context.Context, submissionID, applicantID string) (*DTOSeekerSubmissionDetail, error)
	GetSubmissionsByJobID(ctx context.Context, jobID, compID, statusFilter string) ([]DTORecruiterSubmissionList, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID, compID, status string) error
	UpdateInternalNotes(ctx context.Context, submissionID, compID, notes string) error
	GetSubmissionsEssay(ctx context.Context, jobID string) ([]RawSubmissionEssayRow, error)
}

func NewSubmissionRepository(db *sqlx.DB) SubmissionRepository {
	return &submissionsRepo{db: db}
}

func (r *submissionsRepo) CreateSubmissionTx(ctx context.Context, req ApplyJobRequest, applicantID uuid.UUID, accountID string, cost int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("gagal memulai transaksi apply job", "error", err)
		return err
	}
	defer tx.Rollback()

	queryJob := `
    UPDATE jobs 
    SET 
        filled_slot = filled_slot + 1,
        status = CASE 
            WHEN slot IS NOT NULL AND (filled_slot + 1) >= slot THEN 'close'
            ELSE status
        END,
        updated_at = CURRENT_TIMESTAMP 
    WHERE id = $1 AND status = 'active' AND (slot IS NULL OR filled_slot < slot)`
	res, err := tx.ExecContext(ctx, queryJob, req.JobID)
	if err != nil {
		slog.Error("gagal update slot job", "error", err, "job_id", req.JobID)
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Kuota pelamar sudah penuh")
	}

	if cost > 0 {
		_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2", cost, accountID)
		if err != nil {
			slog.Error("gagal mengurangi saldo wallet", "error", err, "account_id", accountID, "cost", cost)
			return err
		}
	}

	var submissionID uuid.UUID
	query := `INSERT INTO submissions (applicant_id, job_id, web_portfolio_url) VALUES ($1, $2, $3) RETURNING id`
	if err = tx.QueryRowContext(ctx, query, applicantID, req.JobID, req.WebPortfolioURL).Scan(&submissionID); err != nil {
		slog.Error("gagal menyimpan submission", "profile_id", applicantID.String(), "job_id", req.JobID, "error", err)
		return err
	}

	if len(req.Files) > 0 {
		var files []SubmissionFile
		for _, f := range req.Files {
			files = append(files, SubmissionFile{
				ID:            uuid.New(),
				SubmissionsID: submissionID,
				FileType:      f.FileType,
				FileName:      f.FileName,
				FileSizeBytes: f.FileSizeBytes,
				FileURL:       f.FileURL,
			})
		}
		fileQuery := `INSERT INTO submissions_file (id, submissions_id, file_type, file_name, file_size_bytes, file_url) VALUES (:id, :submissions_id, :file_type, :file_name, :file_size_bytes, :file_url)`
		if _, err = tx.NamedExecContext(ctx, fileQuery, files); err != nil {
			slog.Error("gagal menyimpan berkas submission", "submission_id", submissionID.String(), "error", err)
			return err
		}
	}

	if len(req.Answers) > 0 {
		var answers []EssayAnswers
		for _, a := range req.Answers {
			qID, err := uuid.Parse(a.QuestionID)
			if err != nil {
				slog.Error("format question_id tidak valid", "question_id", a.QuestionID, "error", err)
				return err
			}
			answers = append(answers, EssayAnswers{
				SubmissionsID:   submissionID,
				EssayQuestionID: qID,
				Answer:          a.Answer,
			})
		}
		answerQuery := `INSERT INTO essay_answer (submissions_id, essay_question_id, answer) VALUES (:submissions_id, :essay_question_id, :answer)`
		if _, err = tx.NamedExecContext(ctx, answerQuery, answers); err != nil {
			slog.Error("gagal menyimpan jawaban essay", "submission_id", submissionID.String(), "error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("gagal commit transaksi apply job", "error", err)
		return err
	}

	return nil
}

func (r *submissionsRepo) GetSeekerSubmissions(ctx context.Context, applicantID string, statusFilter string, keyword string) ([]DTOSeekerSubmissionResponse, error) {
	var submissions []DTOSeekerSubmissionResponse
	query := `
		SELECT 
			s.id, j.position, c.company_name, c.domicile AS location, j.workplace, s.status, s.applied_at
		FROM submissions s
		JOIN jobs j ON s.job_id = j.id
		JOIN companies c ON j.company_id = c.id
		WHERE s.applicant_id = $1
	`
	args := []interface{}{applicantID}
	argCount := 2

	switch statusFilter {
	case "accepted":
		query += " AND s.status = 'accepted'"
	case "reviewed":
		query += " AND s.status = 'reviewed'"
	case "rejected":
		query += " AND s.status = 'rejected'"
	}

	if keyword != "" {
		query += fmt.Sprintf(" AND (j.position ILIKE $%d OR c.company_name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+keyword+"%")
	}

	query += " ORDER BY s.applied_at DESC"

	err := r.db.SelectContext(ctx, &submissions, query, args...)
	if err != nil {
		slog.Error("gagal mengambil data submissions seeker", "applicant_id", applicantID, "error", err)
		return nil, err
	}
	return submissions, nil
}

func (r *submissionsRepo) GetSubmissionDetail(ctx context.Context, submissionID, applicantID string) (*DTOSeekerSubmissionDetail, error) {
	var detail DTOSeekerSubmissionDetail

	querySub := `
		SELECT 
			s.id,
			s.status,
			s.applied_at,
			s.reviewed_at,
			s.decided_at,
			s.web_portfolio_url,
			j.position,
			j.workplace,
			j.employment,
			c.company_name,
			c.domicile AS location,
			c.avatar_url AS company_avatar
		FROM submissions s
		JOIN jobs j ON s.job_id = j.id
		JOIN companies c ON j.company_id = c.id
		WHERE s.id = $1 AND s.applicant_id = $2
	`

	var (
		subID         uuid.UUID
		webPort       sql.NullString
		companyAvatar sql.NullString
		reviewedAt    sql.NullTime
		decidedAt     sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, querySub, submissionID, applicantID).Scan(
		&subID,
		&detail.Status,
		&detail.AppliedAt,
		&reviewedAt,
		&decidedAt,
		&webPort,
		&detail.Position,
		&detail.WorkPlace,
		&detail.Employment,
		&detail.CompanyName,
		&detail.Location,
		&companyAvatar,
	)
	if err != nil {
		slog.Error("gagal mengambil data detail submission seeker", "submission_id", submissionID, "applicant_id", applicantID, "error", err)
		return nil, err
	}

	detail.SubmissionID = subID.String()
	if webPort.Valid && webPort.String != "" {
		detail.PortfolioLinkWeb = &webPort.String
	}
	if companyAvatar.Valid && companyAvatar.String != "" {
		detail.CompanyAvatar = &companyAvatar.String
	}
	if reviewedAt.Valid {
		detail.ReviewedAt = &reviewedAt.Time
	}
	if decidedAt.Valid {
		detail.DecidedAt = &decidedAt.Time
	}

	queryFiles := `SELECT file_type, file_name, file_url FROM submissions_file WHERE submissions_id = $1`
	var files []SubmissionFile
	if err = r.db.SelectContext(ctx, &files, queryFiles, submissionID); err != nil {
		slog.Warn("gagal mengambil file submission seeker", "submission_id", submissionID, "error", err)
	} else {
		for _, f := range files {
			switch f.FileType {
			case "cv":
				detail.CVDocName = f.FileName
				detail.CVDocURL = f.FileURL
			case "portfolio":
				pName := f.FileName
				pURL := f.FileURL
				detail.PortfolioDocName = &pName
				detail.PortfolioDocURL = &pURL
			}
		}
	}

	return &detail, nil
}

func (r *submissionsRepo) GetSubmissionsByJobID(ctx context.Context, jobID, compID, statusFilter string) ([]DTORecruiterSubmissionList, error) {
	var results []DTORecruiterSubmissionList
	query := `
		SELECT 
			s.id AS submission_id, 
			COALESCE(a.first_name, '') || ' ' || COALESCE(a.last_name, '') AS applicant_name, 
			acc.email AS applicant_email,
			COALESCE(acc.phone_number, '') AS applicant_phone,
			s.status, 
			s.applied_at, 
			s.internal_notes, 
			s.web_portfolio_url
		FROM submissions s
		JOIN applicants a ON s.applicant_id = a.id
		JOIN accounts acc ON a.account_id = acc.id
		JOIN jobs j ON s.job_id = j.id
		WHERE s.job_id = $1 AND j.company_id = $2
	`
	args := []interface{}{jobID, compID}

	if statusFilter != "" && statusFilter != "Semua" {
		query += ` AND s.status = $3`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY s.applied_at DESC`

	type TempSubmission struct {
		SubmissionID    string         `db:"submission_id"`
		ApplicantName   string         `db:"applicant_name"`
		ApplicantEmail  string         `db:"applicant_email"`
		ApplicantPhone  string         `db:"applicant_phone"`
		Status          string         `db:"status"`
		AppliedAt       time.Time      `db:"applied_at"`
		InternalNotes   sql.NullString `db:"internal_notes"`
		WebPortfolioURL sql.NullString `db:"web_portfolio_url"`
	}

	var tempSubs []TempSubmission
	if err := r.db.SelectContext(ctx, &tempSubs, query, args...); err != nil {
		slog.Error("gagal mengambil daftar pelamar job", "job_id", jobID, "company_id", compID, "error", err)
		return nil, err
	}

	for _, ts := range tempSubs {
		dto := DTORecruiterSubmissionList{
			SubmissionID:   ts.SubmissionID,
			ApplicantName:  ts.ApplicantName,
			ApplicantEmail: ts.ApplicantEmail,
			ApplicantPhone: ts.ApplicantPhone,
			Status:         ts.Status,
			AppliedAt:      ts.AppliedAt,
			Answers:        []DTOEssayAnswer{},
		}
		if ts.InternalNotes.Valid {
			dto.InternalNotes = &ts.InternalNotes.String
		}
		if ts.WebPortfolioURL.Valid {
			dto.PortfolioLinkWeb = &ts.WebPortfolioURL.String
		}

		var files []SubmissionFile
		if err := r.db.SelectContext(ctx, &files, `SELECT file_type, file_url FROM submissions_file WHERE submissions_id = $1`, ts.SubmissionID); err != nil {
			slog.Warn("gagal mengambil file submission recruiter", "submission_id", ts.SubmissionID, "error", err)
		} else {
			for _, f := range files {
				switch f.FileType {
				case "cv":
					dto.CVDocURL = f.FileURL
				case "portfolio":
					pURL := f.FileURL
					dto.PortfolioDocURL = &pURL
				}
			}
		}

		essayQuery := `
			SELECT q.question_detail AS question, e.answer 
			FROM essay_answer e
			JOIN essay_question q ON e.essay_question_id = q.id
			WHERE e.submissions_id = $1
		`
		if err := r.db.SelectContext(ctx, &dto.Answers, essayQuery, ts.SubmissionID); err != nil {
			slog.Warn("gagal mengambil jawaban essay submission", "submission_id", ts.SubmissionID, "error", err)
		}

		results = append(results, dto)
	}

	return results, nil
}

func (r *submissionsRepo) UpdateSubmissionStatus(ctx context.Context, submissionID, compID, status string) error {
	query := `
		UPDATE submissions 
		SET 
			status = $1, 
			reviewed_at = CASE 
				WHEN $1 IN ('reviewed', 'shortlisted', 'accepted', 'rejected') THEN COALESCE(submissions.reviewed_at, CURRENT_TIMESTAMP)
				ELSE submissions.reviewed_at
			END,
			decided_at = CASE 
				WHEN $1 IN ('accepted', 'rejected') THEN CURRENT_TIMESTAMP
				ELSE submissions.decided_at
			END,
			updated_at = CURRENT_TIMESTAMP
		FROM jobs
		WHERE submissions.job_id = jobs.id 
		  AND submissions.id = $2 
		  AND jobs.company_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, status, submissionID, compID)
	if err != nil {
		slog.Error("gagal update status submission", "submission_id", submissionID, "company_id", compID, "status", status, "error", err)
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("lamaran tidak ditemukan atau tidak memiliki akses")
	}
	return nil
}

func (r *submissionsRepo) UpdateInternalNotes(ctx context.Context, submissionID, compID, notes string) error {
	query := `
		UPDATE submissions 
		SET internal_notes = $1, updated_at = CURRENT_TIMESTAMP
		FROM jobs
		WHERE submissions.job_id = jobs.id 
		  AND submissions.id = $2 
		  AND jobs.company_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, notes, submissionID, compID)
	if err != nil {
		slog.Error("gagal update catatan internal submission", "submission_id", submissionID, "company_id", compID, "error", err)
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("lamaran tidak ditemukan atau tidak memiliki akses")
	}
	return nil
}

func (r *submissionsRepo) GetSubmissionsEssay(ctx context.Context, jobID string) ([]RawSubmissionEssayRow, error) {
	query := `
		SELECT 
			s.id AS submission_id,
			s.job_id,
			a.avatar_url,
			a.first_name || ' ' || a.last_name AS fullname,
			a.domicile,
			q.question_detail,
			ans.answer AS answer_detail
		FROM submissions s
		JOIN applicants a ON s.applicant_id = a.id
		JOIN essay_answer ans ON s.id = ans.submissions_id
		JOIN essay_question q ON ans.essay_question_id = q.id
		WHERE s.job_id = $1 AND status = 'accepted'`

	var rows []RawSubmissionEssayRow
	err := r.db.SelectContext(ctx, &rows, query, jobID)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
