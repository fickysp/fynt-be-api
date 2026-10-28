package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/company"
	"github.com/fickysp/fynt-be-api/internal/sectors"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/fickysp/fynt-be-api/pkg/bcrypt"
	"github.com/fickysp/fynt-be-api/pkg/jwt"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, jwtSecret, refreshSecret string, req DTOLoginRequest) (*DTOLoginResponse, error)
	RegisterApplicant(ctx context.Context, req DTOApplicantRegisterRequest) error
	RegisterCompany(ctx context.Context, req DTOCompanyRegisterRequest) error
	Logout(ctx context.Context, accID string) error
}
type authUsecase struct {
	authRepo   AuthRepository
	profilRepo ProfileRepository
	walletRepo wallet.WalletRepository
	sectorRepo sectors.SectorRepository
}

func NewAuthUsecase(ar AuthRepository, pr ProfileRepository, wr wallet.WalletRepository, sr sectors.SectorRepository) AuthUsecase {
	return &authUsecase{
		authRepo:   ar,
		profilRepo: pr,
		walletRepo: wr,
		sectorRepo: sr,
	}
}

func (u *authUsecase) RegisterApplicant(ctx context.Context, req DTOApplicantRegisterRequest) error {
	acc, err := u.authRepo.FindByEmail(ctx, req.Email)
	if acc != nil {
		return sentinel.ErrEmailIsUsed
	} else if err != nil {
		return sentinel.ErrInternalServerError
	}

	hashPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		Email:       req.Email,
		Password:    &hashPassword,
		PhoneNumber: &req.PhoneNumber,
		Role:        RoleSeeker,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	accountID, err := u.authRepo.Create(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to create applicant account: %w", err)
	}

	uuidAccID, err := uuid.Parse(accountID)
	if err != nil {
		return fmt.Errorf("failed to parse account id: %w", err)
	}

	applicantData := &applicant.Applicant{
		AccountID: uuidAccID,
		FirstName: &req.FirstName,
		LastName:  &req.LastName,
	}

	if err := u.profilRepo.CreateApplicant(ctx, applicantData); err != nil {
		return fmt.Errorf("failed to create applicant profile: %w", err)
	}

	if err := u.walletRepo.Create(ctx, uuidAccID, 2); err != nil {
		return fmt.Errorf("failed to create applicant wallet: %w", err)
	}

	return nil
}

func (u *authUsecase) RegisterCompany(ctx context.Context, req DTOCompanyRegisterRequest) error {
	acc, err := u.authRepo.FindByEmail(ctx, req.Email)
	if acc != nil {
		slog.Info("email is used")
		return sentinel.ErrEmailIsUsed
	} else if err != nil {
		return sentinel.ErrInternalServerError
	}

	hashPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		Email:       req.Email,
		Password:    &hashPassword,
		PhoneNumber: &req.PhoneNumber,
		Role:        RoleRecruiter,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	accountID, err := u.authRepo.Create(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to create company account: %w", err)
	}

	uuidAccID, err := uuid.Parse(accountID)
	if err != nil {
		return fmt.Errorf("failed to parse account id: %w", err)
	}

	var finalSectorID uuid.UUID
	sectorData, err := u.sectorRepo.GetSectorByName(ctx, req.Sector)

	if sectorData == nil {
		newSectorID, err := u.sectorRepo.Create(ctx, req.Sector)
		if err != nil {
			return fmt.Errorf("failed to create new sector: %w", err)
		}
		finalSectorID, err = uuid.Parse(newSectorID)
		if err != nil {
			return fmt.Errorf("failed to parse new sector id: %w", err)
		}
	} else {
		finalSectorID = sectorData.ID
	}

	companyData := &company.Company{
		AccountID:   uuidAccID,
		SectorID:    finalSectorID,
		CompanyName: req.CompanyName,
		Domicile:    req.Domicile,
		WebURL:      &req.WebURL,
		Address:     req.Address,
		OwnerName:   req.OwnerName,
	}

	if err := u.profilRepo.CreateCompany(ctx, companyData); err != nil {
		return fmt.Errorf("failed to create company profile: %w", err)
	}

	if err := u.walletRepo.Create(ctx, uuidAccID, 5); err != nil {
		return fmt.Errorf("failed to create company wallet: %w", err)
	}

	return nil
}

func (u *authUsecase) Login(ctx context.Context, jwtSecret, refreshSecret string, req DTOLoginRequest) (*DTOLoginResponse, error) {
	profile, err := u.authRepo.FindByEmail(ctx, req.Email)
	if profile == nil && err != nil {
		return nil, sentinel.ErrInternalServerError
	} else if profile == nil && err == nil {
		slog.Warn("failed to find user:", "email_account", req.Email)
		return nil, sentinel.ErrUserNotFound
	}

	if !bcrypt.CheckPassword(*profile.Password, req.Password) {
		return nil, sentinel.ErrInvalidCredentials
	}

	var profileData any

	switch profile.Role {
	case "seeker":
		applicant, err := u.profilRepo.GetApplicantByAccountID(ctx, profile.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get applicant profile: %w", err)
		}
		profileData = applicant

	case "recruiter":
		company, err := u.profilRepo.GetCompanyByAccountID(ctx, profile.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get company profile: %w", err)
		}
		profileData = company
	}

	accessToken, err := jwt.GenerateAccessToken(profile.ID.String(), string(profile.Role), profile.ID.String(), jwtSecret, profile.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken(profile.ID.String(), string(profile.Role), profile.ID.String(), refreshSecret, profile.UpdatedAt, req.RememberMe)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	data := &DataLogin{
		Profile: profileData,
	}

	return &DTOLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Data:         *data,
	}, nil
}
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken, refreshSecret string) (*DTORefreshResponse, error) {
	claims, err := jwt.ValidateToken(refreshToken, os.Getenv("REFRESH_TOKEN_SECRET"))
	if err != nil {
		slog.Warn("invalid token", "warn", err)
		return nil, err
	}

	if claims.Type != "refresh" {
		slog.Error("token type invalid", "error", err)
		return nil, err
	}

	account, err := u.authRepo.GetById(ctx, claims.ID)
	if err != nil {
		slog.Error("failed to get account", "error", err)
		return nil, err
	}

	if *account.RefreshToken != refreshToken {
		slog.Error("refresh token invalid or expired", "error", err)
		return nil, err
	}

	var profID string

	switch account.Role {
	case "seeker":
		applicant, err := u.profilRepo.GetApplicantByAccountID(ctx, account.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get applicant profile: %w", err)
		}
		profID = applicant.ID.String()

	case "recruiter":
		company, err := u.profilRepo.GetCompanyByAccountID(ctx, account.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get company profile: %w", err)
		}
		profID = company.ID.String()
	}

	accToken, err := jwt.GenerateAccessToken(account.ID.String(), string(account.Role), profID, refreshSecret, account.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &DTORefreshResponse{
		AccessToken: accToken,
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, accID string) error {
	return u.authRepo.RemoveRefreshToken(ctx, accID)
}
