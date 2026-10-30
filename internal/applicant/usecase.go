package applicant

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/google/uuid"
)

type applicantUsecase struct {
	applicantRepository ApplicantRepository
	storage             storage.SupabaseStorage
}

type ApplicantUsecase interface {
	UpdateProfile(ctx context.Context, profID string, req DTOApplicantUpdateRequest) error
	UpdateAvatar(ctx context.Context, profID string, avFile *multipart.FileHeader) (string, error)
	GetApplicant(ctx context.Context, accID string) (*Applicant, error)
}

func NewApplicantUsecase(aR ApplicantRepository, s storage.SupabaseStorage) ApplicantUsecase {
	return &applicantUsecase{
		applicantRepository: aR,
		storage:             s,
	}
}

// Usecase khusus memperbarui data teks profil
func (u *applicantUsecase) UpdateProfile(ctx context.Context, profID string, req DTOApplicantUpdateRequest) error {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid profile ID format: %w", err)
	}

	app := &Applicant{
		ID:              appUUID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Gender:          req.Gender,
		Domicile:        req.Domicile,
		About:           req.About,
		WebPortfolioURL: req.WebPortfolioURL,
	}

	if req.BirthDate != nil && *req.BirthDate != "" {
		bdTime, err := time.Parse("02/01/2006", *req.BirthDate)
		if err != nil {
			return fmt.Errorf("format tanggal lahir tidak valid (DD/MM/YYYY): %w", err)
		}
		app.BirthDate = &bdTime
	}

	if err := u.applicantRepository.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update applicant profile: %w", err)
	}

	return nil
}

// Usecase khusus memperbarui avatar
func (u *applicantUsecase) UpdateAvatar(ctx context.Context, profID string, avFile *multipart.FileHeader) (string, error) {
	// Validasi ukuran (maks 2MB) & tipe file jika belum dilakukan di handler
	const maxFileSize = 2 * 1024 * 1024
	if avFile.Size > maxFileSize {
		return "", fmt.Errorf("ukuran file terlalu besar, maksimal 2MB")
	}

	// Upload ke Supabase
	url, err := u.storage.UploadFile(avFile, "profiles")
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah avatar ke storage: %w", err)
	}

	// Simpan URL ke database menggunakan method repository yang sudah ada
	if err := u.applicantRepository.UploadAvatar(ctx, profID, url); err != nil {
		return "", fmt.Errorf("gagal memperbarui avatar di database: %w", err)
	}

	return url, nil
}

func (u *applicantUsecase) GetApplicant(ctx context.Context, accID string) (*Applicant, error) {
	return u.applicantRepository.GetApplicantByAccountID(ctx, accID)
}
