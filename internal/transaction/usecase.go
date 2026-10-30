package transaction

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

type TransactionUsecase interface {
	HandleNotification(ctx context.Context, midtransServerKey string, req DTOMidtransNotification) error
}

type transactionUsecase struct {
	trxRepo TransactionRepository
}

func NewTransactionUsecase(tr TransactionRepository) TransactionUsecase {
	return &transactionUsecase{
		trxRepo: tr,
	}
}

func (u *transactionUsecase) HandleNotification(ctx context.Context, midtransServerKey string, req DTOMidtransNotification) error {
	raw := req.OrderID + req.StatusCode + req.GrossAmount + midtransServerKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])
	slog.Info("debug signature",
		"order_id", req.OrderID,
		"status_code", req.StatusCode,
		"gross_amount", req.GrossAmount,
		"server_key_len", len(midtransServerKey),
		"expected", expected,
		"received", req.SignatureKey,
		"match", expected == req.SignatureKey,
	)

	if expected != req.SignatureKey {
		slog.Warn("signature midtrans tidak valid", "order_id", req.OrderID)
		return fiber.NewError(fiber.StatusUnauthorized, "Signature tidak valid")
	}

	var internalStatus string
	switch req.TransactionStatus {
	case "capture":
		if req.FraudStatus == "accept" {
			internalStatus = "settlement"
		} else if req.FraudStatus == "challenge" {
			internalStatus = "pending"
		} else {
			internalStatus = "cancel"
		}
	case "settlement":
		internalStatus = "settlement"
	case "pending":
		internalStatus = "pending"
	case "deny", "cancel", "expire":
		internalStatus = "cancel"
	default:
		slog.Warn("transaction status tidak dikenali", "status", req.TransactionStatus, "order_id", req.OrderID)
		return fmt.Errorf("unknown transaction status: %s", req.TransactionStatus)
	}

	if err := u.trxRepo.ProcessWebhookTx(ctx, req.OrderID, internalStatus, req.PaymentType); err != nil {
		slog.Error("gagal proses webhook transaksi", "error", err, "order_id", req.OrderID)
		return fmt.Errorf("gagal memproses transaksi: %w", err)
	}

	return nil
}
