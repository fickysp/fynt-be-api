package education

import (
	"time"

	"github.com/google/uuid"
)

type Education struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ApplicantID  uuid.UUID `db:"applicant_id" json:"applicant_id"`
	Institution  string    `db:"institution" json:"institution"`
	Degree       string    `db:"degree" json:"degree"`
	FieldOfStudy *string   `db:"field_of_study" json:"field_of_study,omitempty"`
	Periode      string    `db:"periode" json:"periode"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type DTOEducationRequest struct {
	Institution  string  `json:"institution"`
	Degree       string  `json:"degree"`
	FieldOfStudy *string `json:"field_of_study"`
	Periode      string  `json:"periode"`
}

type DTOUpdateEducationRequest struct {
	Institution  *string `json:"institution"`
	Degree       *string `json:"degree"`
	FieldOfStudy *string `json:"field_of_study"`
	Periode      *string `json:"periode"`
}

type SwaggerEducationsResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Get data success"`
	Data    []Education `json:"data"`
}
