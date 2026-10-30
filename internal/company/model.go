package company

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID            uuid.UUID `db:"id" json:"id"`
	AccountID     uuid.UUID `db:"account_id" json:"account_id"`
	SectorID      uuid.UUID `db:"sector_id" json:"sector_id"`
	CompanyName   string    `db:"company_name" json:"company_name"`
	Domicile      string    `db:"domicile" json:"domicile"`
	WebURL        *string   `db:"web_url" json:"web_url"`
	Address       string    `db:"address" json:"address,omitempty"`
	OwnerName     string    `db:"owner_name" json:"owner_name"`
	Description   *string   `db:"description" json:"description,omitempty"`
	AvatarURL     *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	AverageRating float32   `db:"average_rating" json:"average_rating"`
	TotalReviews  int       `db:"total_reviews" json:"total_reviews"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type DTOCompanyRegisterRequest struct {
	CompanyName    string `json:"company_name" validate:"required"`
	Sector         string `json:"sector" validate:"required"`
	Domicile       string `json:"domicile" validate:"required"`
	WebURL         string `json:"web_url"`
	Address        string `json:"address" validate:"required"`
	OwnerName      string `json:"owner_name" validate:"required"`
	PhoneNumber    string `json:"phone_number" validate:"required"`
	Password       string `json:"password" validate:"required,min=8" `
	VerifyPassword string `json:"verify_password" validate:"required,eqfield=Password"`
	CallbackUrl    string `json:"callback_url"`
}
type DTOCompanyUpdateRequest struct {
	CompanyName string  `json:"company_name"`
	Sector      string  `json:"sector_id" validate:"omitempty,uuid"`
	Domicile    string  `json:"domicile"`
	WebURL      string  `json:"web_url" validate:"omitempty,url"`
	Address     string  `json:"address"`
	OwnerName   string  `json:"owner_name"`
	Description *string `json:"description,omitempty"`
}
