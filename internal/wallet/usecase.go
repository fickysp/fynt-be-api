package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/company"
	"github.com/fickysp/fynt-be-api/internal/transaction"
)

type PaymentGateway interface {
	CreatePaymentURL(orderID string, amount int64, customerName string) (string, string, error)
}

type ProfileRepository interface {
	GetCompanyByAccountID(ctx context.Context, accID string) (*company.Company, error)
	GetApplicantByAccountID(ctx context.Context, accID string) (*applicant.Applicant, error)
}

type WalletUsecase interface {
	GetBalance(ctx context.Context, accID string) (*DTOBalanceResponse, error)
	RequestTopUp(ctx context.Context, role string, req DTOTopUpRequest) (*DTOTopUpResponse, error)
}

type walletUsecase struct {
	walletRepo WalletRepository
	transRepo  transaction.TransactionRepository
	pgClient   PaymentGateway
	profilRepo ProfileRepository
}

func NewWalletUsecase(wr WalletRepository, tr transaction.TransactionRepository, pg PaymentGateway, pr ProfileRepository) WalletUsecase {
	return &walletUsecase{
		walletRepo: wr,
		transRepo:  tr,
		pgClient:   pg,
		profilRepo: pr,
	}
}

func (u *walletUsecase) GetBalance(ctx context.Context, accID string) (*DTOBalanceResponse, error) {
	walletData, err := u.walletRepo.GetWalletByID(ctx, accID)
	if err != nil {
		return nil, errors.New("failed to retrieve wallet balance")
	}

	return &DTOBalanceResponse{
		Balance: walletData.Balance,
	}, nil
}

func (u *walletUsecase) RequestTopUp(ctx context.Context, role string, req DTOTopUpRequest) (*DTOTopUpResponse, error) {
	accUUID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errors.New("invalid account id")
	}

	if req.Amount < 10000 {
		return nil, errors.New("minimum top-up amount is 10000")
	}

	var customerName string
	if role == "seeker" {
		applicant, err := u.profilRepo.GetApplicantByAccountID(ctx, req.AccountID)
		if err != nil {
			return nil, errors.New("applicant profile not found")
		}
		customerName = strings.TrimSpace(*applicant.FirstName + " " + *applicant.LastName)
	} else if role == "recruiter" {
		company, err := u.profilRepo.GetCompanyByAccountID(ctx, req.AccountID)
		if err != nil {
			return nil, errors.New("company profile not found")
		}
		customerName = company.CompanyName
	} else {
		return nil, errors.New("unauthorized role for top-up")
	}

	orderID := fmt.Sprintf("TRX-FYNT-%d-%s", time.Now().Unix(), accUUID.String()[:8])

	trx := &transaction.Transaction{
		ID:        uuid.New(),
		AccountID: accUUID,
		OrderID:   orderID,
		Amount:    req.Amount,
		Status:    "pending",
	}

	if err := u.transRepo.Create(ctx, trx); err != nil {
		return nil, errors.New("failed to initialize transaction record")
	}

	snapToken, snapURL, err := u.pgClient.CreatePaymentURL(orderID, req.Amount, customerName)
	if err != nil {
		return nil, errors.New("payment gateway is currently unavailable")
	}

	if err := u.transRepo.UpdateSnap(ctx, orderID, snapToken, snapURL); err != nil {
		return nil, errors.New("failed to attach payment token to transaction")
	}

	return &DTOTopUpResponse{
		OrderID:   orderID,
		Amount:    req.Amount,
		SnapToken: snapToken,
		SnapURL:   snapURL,
	}, nil
}
