package experiences

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ExperienceUsecase interface {
	AddExperience(ctx context.Context, profID string, req DTOExperienceRequest) error
	RemoveExperience(ctx context.Context, profID, expID string) error
	UpdateExperience(ctx context.Context, profID, expID string, req DTOUpdateExperienceRequest) error
	GetExperiences(ctx context.Context, profID string) ([]Experience, error)
}

type experienceUsecase struct {
	repo ExperienceRepository
}

func NewExperienceUsecase(repo ExperienceRepository) ExperienceUsecase {
	return &experienceUsecase{repo: repo}
}

func (u *experienceUsecase) AddExperience(ctx context.Context, profID string, req DTOExperienceRequest) error {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid profile ID")
	}

	exp := &Experience{
		ApplicantID: appUUID,
		Position:    req.Position,
		CompanyName: req.CompanyName,
		Periode:     req.Periode,
		Jobdesc:     req.Jobdesc,
	}

	return u.repo.CreateExperience(ctx, exp)
}

func (u *experienceUsecase) RemoveExperience(ctx context.Context, profID, expID string) error {
	appUUID, _ := uuid.Parse(profID)
	expUUID, err := uuid.Parse(expID)
	if err != nil {
		return fmt.Errorf("invalid experience ID")
	}
	return u.repo.DeleteExperience(ctx, expUUID, appUUID)
}
func (u *experienceUsecase) UpdateExperience(ctx context.Context, profID, expID string, req DTOUpdateExperienceRequest) error {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid applicant ID")
	}

	expUUID, err := uuid.Parse(expID)
	if err != nil {
		return fmt.Errorf("invalid experience ID")
	}

	exp := &Experience{
		ID:          expUUID,
		ApplicantID: appUUID,
	}

	if req.Position != nil {
		exp.Position = *req.Position
	}
	if req.CompanyName != nil {
		exp.CompanyName = *req.CompanyName
	}
	if req.Jobdesc != nil {
		exp.Jobdesc = req.Jobdesc
	}
	if req.Periode != nil {
		exp.Periode = *req.Periode
	}

	return u.repo.UpdateExperience(ctx, exp)
}
func (u *experienceUsecase) GetExperiences(ctx context.Context, profID string) ([]Experience, error) {
	exps, err := u.repo.GetExperience(ctx, profID)
	if err != nil {
		return nil, err
	}
	if exps == nil {
		return []Experience{}, err
	}
	return exps, nil
}
