package transaction

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository interface {
	Create(ctx context.Context, trx *Transaction) error
	UpdateSnap(ctx context.Context, orderID, sToken, sURL string) error
	ProcessWebhookTx(ctx context.Context, orderID, status, paymentType string) error
}

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, trx *Transaction) error {
	query := `INSERT INTO transactions (id, account_id, order_id, amount, status) 
              VALUES (:id, :account_id, :order_id, :amount, :status)`
	_, err := r.db.NamedExecContext(ctx, query, trx)
	return err
}

func (r *transactionRepository) UpdateSnap(ctx context.Context, orderID, sToken, sURL string) error {
	query := `UPDATE transactions SET snap_token = $1, snap_url = $2, updated_at = CURRENT_TIMESTAMP WHERE order_id = $3`
	_, err := r.db.ExecContext(ctx, query, sToken, sURL, orderID)
	return err
}

func (r *transactionRepository) ProcessWebhookTx(ctx context.Context, orderID, status, paymentType string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var trx Transaction
	if err := tx.GetContext(ctx, &trx, "SELECT account_id, amount, status FROM transactions WHERE order_id = $1 FOR UPDATE", orderID); err != nil {
		return err
	}

	if trx.Status == "settlement" || trx.Status == "success" {
		return nil
	}

	if _, err := tx.ExecContext(ctx, "UPDATE transactions SET status = $1, payment_type = $2, updated_at = CURRENT_TIMESTAMP WHERE order_id = $3", status, paymentType, orderID); err != nil {
		return err
	}

	if status == "settlement" || status == "success" {
		walletQuery := "UPDATE wallets SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2"
		if _, err := tx.ExecContext(ctx, walletQuery, trx.Amount, trx.AccountID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
