package jobs

import (
	"context"
	"errors"
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
	UpdateJob(ctx context.Context, accID, jobID string, req DTOUpdateJobRequest) error
	GetJobByID(ctx context.Context, jobID string) (*DTOJobsResponses, error)
	GetJobSummary(ctx context.Context, jobID string) (*DTOJobSummaryResponses, error)
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
		essayQuestions = append(essayQuestions, questions.Questions{
			ID:              uuid.New(),
			JobID:           jobUUID,
			QuestionsDetail: u.sanitizer.Sanitize(q),
		})
	}

	return u.jobRepo.CreateJobWithQuestionsTx(ctx, job, essayQuestions, accId, wallet.Balance-cost)
}

func (u *jobUsecase) UpdateJob(ctx context.Context, accID, jobID string, req DTOUpdateJobRequest) error {
	existingJob, err := u.jobRepo.GetJobById(ctx, jobID)
	if err != nil {
		return errors.New("job not found")
	}

	var newBalance *int64

	if req.ExtendCoins > 0 {
		wallet, err := u.walletRepo.GetWalletByID(ctx, accID)
		if err != nil {
			return err
		}

		if wallet.Balance < req.ExtendCoins {
			return errors.New("you do not have enough coins to create a job.")
		}

		calcBalance := wallet.Balance - req.ExtendCoins
		newBalance = &calcBalance

		newClosingDate := existingJob.ClosingDate.AddDate(0, 0, int(req.ExtendCoins*7))
		existingJob.ClosingDate = &newClosingDate
	}

	if req.JobCategoryID != nil {
		catUUID, _ := uuid.Parse(*req.JobCategoryID)
		existingJob.JobCategoryID = catUUID
	}
	if req.Position != nil {
		existingJob.Position = *req.Position
	}
	if req.Location != nil {
		existingJob.Location = *req.Location
	}
	if req.Workplace != nil {
		existingJob.Workplace = *req.Workplace
	}
	if req.Employment != nil {
		existingJob.Employment = *req.Employment
	}
	if req.SalaryDesc != nil {
		existingJob.SalaryDesc = *req.SalaryDesc
	}
	if req.JobDesc != nil {
		existingJob.JobDesc = u.sanitizer.Sanitize(*req.JobDesc)
	}
	if req.Slot != nil {
		existingJob.Slot = req.Slot
	}

	var newQuestions []questions.Questions

	if req.EssayQuestions != nil {
		newQuestions = make([]questions.Questions, 0)
		for _, q := range req.EssayQuestions {
			if strings.TrimSpace(q) == "" {
				continue
			}
			newQuestions = append(newQuestions, questions.Questions{
				ID:              uuid.New(),
				JobID:           existingJob.ID,
				QuestionsDetail: u.sanitizer.Sanitize(q),
			})
		}
	}

	return u.jobRepo.UpdateJobWithQuestionsTx(ctx, existingJob, newQuestions, accID, newBalance)
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
	allJobs, err := u.jobRepo.CountAllJobs(ctx, compID)
	if err != nil {
		return nil, err
	}
	activeJobs, err := u.jobRepo.CountActiveJobs(ctx, compID)
	if err != nil {
		return nil, err
	}
	closeJobs, err := u.jobRepo.CountCloseJobs(ctx, compID)
	if err != nil {
		return nil, err
	}

	return &DTOJobSummaryResponses{
		ActiveJobs: *activeJobs,
		AllJobs:    *allJobs,
		CLoseJobs:  *closeJobs,
	}, nil
}
