package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository interface {
	Create(ctx context.Context, trx *Transaction) error
	Delete(ctx context.Context, orderID string) error
	ProcessWebhookTx(ctx context.Context, orderID, status, paymentType string) error
}

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, trx *Transaction) error {
	query := `INSERT INTO transactions (id, account_id, order_id, coin_amount, amount, status, snap_token, snap_url) 
	          VALUES (:id, :account_id, :order_id, :coin_amount, :amount, :status, :snap_token, :snap_url)`
	_, err := r.db.NamedExecContext(ctx, query, trx)
	return err
}

func (r *transactionRepository) Delete(ctx context.Context, orderID string) error {
	query := `DELETE FROM transactions WHERE order_id = $1`
	_, err := r.db.ExecContext(ctx, query, orderID)
	return err
}

func (r *transactionRepository) ProcessWebhookTx(ctx context.Context, orderID, status, paymentType string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var trx Transaction
	if err := tx.GetContext(ctx, &trx,
		`SELECT account_id, coin_amount, amount, status FROM transactions WHERE order_id = $1 FOR UPDATE`,
		orderID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("transaksi dengan order_id %s tidak ditemukan", orderID)
		}
		return fmt.Errorf("gagal ambil data transaksi: %w", err)
	}

	if trx.Status == "settlement" || trx.Status == "success" {
		slog.Info("transaksi sudah diproses sebelumnya, skip", "order_id", orderID)
		return nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE transactions SET status = $1, payment_type = $2, updated_at = CURRENT_TIMESTAMP WHERE order_id = $3`,
		status, paymentType, orderID,
	); err != nil {
		return fmt.Errorf("gagal update status transaksi: %w", err)
	}

	if status == "settlement" {
		result, err := tx.ExecContext(ctx,
			`UPDATE wallets SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE account_id = $2`,
			trx.CoinAmount, trx.AccountID,
		)
		if err != nil {
			return fmt.Errorf("gagal update saldo wallet: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("gagal cek rows affected wallet: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("wallet untuk account_id %s tidak ditemukan", trx.AccountID)
		}
	}

	return tx.Commit()
}
