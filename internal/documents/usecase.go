package documents

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"

	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/google/uuid"
)

type DocumentUsecase interface {
	UploadDocument(ctx context.Context, profID string, docType string, fileHeader *multipart.FileHeader) error
	GetDocuments(ctx context.Context, profID string) ([]DTODocumentResponse, error)
	DeleteDocument(ctx context.Context, profID, docID string) error
	SetActiveDocument(ctx context.Context, profID, docID, docType string) error
}

type documentUsecase struct {
	repo    DocumentRepository
	storage storage.SupabaseStorage
}

func NewDocumentUsecase(repo DocumentRepository, storage storage.SupabaseStorage) DocumentUsecase {
	return &documentUsecase{repo: repo, storage: storage}
}

func (u *documentUsecase) UploadDocument(ctx context.Context, profID, docType string, fileHeader *multipart.FileHeader) error {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid profile ID")
	}

	folderName := docType + "s"
	uploadedURL, err := u.storage.UploadFile(fileHeader, folderName)
	if err != nil {
		slog.Error("document upload failed",
			"file_name", fileHeader.Filename,
			"file_size", fileHeader.Size,
			"error", err,
		)
		return fmt.Errorf("gagal mengunggah dokumen: %w", err)
	}

	doc := &Document{
		ApplicantID:   appUUID,
		DocumentsType: docType,
		FileName:      fileHeader.Filename,
		FileURL:       uploadedURL,
		IsActive:      false,
	}

	slog.Info("document uploaded successfully",
		"file_name", fileHeader.Filename,
		"file_size", fileHeader.Size,
		"storage_folder", folderName,
	)

	return u.repo.Create(ctx, doc)
}

func (u *documentUsecase) GetDocuments(ctx context.Context, profID string) ([]DTODocumentResponse, error) {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID")
	}

	docs, err := u.repo.GetByApplicantID(ctx, appUUID)
	if err != nil {
		return nil, err
	}

	var responses []DTODocumentResponse
	for _, d := range docs {
		responses = append(responses, DTODocumentResponse{
			ID:            d.ID.String(),
			DocumentsType: d.DocumentsType,
			FileName:      d.FileName,
			FileURL:       d.FileURL,
			IsActive:      d.IsActive,
			CreatedAt:     d.CreatedAt,
		})
	}
	return responses, nil
}

func (u *documentUsecase) DeleteDocument(ctx context.Context, profID, docID string) error {
	appUUID, _ := uuid.Parse(profID)
	docUUID, err := uuid.Parse(docID)
	if err != nil {
		return fmt.Errorf("invalid document ID")
	}

	return u.repo.Delete(ctx, docUUID, appUUID)
}

func (u *documentUsecase) SetActiveDocument(ctx context.Context, profID, docID, docType string) error {
	appUUID, _ := uuid.Parse(profID)
	docUUID, err := uuid.Parse(docID)
	if err != nil {
		return fmt.Errorf("invalid document ID")
	}
	return u.repo.SetActive(ctx, docUUID, appUUID, docType)
}
