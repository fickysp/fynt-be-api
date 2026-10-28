package auth

import (
	"context"

	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/company"
)

type ProfileRepository interface {
	CreateApplicant(ctx context.Context, applicant *applicant.Applicant) error
	CreateCompany(ctx context.Context, company *company.Company) error
	GetCompanyByAccountID(ctx context.Context, accID string) (*company.Company, error)
	GetApplicantByAccountID(ctx context.Context, accID string) (*applicant.Applicant, error)
}

type profileRepository struct {
	companyRepo   company.CompanyRepository
	applicantRepo applicant.ApplicantRepository
}

func NewProfileRepositoryAdapter(cR company.CompanyRepository, aR applicant.ApplicantRepository) ProfileRepository {
	return &profileRepository{
		companyRepo:   cR,
		applicantRepo: aR,
	}
}

func (r *profileRepository) GetCompanyByAccountID(ctx context.Context, accID string) (*company.Company, error) {
	return r.companyRepo.GetCompanyByAccountID(ctx, accID)
}
func (r *profileRepository) GetApplicantByAccountID(ctx context.Context, accID string) (*applicant.Applicant, error) {
	return r.applicantRepo.GetApplicantByAccountID(ctx, accID)
}
func (r *profileRepository) CreateApplicant(ctx context.Context, app *applicant.Applicant) error {
	return r.applicantRepo.Create(ctx, app)
}
func (r *profileRepository) CreateCompany(ctx context.Context, comp *company.Company) error {
	return r.companyRepo.Create(ctx, comp)
}
