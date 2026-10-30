package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	RefreshToken(ctx context.Context, refreshToken, jwtSecret, refreshSecret string) (*DTORefreshResponse, error)
	OauthGoogleSeekerLogin(ctx context.Context, refeshToken string, userInfo DTOGoogleUserInfo) (string, error)
	OauthGoogleRecruiterLogin(ctx context.Context, refeshToken string, userInfo DTOGoogleUserInfo) (string, error)
	CreateExchangeCode(ctx context.Context, internalJWT string) (string, error)
	ExchangeCodeForToken(ctx context.Context, code string) (string, error)
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
		slog.Info("registration rejected: email is already in use", "email", req.Email)
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
	account, err := u.authRepo.FindByEmail(ctx, req.Email)
	if account == nil && err != nil {
		return nil, sentinel.ErrInternalServerError
	} else if account == nil && err == nil {
		slog.Warn("login failed: account not found", "email", req.Email)
		return nil, sentinel.ErrUserNotFound
	}

	if !bcrypt.CheckPassword(*account.Password, req.Password) {
		return nil, sentinel.ErrInvalidCredentials
	}

	var profileData any
	var profID string
	switch account.Role {
	case "seeker":
		applicant, err := u.profilRepo.GetApplicantByAccountID(ctx, account.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get applicant profile: %w", err)
		}
		profileData = applicant
		profID = applicant.ID.String()

	case "recruiter":
		company, err := u.profilRepo.GetCompanyByAccountID(ctx, account.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get company profile: %w", err)
		}
		profileData = company
		profID = company.ID.String()
	}
	accessToken, err := jwt.GenerateAccessToken(account.ID.String(), string(account.Role), profID, jwtSecret, account.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken(account.ID.String(), string(account.Role), profID, refreshSecret, account.UpdatedAt, req.RememberMe)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := u.authRepo.SaveRefresh(ctx, refreshToken, account.ID.String()); err != nil {
		slog.Error("gagal menyimpan refresh token", "error", err, "account_id", account.ID)
		return &DTOLoginResponse{}, err
	}

	data := &DataLogin{
		Profile: profileData,
	}

	return &DTOLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         string(account.Role),
		Data:         *data,
	}, nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken, jwtSecret, refreshSecret string) (*DTORefreshResponse, error) {
	claims, err := jwt.ValidateToken(refreshToken, refreshSecret)
	if err != nil {
		slog.Warn("refresh token validation failed", "error", err)
		return nil, err
	}

	if claims.Type != "refresh_token" {
		slog.Error("refresh token rejected: invalid token type", "token_type", claims.Type)
		return nil, sentinel.ErrUnauthorized
	}

	account, err := u.authRepo.GetById(ctx, claims.ID)
	if err != nil {
		slog.Error("failed to retrieve account for refresh token", "account_id", claims.ID, "error", err)
		return nil, err
	}

	if account.RefreshToken == nil || *account.RefreshToken != refreshToken {
		slog.Error("refresh token rejected: token does not match the stored token", "account_id", claims.ID)
		return nil, sentinel.ErrUnauthorized
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

	accToken, err := jwt.GenerateAccessToken(account.ID.String(), string(account.Role), profID, jwtSecret, account.UpdatedAt)
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
func (u *authUsecase) OauthGoogleSeekerLogin(ctx context.Context, refreshSecret string, userInfo DTOGoogleUserInfo) (string, error) {
	account, err := u.authRepo.FindByEmail(ctx, userInfo.Email)

	var strAccID string
	var profID string
	var tokenTimestamp time.Time

	if account == nil {
		authProvider := "google"
		acc := &Account{
			Email:        userInfo.Email,
			AuthProvider: &authProvider,
			Role:         "seeker",
		}

		accID, err := u.authRepo.Create(ctx, acc)
		if err != nil {
			return "", err
		}
		strAccID = accID
		uuidAccID, _ := uuid.Parse(accID)

		applicantData := &applicant.Applicant{
			AccountID: uuidAccID,
			FirstName: &userInfo.GivenName,
			LastName:  &userInfo.FamilyName,
			AvatarURL: &userInfo.Picture,
		}

		if err := u.profilRepo.CreateApplicant(ctx, applicantData); err != nil {
			return "", err
		}

		if err := u.walletRepo.Create(ctx, uuidAccID, 2); err != nil {
			return "", err
		}

		apl, _ := u.profilRepo.GetApplicantByAccountID(ctx, strAccID)
		profID = apl.ID.String()
		tokenTimestamp = time.Now()

	} else {
		if account.Role != "seeker" {
			return "", errors.New("new message")
		}

		strAccID = account.ID.String()
		tokenTimestamp = account.UpdatedAt

		apl, err := u.profilRepo.GetApplicantByAccountID(ctx, strAccID)
		if err != nil {
			return "", err
		}
		profID = apl.ID.String()
	}

	token, err := jwt.GenerateAccessToken(strAccID, "seeker", profID, refreshSecret, tokenTimestamp)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *authUsecase) OauthGoogleRecruiterLogin(ctx context.Context, refreshSecret string, userInfo DTOGoogleUserInfo) (string, error) {
	account, err := u.authRepo.FindByEmail(ctx, userInfo.Email)

	var strAccID string
	var profID string
	var tokenTimestamp time.Time

	if account == nil {
		authProvider := "google"
		acc := &Account{
			Email:        userInfo.Email,
			AuthProvider: &authProvider,
			Role:         "recruiter",
		}

		accID, err := u.authRepo.Create(ctx, acc)
		if err != nil {
			return "", err
		}
		strAccID = accID
		tokenTimestamp = time.Now()

		profID = ""

	} else {
		if account.Role != "recruiter" {
			return "", errors.New("email ini sudah terdaftar sebagai seeker")
		}

		strAccID = account.ID.String()
		tokenTimestamp = account.UpdatedAt

		company, err := u.profilRepo.GetCompanyByAccountID(ctx, strAccID)
		if err == nil && company != nil {
			profID = company.ID.String()
		}
	}

	token, err := jwt.GenerateAccessToken(strAccID, "recruiter", profID, refreshSecret, tokenTimestamp)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *authUsecase) CreateExchangeCode(ctx context.Context, internalJWT string) (string, error) {
	exchangeCode := uuid.New().String()
	expiresAt := time.Now().Add(2 * time.Minute)

	err := u.authRepo.SaveExchangeCode(ctx, exchangeCode, internalJWT, expiresAt)
	if err != nil {
		slog.Error("failed to save exchange code", "error", err.Error())
		return "", err
	}

	return exchangeCode, nil
}

func (u *authUsecase) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	token, err := u.authRepo.GetAndConsumeExchangeCode(ctx, code)
	if err != nil {
		slog.Warn("invalid or expired exchange code attempted", "code", code, "error", err.Error())
		return "", err
	}

	return token, nil
}
