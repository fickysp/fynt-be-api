package jobs

import (
	"context"
	"fmt"
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
	UpdateJobWithQuestionsTx(ctx context.Context, job *Job, questions []questions.Questions, accID string, newBalance *int64) error

	CountAllJobs(ctx context.Context, compID string) (*uint16, error)
	CountCloseJobs(ctx context.Context, compID string) (*uint16, error)
	CountActiveJobs(ctx context.Context, compID string) (*uint16, error)
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
		questionQuery := `INSERT INTO essay_question (id, job_id, question_detail) VALUES (:id, :job_id, :questions_detail)`
		if _, err = tx.NamedExecContext(ctx, questionQuery, questions); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *jobRepository) UpdateJobWithQuestionsTx(ctx context.Context, job *Job, questions []questions.Questions, accID string, newBalance *int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if newBalance != nil {
		_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2", *newBalance, accID)
		if err != nil {
			return err
		}
	}

	var setClauses []string
	var args []interface{}
	argCounter := 1

	appendClause := func(column string, value interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argCounter))
		args = append(args, value)
		argCounter++
	}

	if job.JobCategoryID != uuid.Nil {
		appendClause("job_category_id", job.JobCategoryID)
	}
	if job.Position != "" {
		appendClause("position", job.Position)
	}
	if job.Location != "" {
		appendClause("location", job.Location)
	}
	if job.Workplace != "" {
		appendClause("workplace", job.Workplace)
	}
	if job.Employment != "" {
		appendClause("employment", job.Employment)
	}
	if job.SalaryDesc != "" {
		appendClause("salary_desc", job.SalaryDesc)
	}
	if job.JobDesc != "" {
		appendClause("jobdesc", job.JobDesc)
	}
	if job.Status != "" {
		appendClause("status", job.Status)
	}
	if job.Slot != nil {
		appendClause("slot", job.Slot)
	}
	if job.CoinApplyment != 0 {
		appendClause("coin_applyment", job.CoinApplyment)
	}
	if job.ClosingDate != nil {
		appendClause("closing_date", job.ClosingDate)
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
		query := fmt.Sprintf("UPDATE jobs SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argCounter)
		args = append(args, job.ID)

		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	if questions != nil {
		if _, err = tx.ExecContext(ctx, "DELETE FROM essay_question WHERE job_id = $1", job.ID); err != nil {
			return err
		}

		if len(questions) > 0 {
			questionQuery := `INSERT INTO essay_question (id, job_id, question_detail) VALUES (:id, :job_id, :questions_detail)`
			if _, err = tx.NamedExecContext(ctx, questionQuery, questions); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *jobRepository) CountAllJobs(ctx context.Context, accID string) (*uint16, error) {
	var totalJobs uint16
	query := `SELECT COUNT(*) AS total FROM jobs WHERE account_id = $1`
	if err := r.db.SelectContext(ctx, &totalJobs, query, accID); err != nil {
		return nil, err
	}

	return &totalJobs, nil
}
func (r *jobRepository) CountActiveJobs(ctx context.Context, accID string) (*uint16, error) {
	var totalJobs uint16
	query := `SELECT COUNT(*) AS total FROM jobs WHERE account_id = $1 AND status = 'active'`
	if err := r.db.SelectContext(ctx, &totalJobs, query, accID); err != nil {
		return nil, err
	}

	return &totalJobs, nil
}
func (r *jobRepository) CountCloseJobs(ctx context.Context, accID string) (*uint16, error) {
	var totalJobs uint16
	query := `SELECT COUNT(*) AS total FROM jobs WHERE account_id = $1 AND status = 'close'`
	if err := r.db.SelectContext(ctx, &totalJobs, query, accID); err != nil {
		return nil, err
	}

	return &totalJobs, nil
}
