package company

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type CompanyRepository interface {
	Create(ctx context.Context, com *Company) error
	GetCompanyByAccountID(ctx context.Context, accID string) (*Company, error)
}

type companyRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, com *Company) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO companies (account_id, company_name, sector_id, domicile, web_url, address, owner_name)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.ExecContext(ctx,
		query,
		com.AccountID,
		com.CompanyName,
		com.SectorID,
		com.Domicile,
		com.WebURL,
		com.Address,
		com.OwnerName)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
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
