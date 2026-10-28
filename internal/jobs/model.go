package jobs

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID            uuid.UUID  `db:"id"`
	CompanyID     uuid.UUID  `db:"company_id"`
	JobCategoryID uuid.UUID  `db:"job_category_id"`
	Position      string     `db:"position"`
	Location      string     `db:"location"`
	Workplace     string     `db:"workplace"`
	Employment    string     `db:"employment"`
	SalaryDesc    string     `db:"salary_desc"`
	JobDesc       string     `db:"jobdesc"`
	Status        string     `db:"status"`
	Slot          *int       `db:"slot"`
	FilledSlot    int        `db:"filled_slot"`
	CoinApplyment int64      `db:"coin_applyment"`
	ClosingDate   *time.Time `db:"closing_date"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type JobCategory struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

// DTO
type DTOJobCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type DTOJobRequest struct {
	CompanyID      string   `json:"-"`
	JobCategoryID  string   `json:"job_category_id" validate:"required,uuid"`
	Position       string   `json:"position" validate:"required"`
	Location       string   `json:"location" validate:"required"`
	Workplace      string   `json:"workplace" validate:"required"`
	Employment     string   `json:"employment" validate:"required"`
	SalaryDesc     string   `json:"salary_desc" validate:"required"`
	JobDesc        string   `json:"jobdesc" validate:"required"`
	Slot           *int     `json:"slot"`
	ExtendCoins    int64    `json:"extend_coins" validate:"omitempty"`
	EssayQuestions []string `json:"essay_questions" validate:"required,max=5,dive"`
}

type DTOUpdateJobRequest struct {
	CompanyID      string   `json:"-"`
	JobCategoryID  *string  `json:"job_category_id" validate:"omitempty,uuid"`
	Position       *string  `json:"position"`
	Location       *string  `json:"location"`
	Workplace      *string  `json:"workplace"`
	Employment     *string  `json:"employment"`
	SalaryDesc     *string  `json:"salary_desc"`
	JobDesc        *string  `json:"jobdesc"`
	Slot           *int     `json:"slot"`
	ExtendCoins    int64    `json:"extend_coins"`
	EssayQuestions []string `json:"essay_questions"`
}

type DTOJobCategoryResponses struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DTOJobsResponses struct {
	ID            uuid.UUID  `json:"id"`
	CompanyID     uuid.UUID  `json:"company_id"`
	JobCategoryID uuid.UUID  `json:"job_category_id"`
	Position      string     `json:"position"`
	Location      string     `json:"location"`
	Workplace     string     `json:"workplace"`
	Employment    string     `json:"employment"`
	SalaryDesc    string     `json:"salary_desc"`
	JobDesc       string     `json:"jobdesc"`
	Status        string     `json:"status"`
	Slot          *int       `json:"slot"`
	FilledSlot    int        `json:"filled_slot"`
	CoinApplyment int64      `json:"coin_applyment"`
	ClosingDate   *time.Time `json:"closing_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type DTOJobSummaryResponses struct {
	AllJobs    uint16 `json:"all_jobs"`
	ActiveJobs uint16 `json:"active_jobs"`
	CLoseJobs  uint16 `json:"close_jobs"`
}
