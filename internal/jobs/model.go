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
	ImageURL      *string    `db:"image_url"`
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

type DTOJobUpdate struct {
	CompanyID   *string `json:"-" validate:"-"`
	JobDesc     *string `json:"jobdesc" validate:"omitempty"`
	Status      *string `json:"status" validate:"omitempty,oneof=active close draft"`
	ExtendCoins *int    `json:"extend_coins" validate:"omitempty,min=0"`
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

type DTOJobSearchResponses struct {
	ID            string `json:"id"`
	CompanyName   string `json:"company_name"`
	Position      string `json:"position"`
	Location      string `json:"location"`
	Workplace     string `json:"workplace"`
	Employment    string `json:"employment"`
	SalaryDesc    string `json:"salary_desc"`
	Status        string `json:"status"`
	Slot          int    `json:"slot"`
	FilledSlot    int    `json:"filled_slot"`
	CoinApplyment int64  `json:"coin_applyment"`
	ClosingDate   string `json:"closing_date"`
}

type DTOJobSummaryResponses struct {
	TotalJobs              uint16 `db:"total_jobs" json:"total_jobs"`
	ActiveJobs             uint16 `db:"active_jobs" json:"active_jobs"`
	CloseJobs              uint16 `db:"close_jobs" json:"close_jobs"`
	TotalSubmissions       uint16 `db:"total_submissions" json:"total_submissions"`
	SubmissionsInActiveJob uint64 `db:"submissions_in_active_job" json:"submissions_in_active_job"`
	SubmissionsInReview    uint16 `db:"submissions_in_review" json:"submissions_in_review"`
}

type SwaggerJobCategoriesResponse struct {
	Success bool                      `json:"success" example:"true"`
	Message string                    `json:"message" example:"Get data success"`
	Data    []DTOJobCategoryResponses `json:"data"`
}

type SwaggerJobsResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"Get data success"`
	Data    []DTOJobsResponses `json:"data"`
}

type SwaggerJobResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"Get data success"`
	Data    *DTOJobsResponses `json:"data"`
}

type SwaggerJobSearchResponse struct {
	Success bool                    `json:"success" example:"true"`
	Message string                  `json:"message" example:"Get data success"`
	Data    []DTOJobSearchResponses `json:"data"`
}

type SwaggerJobSummaryResponse struct {
	Success bool                    `json:"success" example:"true"`
	Message string                  `json:"message" example:"Get data success"`
	Data    *DTOJobSummaryResponses `json:"data"`
}
