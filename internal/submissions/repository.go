package submissions

import (
	"context"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type submissionsRepo struct {
	db *sqlx.DB
}

type SubmissionRepository interface {
	CreateSubmissionTx(ctx context.Context, req ApplyJobRequest, applicantID uuid.UUID, accountID string, cost int64) error
	GetAllSubmission(ctx context.Context, applId string) (*[]Submission, error)
}

func NewSubmissionRepository(db *sqlx.DB) SubmissionRepository {
	return &submissionsRepo{
		db: db,
	}
}

func (r *submissionsRepo) CreateSubmissionTx(ctx context.Context, req ApplyJobRequest, applicantID uuid.UUID, accountID string, cost int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if cost > 0 {
		_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2", cost, accountID)
		if err != nil {
			return err
		}
	}

	var submissionID uuid.UUID
	query := `INSERT INTO submissions (applicant_id, job_id, web_portfolio_url) VALUES ($1, $2, $3) RETURNING id`
	if err = tx.QueryRowContext(ctx, query, applicantID, req.JobID, req.WebPortfolioURL).Scan(&submissionID); err != nil {
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
			return err
		}
	}

	if len(req.Answers) > 0 {
		var answers []EssayAnswers
		for _, a := range req.Answers {
			qID, _ := uuid.Parse(a.QuestionID)
			answers = append(answers, EssayAnswers{
				SubmissionsID:   submissionID,
				EssayQuestionID: qID,
				Answer:          a.Answer,
			})
		}
		answerQuery := `INSERT INTO essay_answer (submissions_id, essay_question_id, answer) VALUES (:submissions_id, :essay_question_id, :answer)`
		if _, err = tx.NamedExecContext(ctx, answerQuery, answers); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *submissionsRepo) GetAllSubmission(ctx context.Context, applId string) (*[]Submission, error) {
	var result []Submission

	query := `SELECT 
    sj.id,
    sj.applicant_id,
    sj.resume_url,
    sj.portfolio_url,
    sj.web_portfolio_url,
    sj.status,
    sj.applied_at,
    sj.job_id,
    c.company_name,
    jv.location,
    jv.workplace,
    jv.employment,
    jv.position
    FROM submit_jobs sj
    JOIN jobs jv ON sj.job_id = jv.id
    JOIN companies c ON jv.company_id = c.id
    JOIN applicants a ON sj.applicant_id = a.id
    WHERE sj.applicant_id = $1
    ORDER BY sj.applied_at DESC;`

	if err := r.db.SelectContext(ctx, &result, query, applId); err != nil {
		return nil, err
	}

	return &result, nil
}
