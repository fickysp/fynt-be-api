package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	Create(ctx context.Context, account *Account) (string, error)
	FindByEmail(ctx context.Context, email string) (*Account, error)
	GetById(ctx context.Context, id string) (*Account, error)
	RemoveRefreshToken(ctx context.Context, accID string) error
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) Create(ctx context.Context, account *Account) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}

	defer tx.Rollback()

	query := `INSERT INTO accounts (email, password, phone_number, role)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	var returnedID string
	err = tx.QueryRowxContext(ctx, query, account.Email, account.Password, account.PhoneNumber, account.Role).Scan(&returnedID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", nil
	}
	return returnedID, nil
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	var account Account
	query := `SELECT * FROM accounts WHERE email=$1`
	err := r.db.GetContext(ctx, &account, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &account, err
}

func (r *authRepository) GetById(ctx context.Context, id string) (*Account, error) {
	var account Account
	query := `SELECT * FROM accounts WHERE id=$1`
	err := r.db.GetContext(ctx, &account, query, id)
	return &account, err
}

func (r *authRepository) RemoveRefreshToken(ctx context.Context, accID string) error {
	query := `UPDATE accounts SET refresh_token = NULL, updated_at = CURENT TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, accID)
	return err
}
