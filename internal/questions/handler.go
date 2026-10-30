package questions

import (
	"errors"
	"log/slog"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/gofiber/fiber/v3"
)

type QuestionHandler struct {
	usecase QuestionUsecase
}

func NewQuestionHandler(u QuestionUsecase) *QuestionHandler {
	return &QuestionHandler{usecase: u}
}

// @Summary Get Questions
// @Description Mengambil pertanyaan essay pada lowongan
// @Tags Questions
// @Produce json
// @Param job_id query string true "Job ID"
// @Success 200 {object} SwaggerQuestionsResponse "Sukses mengambil data"
// @Failure 400 {object} response.SwaggerErrorResponse "Format Job ID tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/questions [get]
func (h *QuestionHandler) GetQuestions(c fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Parameter Job ID tidak boleh kosong",
			Data:    nil,
		})
	}

	qs, err := h.usecase.GetQuestionsByJob(c.Context(), jobID)
	if err != nil {
		slog.Error("gagal mengambil pertanyaan", "error", err, "job_id", jobID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerQuestionsResponse{
		Success: true,
		Message: "Sukses mengambil data",
		Data:    qs,
	})
}

// @Summary Update Question
// @Description Mengubah pertanyaan essay
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Param request body DTOUpdateQuestion true "Question update payload"
// @Success 200 {object} response.BaseResponse "Pertanyaan berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/questions/{id} [put]
func (h *QuestionHandler) UpdateQuestion(c fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Parameter Job ID tidak boleh kosong",
			Data:    nil,
		})
	}

	var req []DTOUpdateQuestion
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format data tidak valid",
			Data:    nil,
		})
	}

	if err := h.usecase.EditQuestion(c.Context(), jobID, req); err != nil {
		slog.Error("gagal mengubah pertanyaan", "error", err, "job_id", jobID)

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(response.BaseResponse{
				Success: false,
				Message: fiberErr.Message,
				Data:    nil,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Pertanyaan berhasil diperbarui",
		Data:    nil,
	})
}

// @Summary Delete Question
// @Description Menghapus pertanyaan essay
// @Tags Questions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} response.BaseResponse "Pertanyaan berhasil dihapus"
// @Failure 400 {object} response.SwaggerErrorResponse "Invalid question ID"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/questions/{id} [delete]
func (h *QuestionHandler) DeleteQuestion(c fiber.Ctx) error {
	questionID := c.Params("id")

	if err := h.usecase.RemoveQuestion(c.Context(), questionID); err != nil {
		slog.Error("gagal menghapus pertanyaan", "error", err, "question_id", questionID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Pertanyaan berhasil dihapus", Data: nil})
}
