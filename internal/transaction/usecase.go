package transaction

import (
	"context"
	"errors"
)

type TransactionUsecase interface {
	HandleNotification(ctx context.Context, req DTOMidtransNotification) error
}

type transactionUsecase struct {
	trxRepo TransactionRepository
}

func NewTransactionUsecase(tr TransactionRepository) TransactionUsecase {
	return &transactionUsecase{
		trxRepo: tr,
	}
}

func (u *transactionUsecase) HandleNotification(ctx context.Context, req DTOMidtransNotification) error {
	var internalStatus string
	switch req.TransactionStatus {
	case "capture", "settlement":
		internalStatus = "settlement"
	case "pending":
		internalStatus = "pending"
	case "deny", "cancel", "expire":
		internalStatus = "cancel"
	default:
		return errors.New("unknown transaction status")
	}

	return u.trxRepo.ProcessWebhookTx(ctx, req.OrderID, internalStatus, req.PaymentType)
}
