package wallet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
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
		return nil, errors.New("Gagal mengambil data wallet user")
	}

	return &DTOBalanceResponse{
		Balance: walletData.Balance,
	}, nil
}

func (u *walletUsecase) RequestTopUp(ctx context.Context, role string, req DTOTopUpRequest) (*DTOTopUpResponse, error) {
	accUUID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format account ID tidak valid")
	}

	if req.CoinAmount <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Minimum top-up adalah 1 koin")
	}
	if req.Amount < 10000 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Minimum top-up adalah Rp10.000")
	}

	var customerName string
	switch role {
	case "seeker":
		applicant, err := u.profilRepo.GetApplicantByAccountID(ctx, req.AccountID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusNotFound, "Profil pelamar tidak ditemukan")
		}
		firstName := ""
		lastName := ""
		if applicant.FirstName != nil {
			firstName = *applicant.FirstName
		}
		if applicant.LastName != nil {
			lastName = *applicant.LastName
		}
		customerName = strings.TrimSpace(firstName + " " + lastName)
		if customerName == "" {
			customerName = "Pengguna"
		}
	case "recruiter":
		company, err := u.profilRepo.GetCompanyByAccountID(ctx, req.AccountID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusNotFound, "Profil perusahaan tidak ditemukan")
		}
		customerName = company.CompanyName
	default:
		return nil, fiber.NewError(fiber.StatusForbidden, "Role tidak diizinkan untuk top-up")
	}

	orderID := fmt.Sprintf("TRX-FYNT-%d-%s", time.Now().Unix(), accUUID.String()[:8])

	snapToken, snapURL, err := u.pgClient.CreatePaymentURL(orderID, req.Amount, customerName)
	if err != nil {
		slog.Error("gagal membuat payment URL", "error", err, "order_id", orderID)
		return nil, fiber.NewError(fiber.StatusServiceUnavailable, "Payment gateway tidak tersedia, coba beberapa saat lagi")
	}

	trx := &transaction.Transaction{
		ID:         uuid.New(),
		AccountID:  accUUID,
		OrderID:    orderID,
		CoinAmount: req.CoinAmount,
		Amount:     req.Amount,
		Status:     "pending",
		SnapToken:  &snapToken,
		SnapURL:    &snapURL,
	}

	if err := u.transRepo.Create(ctx, trx); err != nil {
		slog.Error("gagal simpan transaksi", "error", err, "order_id", orderID, "coin_ammount", req.CoinAmount)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan transaksi")
	}

	return &DTOTopUpResponse{
		OrderID:    orderID,
		CoinAmount: req.CoinAmount,
		Amount:     req.Amount,
		SnapToken:  snapToken,
		SnapURL:    snapURL,
	}, nil
}
