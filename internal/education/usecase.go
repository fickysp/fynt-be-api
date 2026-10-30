package education

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type EducationUsecase interface {
	AddEducation(ctx context.Context, profID string, req DTOEducationRequest) error
	RemoveEducation(ctx context.Context, profID, eduID string) error
	UpdateEducation(ctx context.Context, profID, eduID string, req DTOUpdateEducationRequest) error
	GetEducations(ctx context.Context, profID string) ([]Education, error)
}

type educationUsecase struct {
	repo EducationRepository
}

func NewEducationUsecase(repo EducationRepository) EducationUsecase {
	return &educationUsecase{repo: repo}
}

func (u *educationUsecase) AddEducation(ctx context.Context, profID string, req DTOEducationRequest) error {
	appUUID, err := uuid.Parse(profID)
	if err != nil {
		return fmt.Errorf("invalid profile ID")
	}

	edu := &Education{
		ApplicantID:  appUUID,
		Institution:  req.Institution,
		Degree:       req.Degree,
		FieldOfStudy: req.FieldOfStudy,
		Periode:      req.Periode,
	}
	return u.repo.CreateEducation(ctx, edu)
}

func (u *educationUsecase) RemoveEducation(ctx context.Context, profID, eduID string) error {
	appUUID, _ := uuid.Parse(profID)
	eduUUID, err := uuid.Parse(eduID)
	if err != nil {
		return fmt.Errorf("invalid education ID")
	}
	return u.repo.DeleteEducation(ctx, eduUUID, appUUID)
}

func (u *educationUsecase) UpdateEducation(ctx context.Context, profID, eduID string, req DTOUpdateEducationRequest) error {
	appUUID, _ := uuid.Parse(profID)
	eduUUID, err := uuid.Parse(eduID)
	if err != nil {
		return fmt.Errorf("invalid education ID")
	}

	edu := &Education{
		ID:          eduUUID,
		ApplicantID: appUUID,
	}

	if req.Institution != nil {
		edu.Institution = *req.Institution
	}
	if req.Degree != nil {
		edu.Degree = *req.Degree
	}
	if req.FieldOfStudy != nil {
		edu.FieldOfStudy = req.FieldOfStudy
	}
	if req.Periode != nil {
		edu.Periode = *req.Periode
	}

	return u.repo.UpdateEducation(ctx, edu)
}

func (u *educationUsecase) GetEducations(ctx context.Context, profID string) ([]Education, error) {
	edus, err := u.repo.GetEducation(ctx, profID)
	if err != nil {
		return nil, err
	}

	if edus == nil {
		return []Education{}, nil
	}
	return edus, nil
}
