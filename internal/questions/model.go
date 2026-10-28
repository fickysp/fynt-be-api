package questions

import "github.com/google/uuid"

type Questions struct {
	ID              uuid.UUID `db:"id"`
	JobID           uuid.UUID `db:"job_id"`
	QuestionsDetail string    `db:"questions_detail"`
}

type DTOQuestions struct {
	JobID           string `json:"job_id" validate:"required"`
	QuestionsDetail string `json:"questions_detail" validate:"required"`
}
