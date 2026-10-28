package applicant

import (
	"time"

	"github.com/google/uuid"
)

type Applicant struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	AccountID    uuid.UUID  `db:"account_id" json:"account_id"`
	FirstName    *string    `db:"first_name" json:"first_name"`
	LastName     *string    `db:"last_name" json:"last_name"`
	Gender       *string    `db:"gender" json:"gender,omitempty"`
	BirthDate    *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	Domicile     *string    `db:"domicile" json:"domicile,omitempty"`
	AvatarURL    *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	ResumeURL    *string    `db:"resume_url" json:"resume_url,omitempty"`
	PortfolioURL *string    `db:"portfolio_url" json:"portfolio_url,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
