package wallet

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID         uuid.UUID `db:"id" json:"id"`
	AccountID  uuid.UUID `db:"account_id" json:"account_id"`
	Balance    int64     `db:"balance" json:"balance"`
	Created_at time.Time `db:"created_at" json:"created_at"`
	Updated_at time.Time `db:"updated_at" json:"updated_at"`
}

type DTOTopUpRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount" validate:"required,min=10000"`
}

type DTOTopUpResponse struct {
	OrderID   string `json:"order_id"`
	Amount    int64  `json:"amount"`
	SnapToken string `json:"snap_token"`
	SnapURL   string `json:"snap_url"`
}

type DTOBalanceResponse struct {
	Balance int64 `json:"balance"`
}
