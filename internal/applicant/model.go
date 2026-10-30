package applicant

import (
	"time"

	"github.com/google/uuid"
)

type Applicant struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AccountID       uuid.UUID  `db:"account_id" json:"account_id"`
	FirstName       *string    `db:"first_name" json:"first_name"`
	LastName        *string    `db:"last_name" json:"last_name"`
	Gender          *string    `db:"gender" json:"gender,omitempty"`
	BirthDate       *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	About           *string    `db:"about" json:"about,omitempty"`
	Domicile        *string    `db:"domicile" json:"domicile,omitempty"`
	WebPortfolioURL *string    `db:"web_portfolio_url" json:"web_portfolio_url,omitempty"`
	AvatarURL       *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}
type DTOApplicantUpdateRequest struct {
	FirstName       *string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName        *string `json:"last_name" validate:"omitempty,min=2,max=50"`
	Gender          *string `json:"gender" validate:"omitempty,oneof=male female"`
	BirthDate       *string `json:"birth_date" validate:"omitempty,datetime=02/01/2006"`
	About           *string `json:"about" validate:"omitempty,max=500"`
	Domicile        *string `json:"domicile" validate:"omitempty,max=100"`
	WebPortfolioURL *string `json:"web_portfolio_url" validate:"omitempty,url"`
}
type DTOApplicantProfileResponse struct {
	ID              string `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Gender          string `json:"gender,omitempty"`
	BirthDate       string `json:"birth_date,omitempty"`
	About           string `form:"about,omitempty"`
	Domicile        string `json:"domicile,omitempty"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	WebPortfolioURL string `json:"web_portfolio_url,omitempty"`
}
type SwaggerGetApplicantResponse struct {
	Success bool                        `json:"success" example:"true"`
	Message string                      `json:"message" example:"Get data success"`
	Data    DTOApplicantProfileResponse `json:"data"`
}
