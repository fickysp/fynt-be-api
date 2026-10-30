package submissions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/jobs"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/google/uuid"
)

type SubmissionUsecase interface {
	ApplyJob(ctx context.Context, accountID, applicantID string, req ApplyJobRequest) error
	GetMySubmissions(ctx context.Context, applicantID, statusFilter, keyword string) ([]DTOSeekerSubmissionResponse, error)
	GetSubmissionDetail(ctx context.Context, applicantID, submissionID string) (*DTOSeekerSubmissionDetail, error)
	GetSubmissionsByJobID(ctx context.Context, compID, jobID, statusFilter string) ([]DTORecruiterSubmissionList, error)
	UpdateSubmissionStatus(ctx context.Context, compID, submissionID, status string) error
	UpdateInternalNotes(ctx context.Context, compID, submissionID, notes string) error
	GetApplicantEssaysByJob(ctx context.Context, jobID string) ([]DTOSubmissionEssay, error)
}
type submissionUsecase struct {
	submissionRepo SubmissionRepository
	jobRepo        jobs.JobRepository
	walletRepo     wallet.WalletRepository
	applicantRepo  applicant.ApplicantRepository
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
		slog.Error("format applicant_id tidak valid", "applicant_id", applicantID, "error", err)
		return errors.New("User tidak diketahui, silahkan login kembali")
	}

	job, err := u.jobRepo.GetJobById(ctx, req.JobID)
	if err != nil {
		slog.Warn("lowongan tidak ditemukan saat melamar", "job_id", req.JobID, "error", err)
		return errors.New("Lowongan tidak ditemukan")
	}

	if job.Status != "active" {
		slog.Warn("lowongan sudah ditutup saat melamar", "job_id", req.JobID, "status", job.Status)
		return errors.New("Lowongan sudah ditutup")
	}

	if job.Slot != nil && job.FilledSlot >= *job.Slot {
		slog.Warn("kuota lowongan sudah penuh", "job_id", req.JobID)
		return errors.New("Kuota pelamar untuk lowongan ini sudah penuh")
	}

	wallet, err := u.walletRepo.GetWalletByID(ctx, accountID)
	if err != nil {
		slog.Error("gagal melihat saldo wallet pelamar", "account_id", accountID, "error", err)
		return errors.New("Gagal melihat saldo pelamar")
	}

	if wallet.Balance < job.CoinApplyment {
		slog.Warn("saldo koin tidak cukup untuk melamar", "balance", wallet.Balance, "cost", job.CoinApplyment, "account_id", accountID)
		return errors.New("Koin tidak cukup, silahkan isi ulang koin terlebih dahulu")
	}

	return u.submissionRepo.CreateSubmissionTx(ctx, req, appUUID, accountID, job.CoinApplyment)
}

func (u *submissionUsecase) GetMySubmissions(ctx context.Context, applicantID, statusFilter, keyword string) ([]DTOSeekerSubmissionResponse, error) {
	subs, err := u.submissionRepo.GetSeekerSubmissions(ctx, applicantID, statusFilter, keyword)
	if err != nil {
		slog.Error("gagal mengambil data lamaran dari repo", "applicant_id", applicantID, "error", err)
		return nil, err
	}
	if subs == nil {
		return []DTOSeekerSubmissionResponse{}, nil
	}
	return subs, nil
}

func (u *submissionUsecase) GetSubmissionDetail(ctx context.Context, applicantID, submissionID string) (*DTOSeekerSubmissionDetail, error) {
	detail, err := u.submissionRepo.GetSubmissionDetail(ctx, submissionID, applicantID)
	if err != nil {
		slog.Error("gagal mengambil detail lamaran dari repo", "submission_id", submissionID, "applicant_id", applicantID, "error", err)
		return nil, err
	}
	return detail, nil
}

func (u *submissionUsecase) GetSubmissionsByJobID(ctx context.Context, compID, jobID, statusFilter string) ([]DTORecruiterSubmissionList, error) {
	subs, err := u.submissionRepo.GetSubmissionsByJobID(ctx, jobID, compID, statusFilter)
	if err != nil {
		slog.Error("gagal mengambil daftar pelamar lowongan dari repo", "job_id", jobID, "company_id", compID, "error", err)
		return nil, err
	}
	if subs == nil {
		return []DTORecruiterSubmissionList{}, nil
	}
	return subs, nil
}

func (u *submissionUsecase) UpdateSubmissionStatus(ctx context.Context, compID, submissionID, status string) error {
	if err := u.submissionRepo.UpdateSubmissionStatus(ctx, submissionID, compID, status); err != nil {
		slog.Error("gagal update status lamaran di repo", "submission_id", submissionID, "company_id", compID, "error", err)
		return err
	}
	return nil
}

func (u *submissionUsecase) UpdateInternalNotes(ctx context.Context, compID, submissionID, notes string) error {
	if err := u.submissionRepo.UpdateInternalNotes(ctx, submissionID, compID, notes); err != nil {
		slog.Error("gagal update catatan internal di repo", "submission_id", submissionID, "company_id", compID, "error", err)
		return err
	}
	return nil
}

func (u *submissionUsecase) GetApplicantEssaysByJob(ctx context.Context, jobID string) ([]DTOSubmissionEssay, error) {
	rows, err := u.submissionRepo.GetSubmissionsEssay(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data pelamar: %w", err)
	}

	groupedData := make(map[string]*DTOSubmissionEssay)

	for _, row := range rows {
		if _, exists := groupedData[row.SubmissionID]; !exists {
			var avatar string
			if row.AvatarURL != nil {
				avatar = *row.AvatarURL
			}

			var domicile string
			if row.Domicile != nil {
				domicile = *row.Domicile
			}
			groupedData[row.SubmissionID] = &DTOSubmissionEssay{
				SubmissionID: row.SubmissionID,
				JobID:        row.JobID,
				AvatarURL:    avatar,
				Fullname:     row.Fullname,
				Domicile:     domicile,
				Essays:       []DTOEssayQA{},
			}
		}

		groupedData[row.SubmissionID].Essays = append(groupedData[row.SubmissionID].Essays, DTOEssayQA{
			Question: row.Question,
			Answer:   row.Answer,
		})
	}

	var result []DTOSubmissionEssay
	for _, data := range groupedData {
		result = append(result, *data)
	}

	if result == nil {
		result = []DTOSubmissionEssay{}
	}

	return result, nil
}
