package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/fickysp/fynt-be-api/internal/questions"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

type JobUsecase interface {
	GetAllCategories(ctx context.Context) ([]DTOJobCategoryResponses, error)
	GetAllJob(ctx context.Context) ([]DTOJobsResponses, error)
	CreateJob(ctx context.Context, accID string, req DTOJobRequest) error
	UpdateJob(ctx context.Context, jobIDStr string, accountID string, req *DTOJobUpdate) error
	GetJobByID(ctx context.Context, jobID string) (*DTOJobsResponses, error)
	GetJobSummary(ctx context.Context, compID string) (*DTOJobSummaryResponses, error)
	SearchJobs(ctx context.Context, keyword string) ([]DTOJobSearchResponses, error)
}

type SubmissionRepo interface {
}

type jobUsecase struct {
	jobRepo      JobRepository
	walletRepo   wallet.WalletRepository
	questionRepo questions.QuestionsRepository
	sanitizer    *bluemonday.Policy
}

func NewJobUsecase(jr JobRepository, wr wallet.WalletRepository, qr questions.QuestionsRepository) JobUsecase {
	return &jobUsecase{
		jobRepo:      jr,
		walletRepo:   wr,
		questionRepo: qr,
		sanitizer:    bluemonday.UGCPolicy(),
	}
}

func (u *jobUsecase) GetAllCategories(ctx context.Context) ([]DTOJobCategoryResponses, error) {
	categSlice, err := u.jobRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	var resp []DTOJobCategoryResponses
	for _, cat := range categSlice {
		resp = append(resp, DTOJobCategoryResponses{
			ID:   cat.ID.String(),
			Name: cat.Name,
		})
	}

	return resp, nil
}
func (u *jobUsecase) GetAllJob(ctx context.Context) ([]DTOJobsResponses, error) {
	jobSlice, err := u.jobRepo.GetAllJob(ctx)
	if err != nil {
		return nil, err
	}

	var resp []DTOJobsResponses
	for _, job := range jobSlice {
		resp = append(resp, DTOJobsResponses{
			ID:            job.ID,
			CompanyID:     job.CompanyID,
			JobCategoryID: job.JobCategoryID,
			Position:      job.Position,
			Location:      job.Location,
			Workplace:     job.Workplace,
			Employment:    job.Employment,
			SalaryDesc:    job.SalaryDesc,
			JobDesc:       job.JobDesc,
			Status:        job.Status,
			Slot:          job.Slot,
			FilledSlot:    job.FilledSlot,
			CoinApplyment: job.CoinApplyment,
			ClosingDate:   job.ClosingDate,
			CreatedAt:     job.CreatedAt,
			UpdatedAt:     job.UpdatedAt,
		})
	}

	return resp, nil
}

func (u *jobUsecase) CreateJob(ctx context.Context, accId string, req DTOJobRequest) error {
	companyUUID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return errors.New("company id invalid")
	}
	categoryUUID, err := uuid.Parse(req.JobCategoryID)
	if err != nil {
		return errors.New("category id invalid")
	}

	wallet, err := u.walletRepo.GetWalletByID(ctx, accId)
	if err != nil {
		return err
	}

	cost := int64(5) + req.ExtendCoins
	if wallet.Balance < cost {
		return errors.New("you do not have enough coins to create a job.")
	}

	closingDate := time.Now().AddDate(0, 0, 30+int(req.ExtendCoins*7))
	jobUUID := uuid.New()

	job := &Job{
		ID:            jobUUID,
		CompanyID:     companyUUID,
		JobCategoryID: categoryUUID,
		Position:      req.Position,
		Location:      req.Location,
		Workplace:     req.Workplace,
		Employment:    req.Employment,
		SalaryDesc:    req.SalaryDesc,
		JobDesc:       u.sanitizer.Sanitize(req.JobDesc),
		Status:        "active",
		Slot:          req.Slot,
		CoinApplyment: 1,
		ClosingDate:   &closingDate,
	}

	var essayQuestions []questions.Questions
	for _, q := range req.EssayQuestions {
		if strings.TrimSpace(q) == "" {
			continue
		}
		id := uuid.New()
		essayQuestions = append(essayQuestions, questions.Questions{
			ID:              &id,
			JobID:           &jobUUID,
			QuestionsDetail: &q,
		})
	}

	return u.jobRepo.CreateJobWithQuestionsTx(ctx, job, essayQuestions, accId, wallet.Balance-cost)
}

