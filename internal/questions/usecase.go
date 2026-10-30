package questions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type QuestionUsecase interface {
	GetQuestionsByJob(ctx context.Context, jobID string) ([]Questions, error)
	EditQuestion(ctx context.Context, jobId string, req []DTOUpdateQuestion) error
	RemoveQuestion(ctx context.Context, questionID string) error
}

type questionUsecase struct {
	repo QuestionsRepository
}

func NewQuestionUsecase(repo QuestionsRepository) QuestionUsecase {
	return &questionUsecase{repo: repo}
}

func (u *questionUsecase) GetQuestionsByJob(ctx context.Context, jobID string) ([]Questions, error) {
	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format Job ID tidak valid")
	}

	qs, err := u.repo.GetByJobID(ctx, jobUUID)
	if err != nil {
		return nil, err
	}

	if qs == nil {
		return []Questions{}, nil
	}
	return qs, nil
}

func (u *questionUsecase) EditQuestion(ctx context.Context, jobID string, req []DTOUpdateQuestion) error {
	if len(req) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Data pertanyaan tidak boleh kosong")
	}

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Format Job ID tidak valid")
	}

	for i, q := range req {
		if q.ID == nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("ID pertanyaan ke-%d tidak boleh kosong", i+1))
		}
		if q.QuestionsDetail == nil || strings.TrimSpace(*q.QuestionsDetail) == "" {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Detail pertanyaan ke-%d tidak boleh kosong", i+1))
		}
	}

	if err := u.repo.Update(ctx, jobUUID, req); err != nil {
		if strings.Contains(err.Error(), "bukan milik job ini") {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
		return fmt.Errorf("gagal update questions: %w", err)
	}

	return nil
}

func (u *questionUsecase) RemoveQuestion(ctx context.Context, questionID string) error {
	qUUID, err := uuid.Parse(questionID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Format Question ID tidak valid")
	}

	return u.repo.Delete(ctx, qUUID)
}
