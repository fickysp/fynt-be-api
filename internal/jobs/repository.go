package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fickysp/fynt-be-api/internal/questions"
	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type JobRepository interface {
	GetJobById(ctx context.Context, jobID string) (*Job, error)
	GetAllCategories(ctx context.Context) ([]JobCategory, error)
	GetAllJob(ctx context.Context) ([]Job, error)
	CreateJobWithQuestionsTx(ctx context.Context, job *Job, questions []questions.Questions, accID string, newBalance int64) error
	UpdateJobDetails(ctx context.Context, jobID uuid.UUID, accID, jobDesc, status string, extendCoins int, newWalletBalance *int64) error
	SearchJobs(ctx context.Context, keyword string) ([]DTOJobSearchResponses, error)
	GetJobStatistics(ctx context.Context, compID string) (*DTOJobSummaryResponses, error)
}

type JobStats struct {
	TotalJobs  uint16 `db:"total_jobs"`
	ActiveJobs uint16 `db:"active_jobs"`
	CloseJobs  uint16 `db:"close_jobs"`
}
type jobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) GetJobById(ctx context.Context, jobID string) (*Job, error) {

	tx, err := r.db.BeginTxx(ctx, nil)

	defer tx.Rollback()

	var job Job
	query := `SELECT * FROM jobs WHERE id = $1`

	err = r.db.GetContext(ctx, &job, query, jobID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *jobRepository) GetAllCategories(ctx context.Context) ([]JobCategory, error) {
	var categories []JobCategory
	query := `SELECT * FROM job_category ORDER BY name ASC`

	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
func (r *jobRepository) GetAllJob(ctx context.Context) ([]Job, error) {
	var jobs []Job
	query := `SELECT * FROM jobs ORDER BY position ASC`

	err := r.db.SelectContext(ctx, &jobs, query)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *jobRepository) CreateJobWithQuestionsTx(ctx context.Context, job *Job, questions []questions.Questions, accID string, newBalance int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2", newBalance, accID)
	if err != nil {
		return err
	}

	jobQuery := `INSERT INTO jobs (
		id, company_id, job_category_id, position, location, workplace, employment, 
		salary_desc, jobdesc, status, slot, coin_applyment, closing_date, created_at, updated_at
	) VALUES (
		:id, :company_id, :job_category_id, :position, :location, :workplace, :employment, 
		:salary_desc, :jobdesc, :status, :slot, :coin_applyment, :closing_date, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
	)`

	if _, err = tx.NamedExecContext(ctx, jobQuery, job); err != nil {
		return err
	}

	if len(questions) > 0 {
		questionQuery := `INSERT INTO essay_question (id, job_id, question_detail) VALUES (:id, :job_id, :question_detail)`
		if _, err = tx.NamedExecContext(ctx, questionQuery, questions); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *jobRepository) UpdateJobDetails(ctx context.Context, jobID uuid.UUID, accID, jobDesc, status string, extendCoins int, newWalletBalance *int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if extendCoins > 0 && newWalletBalance != nil {
		_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2", *newWalletBalance, accID)
		if err != nil {
			return fmt.Errorf("gagal update wallet balance: %w", err)
		}
	}

	var setClauses []string
	var args []interface{}
	argCounter := 1

	if jobDesc != "" {
		setClauses = append(setClauses, fmt.Sprintf("jobdesc = $%d", argCounter))
		args = append(args, jobDesc)
		argCounter++
	}
	if status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, status)
		argCounter++
	}
	if extendCoins > 0 {
		setClauses = append(setClauses, fmt.Sprintf("coin_applyment = coin_applyment + $%d", argCounter))
		args = append(args, extendCoins)
		argCounter++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
		query := fmt.Sprintf("UPDATE jobs SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argCounter)
		args = append(args, jobID)

		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("gagal update data jobs: %w", err)
		}
	} else {
		return fmt.Errorf("tidak ada data job yang diperbarui")
	}

	if status == "close" {
		rejectQuery := `
			UPDATE submissions 
			SET status = 'rejected', decided_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			WHERE job_id = $1 
			  AND status NOT IN ('reviewed', 'shortlisted', 'accepted', 'rejected')
		`
		if _, err = tx.ExecContext(ctx, rejectQuery, jobID); err != nil {
			return fmt.Errorf("gagal memperbarui status pelamar menjadi rejected: %w", err)
		}
	}

	return tx.Commit()
}

func (r *jobRepository) SearchJobs(ctx context.Context, keyword string) ([]DTOJobSearchResponses, error) {
	var jobs []DTOJobSearchResponses
	query := `
		SELECT 
			j.id,
			c.company_name, 
			j.position,
			j.location,
			j.workplace,
			j.employment,
			j.salary_desc,
			j.status,
			j.slot,
			j.filled_slot,
			j.coin_applyment,
			j.closing_date
		FROM jobs j
		JOIN companies c ON j.company_id = c.id
		WHERE j.position ILIKE $1 OR c.company_name ILIKE $1
		ORDER BY j.created_at DESC
	`

	if err := r.db.SelectContext(ctx, &jobs, query, keyword); err != nil {
		slog.Error("Penyebab error", "keyword", keyword, "error", err)
		return nil, err
	}

	return jobs, nil
}

func (r *jobRepository) GetJobStatistics(ctx context.Context, compID string) (*DTOJobSummaryResponses, error) {
	var stats DTOJobSummaryResponses
	query := `
		SELECT 
			COUNT(DISTINCT j.id) AS total_jobs,
			COUNT(DISTINCT CASE WHEN j.status = 'active' THEN j.id END) AS active_jobs,
			COUNT(DISTINCT CASE WHEN j.status = 'close' THEN j.id END) AS close_jobs,
			
			COUNT(s.id) AS total_submissions,
			COUNT(CASE WHEN j.status = 'active' AND s.id IS NOT NULL THEN s.id END) AS submissions_in_active_job,
			COUNT(CASE WHEN s.status = 'reviewed' THEN s.id END) AS submissions_in_review
		FROM jobs j
		LEFT JOIN submissions s ON j.id = s.job_id
		WHERE j.company_id = $1`

	if err := r.db.GetContext(ctx, &stats, query, compID); err != nil {
		return nil, err
	}

	return &stats, nil
}
