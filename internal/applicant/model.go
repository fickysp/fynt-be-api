package applicant

import (
	"time"

	"github.com/google/uuid"
)

type Education string

const (
	EduSd  Education = "sd"
	EduSmp Education = "smp"
	EduSma Education = "sma"
	EduSmk Education = "smk"
	EduD3  Education = "d3"
	EduD4  Education = "d4"
	EduS1  Education = "s1"
	EduS2  Education = "s2"
)

type Applicant struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	AccountID   string     `db:"account_id" json:"account_id"`
	FirstName   string     `db:"first_name" json:"first_name"`
	LastName    string     `db:"last_name" json:"last_name"`
	Username    *string    `json:"username,omitempty"`
	Gender      *string    `db:"gender" json:"gender,omitempty"`
	BirthDate   *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	Education   *Education `db:"education" json:"education,omitempty"`
	Institution *string    `db:"institution" json:"institution,omitempty"`
	AvatarURL   *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	ResumeURL   *string    `db:"resume_url" json:"resume_url,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
