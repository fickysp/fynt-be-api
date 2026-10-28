package wallet

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type walletRepository struct {
	db *sqlx.DB
}

type WalletRepository interface {
	Create(ctx context.Context, accID uuid.UUID, balance int64) error
	GetWalletByID(ctx context.Context, accID string) (*Wallet, error)
	Update(ctx context.Context, accID string, ammount int64) error
}

func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &walletRepository{
		db: db,
	}
}

func (r *walletRepository) Create(ctx context.Context, accID uuid.UUID, balance int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO wallets (account_id, balance) 
	VALUES ($1, $2)`

	_, err = tx.ExecContext(ctx, query, accID, balance)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *walletRepository) GetWalletByID(ctx context.Context, accID string) (*Wallet, error) {
	var wal Wallet
	query := `SELECT * FROM wallets WHERE account_id = $1`
	err := r.db.GetContext(ctx, &wal, query, accID)
	if err != nil {
		return nil, err
	}
	return &wal, nil
}

func (r *walletRepository) Update(ctx context.Context, accID string, ammount int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	defer tx.Rollback()

	query := `UPDATE wallets 
	SET balance=$1, updated_at = CURRENT_TIMESTAMP
	WHERE account_id = $2`
	_, err = tx.ExecContext(ctx, query, ammount, accID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
