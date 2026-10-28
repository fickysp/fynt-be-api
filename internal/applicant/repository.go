package applicant

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ApplicantRepository interface {
	GetApplicantByAccountID(ctx context.Context, accID string) (*Applicant, error)
	Create(ctx context.Context, app *Applicant) error
	Update(ctx context.Context, app *Applicant) error
}

type applicantRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) ApplicantRepository {
	return &applicantRepository{db: db}
}

func (r *applicantRepository) GetApplicantByAccountID(ctx context.Context, accID string) (*Applicant, error) {
	var app Applicant
	query := `SELECT * FROM applicants WHERE account_id=$1`
	err := r.db.GetContext(ctx, &app, query, accID)
	if err != nil {
		return nil, err
	}
	return &app, err
}

func (r *applicantRepository) Create(ctx context.Context, app *Applicant) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO applicants (account_id, first_name, last_name)
	VALUES ($1, $2, $3)`

	_, err = tx.ExecContext(ctx, query, app.AccountID, app.FirstName, app.LastName)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *applicantRepository) Update(ctx context.Context, app *Applicant) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var setClauses []string
	var args []interface{}
	argCounter := 1

	appendClause := func(column string, value interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argCounter))
		args = append(args, value)
		argCounter++
	}

	if app.FirstName != nil {
		appendClause("first_name", app.FirstName)
	}
	if app.LastName != nil {
		appendClause("last_name", app.LastName)
	}
	if app.Domicile != nil {
		appendClause("domicile", app.Domicile)
	}
	if app.AvatarURL != nil {
		appendClause("portfolio_file_url", app.AvatarURL)
	}
	if app.ResumeURL != nil {
		appendClause("cv_url", app.ResumeURL)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf("UPDATE applicants SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argCounter)
	args = append(args, app.ID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