func (u *jobUsecase) UpdateJob(
	ctx context.Context,
	jobIDStr string,
	accountID string,
	req *DTOJobUpdate,
) error {

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return fmt.Errorf("format ID lowongan tidak valid: %w", err)
	}

	job, err := u.jobRepo.GetJobById(ctx, jobIDStr)
	if err != nil {
		return errors.New("lowongan tidak ditemukan")
	}

	if job.CompanyID.String() != *req.CompanyID {
		return errors.New("Anda tidak memiliki akses untuk memperbarui lowongan ini")
	}

	var jobDesc *string
	var status *string
	var extendCoins *int
	var newBalance *int64

	if req.JobDesc != nil && *req.JobDesc != "" {
		sanitizedDesc := u.sanitizer.Sanitize(*req.JobDesc)
		jobDesc = &sanitizedDesc
	}

	if req.Status != nil && *req.Status != "" {
		status = req.Status
	}

	if req.ExtendCoins != nil && *req.ExtendCoins > 0 {
		wallet, err := u.walletRepo.GetWalletByID(ctx, accountID)
		if err != nil {
			return fmt.Errorf("gagal mengecek saldo wallet: %w", err)
		}

		if wallet.Balance < int64(*req.ExtendCoins) {
			return fmt.Errorf("saldo tidak mencukupi untuk menambah %d koin lamaran", req.ExtendCoins)
		}
		calculatedBalance := wallet.Balance - int64(*req.ExtendCoins)
		newBalance = &calculatedBalance
		extendCoins = req.ExtendCoins
	}

	if jobDesc == nil && status == nil && extendCoins == nil {
		return errors.New("tidak ada data lowongan yang diperbarui")
	}

	var jobDescVal string
	var statusVal string
	var extendCoinsVal int

	if jobDesc != nil {
		jobDescVal = *jobDesc
	}
	if status != nil {
		statusVal = *status
	}
	if extendCoins != nil {
		extendCoinsVal = *extendCoins
	}

	err = u.jobRepo.UpdateJobDetails(ctx, jobID, accountID, jobDescVal, statusVal, extendCoinsVal, newBalance)

	if err != nil {
		slog.Error(
			"Usecase gagal memperbarui lowongan",
			"job_id", jobIDStr,
			"account_id", accountID,
			"error", err,
		)

		return fmt.Errorf("gagal memperbarui data lowongan: %w", err)
	}

	return nil
}

func (u *jobUsecase) GetJobByID(ctx context.Context, jobID string) (*DTOJobsResponses, error) {
	job, err := u.jobRepo.GetJobById(ctx, jobID)
	if err != nil {
		return nil, err
	}

	response := &DTOJobsResponses{
		ID:            job.ID,
		CompanyID:     job.CompanyID,
		JobCategoryID: job.JobCategoryID,
		Position:      job.Position,
		Location:      job.Location,
		Workplace:     job.Workplace,
		Employment:    job.Employment,
		SalaryDesc:    job.SalaryDesc,
		JobDesc:       job.JobDesc,
		Status:        job.Status,
		Slot:          job.Slot,
		FilledSlot:    job.FilledSlot,
		CoinApplyment: job.CoinApplyment,
		ClosingDate:   job.ClosingDate,
		CreatedAt:     job.CreatedAt,
		UpdatedAt:     job.UpdatedAt,
	}

	return response, nil
}

func (u *jobUsecase) GetJobSummary(ctx context.Context, compID string) (*DTOJobSummaryResponses, error) {
	stats, err := u.jobRepo.GetJobStatistics(ctx, compID)
	if err != nil {
		return &DTOJobSummaryResponses{}, err
	}

	return stats, nil
}

func (u *jobUsecase) SearchJobs(ctx context.Context, keyword string) ([]DTOJobSearchResponses, error) {
	searchKeyword := "%" + keyword + "%"
	jobs, err := u.jobRepo.SearchJobs(ctx, searchKeyword)
	if err != nil {
		return nil, err
	}

	if jobs == nil {
		return []DTOJobSearchResponses{}, nil
	}

	return jobs, nil
}
