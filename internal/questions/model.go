package questions

import "github.com/google/uuid"

type Questions struct {
	ID              *uuid.UUID `db:"id" json:"id"`
	JobID           *uuid.UUID `db:"job_id" json:"job_id"`
	QuestionsDetail *string    `db:"question_detail" json:"question_detail"`
}

type DTOQuestions struct {
	JobID           string `json:"job_id" validate:"required"`
	QuestionsDetail string `json:"question_detail" validate:"required"`
}
type DTOUpdateQuestion struct {
	ID              *uuid.UUID `db:"id" json:"id"`
	JobID           *uuid.UUID `db:"job_id" json:"job_id"`
	QuestionsDetail *string    `db:"question_detail" json:"question_detail"`
}

type SwaggerQuestionsResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Get data success"`
	Data    []Questions `json:"data"`
}
