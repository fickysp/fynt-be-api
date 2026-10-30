package submissions

import (
	"time"

	"github.com/google/uuid"
)

type Submission struct {
	ID           uuid.UUID  `db:"id"`
	ApplicantID  uuid.UUID  `db:"applicant_id"`
	JobID        uuid.UUID  `db:"job_id"`
	WebPortfolio string     `db:"web_portfolio_url"`
	Status       string     `db:"status"`
	AppliedAt    time.Time  `db:"applied_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	ReviewedAt   *time.Time `db:"reviewed_at"`
	DecidedAt    *time.Time `db:"decided_at"`
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
	JobID           string               `json:"job_id" validate:"omitempty,uuid"`
	WebPortfolioURL *string              `json:"web_portfolio_url" validate:"omitempty,url"`
	Files           []DTOSubmissionFile  `json:"files" validate:"required,min=1,dive"`
	Answers         []EssayAnswerRequest `json:"answers" validate:"dive"`
}
type DTOSeekerSubmissionResponse struct {
	ID          string    `db:"id" json:"id"`
	Position    string    `db:"position" json:"position"`
	CompanyName string    `db:"company_name" json:"company_name"`
	Location    string    `db:"location" json:"location"`
	WorkPlace   string    `db:"workplace" json:"workplace"`
	Status      string    `db:"status" json:"status"`
	AppliedAt   time.Time `db:"applied_at" json:"applied_at"`
}
type DTOSeekerSubmissionDetail struct {
	SubmissionID     string     `json:"submission_id"`
	Status           string     `json:"status"`
	Position         string     `json:"position"`
	CompanyName      string     `json:"company_name"`
	CompanyAvatar    *string    `json:"company_avatar,omitempty"`
	Location         string     `json:"location"`
	WorkPlace        string     `json:"workplace"`
	Employment       string     `json:"employment"`
	AppliedAt        time.Time  `json:"applied_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	DecidedAt        *time.Time `json:"decided_at,omitempty"`
	CVDocName        string     `json:"cv_doc_name"`
	CVDocURL         string     `json:"cv_doc_url"`
	PortfolioDocName *string    `json:"portfolio_doc_name,omitempty"`
	PortfolioDocURL  *string    `json:"portfolio_doc_url,omitempty"`
	PortfolioLinkWeb *string    `json:"portfolio_link_web,omitempty"`
}
type DTOEssayAnswer struct {
	Question string `db:"question" json:"question"`
	Answer   string `db:"answer" json:"answer"`
}

type DTORecruiterSubmissionList struct {
	SubmissionID   string    `json:"submission_id"`
	ApplicantName  string    `json:"applicant_name"`
	ApplicantEmail string    `json:"applicant_email"`
	ApplicantPhone string    `json:"applicant_phone,omitempty"`
	Status         string    `json:"status"`
	AppliedAt      time.Time `json:"applied_at"`

	CVDocURL         string  `json:"cv_doc_url,omitempty"`
	PortfolioDocURL  *string `json:"portfolio_doc_url,omitempty"`
	PortfolioLinkWeb *string `json:"portfolio_link_web,omitempty"`

	InternalNotes *string          `json:"internal_notes,omitempty"`
	Answers       []DTOEssayAnswer `json:"answers"`
}

type DTOUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=reviewed shortlisted accepted rejected"`
}

type DTOUpdateNotesRequest struct {
	InternalNotes string `json:"internal_notes" validate:"required"`
}

type SwaggerSeekerSubmissionsResponse struct {
	Success bool                          `json:"success" example:"true"`
	Message string                        `json:"message" example:"Get data success"`
	Data    []DTOSeekerSubmissionResponse `json:"data"`
}

type SwaggerSubmissionDetailResponse struct {
	Success bool                       `json:"success" example:"true"`
	Message string                     `json:"message" example:"Get data success"`
	Data    *DTOSeekerSubmissionDetail `json:"data"`
}

type SwaggerRecruiterSubmissionsResponse struct {
	Success bool                         `json:"success" example:"true"`
	Message string                       `json:"message" example:"Get data success"`
	Data    []DTORecruiterSubmissionList `json:"data"`
}
type DTOEssayQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
type DTOSubmissionEssay struct {
	SubmissionID string       `json:"submission_id"`
	JobID        string       `json:"job_id"`
	AvatarURL    string       `json:"avatar_url" validate:"omitempty"`
	Fullname     string       `json:"fullname"`
	Domicile     string       `json:"domicile" validate:"omitempty"`
	Essays       []DTOEssayQA `json:"essays"`
}
type RawSubmissionEssayRow struct {
	SubmissionID string  `db:"submission_id"`
	JobID        string  `db:"job_id"`
	AvatarURL    *string `db:"avatar_url"`
	Fullname     string  `db:"fullname"`
	Domicile     *string `db:"domicile"`
	Question     string  `db:"question_detail"`
	Answer       string  `db:"answer_detail"`
}

type SwaggerSubmissionEssay struct {
	Success bool                 `json:"success" example:"true"`
	Message string               `json:"message" example:"Get data success"`
	Data    []DTOSubmissionEssay `json:"data"`
}
