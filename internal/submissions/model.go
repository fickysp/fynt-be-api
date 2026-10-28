package submissions

import (
	"time"

	"github.com/google/uuid"
)

type Submission struct {
	ID           uuid.UUID `db:"id"`
	ApplicantID  uuid.UUID `db:"applicant_id"`
	JobID        uuid.UUID `db:"job_id"`
	WebPortfolio string    `db:"web_portfolio_url"`
	Status       string    `db:"status"`
	AppliedAt    time.Time `db:"applied_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type SubmissionFile struct {
	ID            uuid.UUID `db:"id"`
	SubmissionsID uuid.UUID `db:"submissions_id"`
	FileType      string    `db:"file_type"`
	FileName      string    `db:"file_name"`
	FileSizeBytes uint64    `db:"file_size_bytes"`
	FileURL       string    `db:"file_url"`
}

type EssayAnswers struct {
	SubmissionsID   uuid.UUID `db:"submissions_id"`
	EssayQuestionID uuid.UUID `db:"essay_question_id"`
	Answer          string    `db:"answer"`
}

type DTOSubmissionFile struct {
	FileType      string `json:"file_type" validate:"required,oneof=cv portfolio"`
	FileName      string `json:"file_name" validate:"required"`
	FileSizeBytes uint64 `json:"file_size_bytes" validate:"required"`
	FileURL       string `json:"file_url" validate:"required,url"`
}

type EssayAnswerRequest struct {
	QuestionID string `json:"question_id" validate:"required,uuid"`
	Answer     string `json:"answer" validate:"required"`
}

type ApplyJobRequest struct {
	JobID           string               `json:"job_id" validate:"required,uuid"`
	WebPortfolioURL *string              `json:"web_portfolio_url" validate:"omitempty,url"`
	Files           []DTOSubmissionFile  `json:"files" validate:"required,min=1,dive"`
	Answers         []EssayAnswerRequest `json:"answers" validate:"dive"`
}

type DTOSubmission struct {
	ApplicantID  uuid.UUID `json:"applicant_id"`
	ResumeURL    string    `json:"resume_url"`
	PortfolioURL string    `json:"portfolio_url"`
	WebPortfolio string    `json:"web_portfolio_url"`
	Status       string    `json:"status"`
	AppliedAt    time.Time `json:"applied_at"`
	Job          DTOJob    `json:"job"`
}

type DTOJob struct {
	JobID       string `json:"job_id"`
	CompanyName string `json:"company_name"`
	Location    string `json:"location"`
	Workplace   string `json:"workplace"`
	Employment  string `json:"employment"`
	Position    string `json:"position"`
}
