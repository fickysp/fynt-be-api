package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	Create(ctx context.Context, account *Account) (string, error)
	SaveRefresh(ctx context.Context, token, accID string) error
	FindByEmail(ctx context.Context, email string) (*Account, error)
	GetById(ctx context.Context, id string) (*Account, error)
	RemoveRefreshToken(ctx context.Context, accID string) error
	CreateOauthAccount(ctx context.Context, account *Account) (string, error)
	SaveExchangeCode(ctx context.Context, code string, token string, expiresAt time.Time) error
	GetAndConsumeExchangeCode(ctx context.Context, code string) (string, error)
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
func (r *authRepository) SaveRefresh(ctx context.Context, token, accID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `UPDATE accounts SET refresh_token = $1 WHERE id = $2`

	_, err = tx.ExecContext(ctx, query, token, accID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return nil
	}
	return nil
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
	return &account, nil
}

func (r *authRepository) GetById(ctx context.Context, id string) (*Account, error) {
	var account Account
	query := `SELECT * FROM accounts WHERE id=$1`
	err := r.db.GetContext(ctx, &account, query, id)
	return &account, err
}

func (r *authRepository) RemoveRefreshToken(ctx context.Context, accID string) error {
	query := `UPDATE accounts SET refresh_token = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, accID)
	return err
}

func (r *authRepository) CreateOauthAccount(ctx context.Context, account *Account) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}

	defer tx.Rollback()

	query := `INSERT INTO accounts (email, auth_provider, role)
	VALUES ($1, $2, $3)
	RETURNING id`

	var returnedID string
	err = tx.QueryRowxContext(ctx, query, account.Email, account.AuthProvider, account.Role).Scan(&returnedID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", nil
	}
	return returnedID, nil
}

func (r *authRepository) SaveExchangeCode(ctx context.Context, code string, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO oauth_exchange_codes (code, token, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, code, token, expiresAt)
	return err
}

func (r *authRepository) GetAndConsumeExchangeCode(ctx context.Context, code string) (string, error) {
	query := `
        DELETE FROM oauth_exchange_codes
        WHERE code = $1 AND expires_at > NOW()
        RETURNING token
    `
	var token string
	err := r.db.GetContext(ctx, &token, query, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("Exchange code tidak ditemukan")
		}
		return "", err
	}

	return token, nil
}
