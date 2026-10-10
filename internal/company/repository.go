package company

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type CompanyRepository interface {
	GetCompanyByAccountID(ctx context.Context, accID string) (*Company, error)
}

type companyRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) GetCompanyByAccountID(ctx context.Context, accID string) (*Company, error) {
	var comp Company
	query := `SELECT * FROM companies WHERE account_id=$1`
	err := r.db.GetContext(ctx, &comp, query, accID)
	if err != nil {
		return nil, err
	}
	return &comp, err
}
