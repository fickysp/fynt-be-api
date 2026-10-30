package experiences

import (
	"time"

	"github.com/google/uuid"
)

type Experience struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ApplicantID uuid.UUID `db:"applicant_id" json:"applicant_id"`
	Position    string    `db:"position" json:"position"`
	CompanyName string    `db:"company_name" json:"company_name"`
	Periode     string    `db:"periode" json:"periode"`
	Jobdesc     *string   `db:"jobdesc" json:"jobdesc,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type DTOExperienceRequest struct {
	Position    string  `json:"position"`
	CompanyName string  `json:"company_name"`
	Periode     string  `json:"periode"`
	Jobdesc     *string `json:"jobdesc"`
}

type DTOUpdateExperienceRequest struct {
	Position    *string `json:"position"`
	CompanyName *string `json:"company_name"`
	Periode     *string `json:"periode"`
	Jobdesc     *string `json:"jobdesc"`
}

type SwaggerExperiencesResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"Get data success"`
	Data    []Experience `json:"data"`
}
