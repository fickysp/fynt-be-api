package company

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

type CompanyRepository interface {
	Create(ctx context.Context, com *Company) error
	GetCompanyByAccountID(ctx context.Context, accID string) (*Company, error)
	InsertPassAndPhone(ctx context.Context, pass, phone string) error
	Update(ctx context.Context, comp *Company) error
	UploadAvatar(ctx context.Context, compID, url string) error
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
func (r *companyRepository) InsertPassAndPhone(ctx context.Context, pass, phone string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `INSERT into accounts (password, phone_number)
	VALUES ($1, $2)`

	_, err = tx.ExecContext(ctx, query, pass, phone)
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil

}

func (r *companyRepository) Update(ctx context.Context, comp *Company) error {
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

	if comp.CompanyName != "" {
		appendClause("company_name", comp.CompanyName)
	}
	if comp.Domicile != "" {
		appendClause("domicile", comp.Domicile)
	}
	if comp.Address != "" {
		appendClause("address", comp.Address)
	}
	if comp.OwnerName != "" {
		appendClause("owner_name", comp.OwnerName)
	}

	if comp.SectorID != uuid.Nil {
		appendClause("sector_id", comp.SectorID)
	}

	if comp.WebURL != nil {
		appendClause("web_url", comp.WebURL)
	}
	if comp.Description != nil {
		appendClause("description", comp.Description)
	}
	if comp.AvatarURL != nil {
		appendClause("avatar_url", comp.AvatarURL)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf("UPDATE companies SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argCounter)
	args = append(args, comp.ID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *companyRepository) UploadAvatar(ctx context.Context, compID, url string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE companies SET avatar_url = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, url, compID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
