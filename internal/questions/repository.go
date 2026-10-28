package questions

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type QuestionsRepository interface {
	Create(ctx context.Context, qu *Questions) error
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
	query := `INSERT INTO essay_question (job_id, questions_detail)
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
