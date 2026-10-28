package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID `db:"id" json:"id"`
	AccountID   uuid.UUID `db:"account_id" json:"account_id"`
	OrderID     string    `db:"order_id" json:"order_id"`
	Amount      int64     `db:"amount" json:"amount"`
	PaymentType *string   `db:"payment_type" json:"payment_type,omitempty"`
	Status      string    `db:"status" json:"status"`
	SnapToken   *string   `db:"snap_token" json:"snap_token,omitempty"`
	SnapURL     *string   `db:"snap_url" json:"snap_url,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type DTOMidtransNotification struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
	GrossAmount       string `json:"gross_amount"`
}
