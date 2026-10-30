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
	UploadAvatar(ctx context.Context, appID, url string) error
}

type applicantRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) ApplicantRepository {
	return &applicantRepository{db: db}
}

func (r *applicantRepository) GetApplicantByAccountID(ctx context.Context, accID string) (*Applicant, error) {
	var app Applicant
	query := `SELECT id, account_id, first_name, last_name, gender, birth_date,about, domicile, web_portfolio_url, avatar_url, created_at
	FROM applicants WHERE account_id=$1`
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

	query := `INSERT INTO applicants (account_id, first_name, last_name, avatar_url)
	VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(ctx, query, app.AccountID, app.FirstName, app.LastName, app.AvatarURL)
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
	if app.Gender != nil {
		appendClause("gender", app.Gender)
	}
	if app.BirthDate != nil {
		appendClause("birth_date", app.BirthDate)
	}
	if app.About != nil {
		appendClause("about", app.About)
	}
	if app.WebPortfolioURL != nil {
		appendClause("web_portfolio_url", app.WebPortfolioURL)
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
func (r *applicantRepository) UploadAvatar(ctx context.Context, appID, url string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE applicants SET avatar_url = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, url, appID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
