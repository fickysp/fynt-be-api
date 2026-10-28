package submissions

import (
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type SubmissionsHandler struct {
	submissionUsecase SubmissionUsecase
	validator         *validation.StructValidator
}

func NewSubmissionsHandler(su SubmissionUsecase, val *validation.StructValidator) *SubmissionsHandler {
	return &SubmissionsHandler{
		submissionUsecase: su,
		validator:         val,
	}
}
func (h *SubmissionsHandler) ApplyJob(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	applicantID := c.Locals("profile_id").(string)

	var req ApplyJobRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	if err := h.submissionUsecase.ApplyJob(c.Context(), accountID, applicantID, req); err != nil {
		status := fiber.StatusInternalServerError

		if err.Error() == "insufficient coin balance to apply" {
			status = fiber.StatusPaymentRequired
		} else if err.Error() == "job not found" || err.Error() == "invalid applicant identity" {
			status = fiber.StatusNotFound
		} else if err.Error() == "job is already closed" {
			status = fiber.StatusConflict
		}

		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Job application submitted successfully",
	})
}
