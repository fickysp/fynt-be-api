package applicant

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ApplicantRepository interface {
	GetApplicantByAccountID(ctx context.Context, accID string) (*Applicant, error)
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
