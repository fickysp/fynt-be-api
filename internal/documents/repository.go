package documents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByApplicantID(ctx context.Context, applicantID uuid.UUID) ([]Document, error)
	Delete(ctx context.Context, id, applicantID uuid.UUID) error
	SetActive(ctx context.Context, id, applicantID uuid.UUID, docType string) error
}

type documentRepository struct {
	db *sqlx.DB
}

func NewDocumentRepository(db *sqlx.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, doc *Document) error {
	query := `INSERT INTO documents (applicant_id, documents_type, file_name, file_url, is_active) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, doc.ApplicantID, doc.DocumentsType, doc.FileName, doc.FileURL, doc.IsActive)
	return err
}

func (r *documentRepository) GetByApplicantID(ctx context.Context, applicantID uuid.UUID) ([]Document, error) {
	var docs []Document
	query := `SELECT * FROM documents WHERE applicant_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &docs, query, applicantID)
	return docs, err
}

func (r *documentRepository) Delete(ctx context.Context, id, applicantID uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1 AND applicant_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, applicantID)
	return err
}

func (r *documentRepository) SetActive(ctx context.Context, id, applicantID uuid.UUID, docType string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryReset := `UPDATE documents SET is_active = false WHERE applicant_id = $1 AND documents_type = $2`
	if _, err := tx.ExecContext(ctx, queryReset, applicantID, docType); err != nil {
		return err
	}

	querySet := `UPDATE documents SET is_active = true WHERE id = $1 AND applicant_id = $2`
	if _, err := tx.ExecContext(ctx, querySet, id, applicantID); err != nil {
		return err
	}

	return tx.Commit()
}
