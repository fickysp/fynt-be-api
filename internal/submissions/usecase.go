package submissions

import (
	"context"
	"errors"

	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/jobs"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/google/uuid"
)

type submissionUsecase struct {
	submissionRepo SubmissionRepository
	jobRepo        jobs.JobRepository
	walletRepo     wallet.WalletRepository
	applicantRepo  applicant.ApplicantRepository
}

type SubmissionUsecase interface {
	ApplyJob(ctx context.Context, accountID, applicantID string, req ApplyJobRequest) error
}

func NewSubmissionUsecase(sr SubmissionRepository, jr jobs.JobRepository, wr wallet.WalletRepository, ar applicant.ApplicantRepository) SubmissionUsecase {
	return &submissionUsecase{
		submissionRepo: sr,
		jobRepo:        jr,
		walletRepo:     wr,
		applicantRepo:  ar,
	}
}

func (u *submissionUsecase) ApplyJob(ctx context.Context, accountID, applicantID string, req ApplyJobRequest) error {
	appUUID, err := uuid.Parse(applicantID)
	if err != nil {
		return errors.New("invalid applicant identity")
	}

	job, err := u.jobRepo.GetJobById(ctx, req.JobID)
	if err != nil {
		return errors.New("job not found")
	}

	if job.Status != "active" {
		return errors.New("job is already closed")
	}

	wallet, err := u.walletRepo.GetWalletByID(ctx, accountID)
	if err != nil {
		return errors.New("failed to load applicant wallet")
	}

	if wallet.Balance < job.CoinApplyment {
		return errors.New("insufficient coin balance to apply")
	}

	return u.submissionRepo.CreateSubmissionTx(ctx, req, appUUID, accountID, job.CoinApplyment)
}
