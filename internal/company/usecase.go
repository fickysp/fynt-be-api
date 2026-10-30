package company

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/fickysp/fynt-be-api/internal/sectors"
	"github.com/fickysp/fynt-be-api/pkg/bcrypt"
	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/google/uuid"
)

type WalletRepository interface {
	Create(ctx context.Context, accID uuid.UUID, balance int64) error
}
type CompanyUsecase interface {
	RegisterCompany(ctx context.Context, accID string, req DTOCompanyRegisterRequest) error
	UpdateProfile(ctx context.Context, profID string, req DTOCompanyUpdateRequest) error
	UpdateAvatar(ctx context.Context, profID string, avFile *multipart.FileHeader) (string, error)
}
type companyUsecase struct {
	companyRepo CompanyRepository
	walletRepo  WalletRepository
	sectorRepo  sectors.SectorRepository
	storage     storage.SupabaseStorage
}

func NewCompanyUsecase(cr CompanyRepository, wr WalletRepository, sr sectors.SectorRepository, str storage.SupabaseStorage) CompanyUsecase {
	return &companyUsecase{
		companyRepo: cr,
		walletRepo:  wr,
		sectorRepo:  sr,
		storage:     str,
	}
}
func (u *companyUsecase) RegisterCompany(ctx context.Context, accID string, req DTOCompanyRegisterRequest) error {
	hashPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	accUUID, err := uuid.Parse(accID)
	if err != nil {
		return err
	}

	sectorData, err := u.sectorRepo.GetSectorByName(ctx, req.Sector)

	companyData := &Company{
		AccountID:   accUUID,
		SectorID:    sectorData.ID,
		CompanyName: req.CompanyName,
		Domicile:    req.Domicile,
		WebURL:      &req.WebURL,
		Address:     req.Address,
		OwnerName:   req.OwnerName,
	}

	if err := u.companyRepo.InsertPassAndPhone(ctx, hashPassword, req.PhoneNumber); err != nil {
		return fmt.Errorf("failed to insert password and phone number: %w", err)
	}

	if err := u.companyRepo.Create(ctx, companyData); err != nil {
		return fmt.Errorf("failed to create company profile: %w", err)
	}

	if err := u.walletRepo.Create(ctx, accUUID, 5); err != nil {
		return fmt.Errorf("failed to create company wallet: %w", err)
	}

	return nil
}

func (u *companyUsecase) UpdateProfile(ctx context.Context, profID string, req DTOCompanyUpdateRequest) error {
	compUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid profile ID format: %w", err)
	}

	company := &Company{
		ID:          compUUID,
		CompanyName: req.CompanyName,
		SectorID:    uuid.MustParse(req.Sector),
		Domicile:    req.Domicile,
		WebURL:      &req.WebURL,
		Address:     req.Address,
		OwnerName:   req.OwnerName,
		Description: req.Description,
	}

	if err := u.companyRepo.Update(ctx, company); err != nil {
		return fmt.Errorf("failed to update company profile: %w", err)
	}

	return nil
}

func (u *companyUsecase) UpdateAvatar(ctx context.Context, profID string, avFile *multipart.FileHeader) (string, error) {
	compUUID, err := uuid.Parse(profID)
	if err != nil {
		return "", fmt.Errorf("invalid profile ID format: %w", err)
	}

	const maxFileSize = 2 * 1024 * 1024 // Maksimal 2MB
	if avFile.Size > maxFileSize {
		return "", fmt.Errorf("ukuran file terlalu besar, maksimal 2MB")
	}

	url, err := u.storage.UploadFile(avFile, "companies")
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah logo ke storage: %w", err)
	}

	if err := u.companyRepo.UploadAvatar(ctx, compUUID.String(), url); err != nil {
		return "", fmt.Errorf("gagal memperbarui logo di database: %w", err)
	}

	return url, nil
}
