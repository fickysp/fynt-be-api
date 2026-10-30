package documents

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID            uuid.UUID `db:"id" json:"id"`
	ApplicantID   uuid.UUID `db:"applicant_id" json:"applicant_id"`
	DocumentsType string    `db:"documents_type" json:"documents_type"`
	FileName      string    `db:"file_name" json:"file_name"`
	FileURL       string    `db:"file_url" json:"file_url"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type DTODocumentResponse struct {
	ID            string    `json:"id"`
	DocumentsType string    `json:"documents_type"`
	FileName      string    `json:"file_name"`
	FileURL       string    `json:"file_url"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type SwaggerDocumentsResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"Get data success"`
	Data    []DTODocumentResponse `json:"data"`
}

type DTOSetActiveDocumentRequest struct {
	DocumentsType string `json:"documents_type" validate:"required,oneof=cv portfolio"`
}
