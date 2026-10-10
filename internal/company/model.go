package company

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID            uuid.UUID `db:"id" json:"id"`
	AccountID     uuid.UUID `db:"account_id" json:"account_id"`
	CompanyName   string    `db:"company_name" json:"company_name"`
	Sector        []Sector  `db:"-" json:"sector,omitempty"`
	Address       *string   `db:"address" json:"address,omitempty"`
	Description   *string   `db:"description" json:"description,omitempty"`
	ContactPhone  *string   `db:"contact_phone" json:"contact_phone,omitempty"`
	AvatarURL     *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	AverageRating float32   `db:"average_rating" json:"average_rating"`
	TotalReviews  int       `db:"total_reviews" json:"total_reviews"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type Sector struct {
	ID         uuid.UUID `db:"id" json:"id"`
	SectorName string    `db:"sectore_name" json:"sector_name"`
}
