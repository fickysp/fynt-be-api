package questions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type QuestionsRepository interface {
	Create(ctx context.Context, qu *Questions) error
	GetByJobID(ctx context.Context, jobID uuid.UUID) ([]Questions, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, jobID uuid.UUID, questions []DTOUpdateQuestion) error
}
type questionsRepository struct {
	db *sqlx.DB
}

func NewQuestionRepository(db *sqlx.DB) QuestionsRepository {
	return &questionsRepository{
		db: db,
	}
}

func (r *questionsRepository) Create(ctx context.Context, qu *Questions) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()
	query := `INSERT INTO essay_question (job_id, question_detail)
	VALUES ($1, $2)`

	_, err = tx.ExecContext(ctx, query, qu.JobID, qu.QuestionsDetail)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
func (r *questionsRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) ([]Questions, error) {
	var qs []Questions
	query := `SELECT id, job_id, question_detail FROM essay_question WHERE job_id = $1`

	if err := r.db.SelectContext(ctx, &qs, query, jobID); err != nil {
		return nil, err
	}
	return qs, nil
}

func (r *questionsRepository) Update(ctx context.Context, jobID uuid.UUID, questions []DTOUpdateQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE essay_question SET question_detail = $1 WHERE id = $2 AND job_id = $3`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("gagal prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, q := range questions {
		if strings.TrimSpace(*q.QuestionsDetail) == "" {
			continue
		}

		result, err := stmt.ExecContext(ctx, q.QuestionsDetail, q.ID, jobID)
		if err != nil {
			return fmt.Errorf("gagal update question %s: %w", q.ID.String(), err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("gagal cek rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("question %s tidak ditemukan atau bukan milik job ini", q.ID.String())
		}
	}

	return tx.Commit()
}

func (r *questionsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM questions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
