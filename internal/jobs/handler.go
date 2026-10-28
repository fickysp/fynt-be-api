package jobs

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type JobsHandler struct {
	jobsUsecase JobUsecase
	validator   *validation.StructValidator
	cfg         *config.Config
}

func NewJobsHandler(ju JobUsecase, val *validation.StructValidator, c *config.Config) *JobsHandler {
	return &JobsHandler{
		jobsUsecase: ju,
		validator:   val,
		cfg:         c,
	}
}

func (h *JobsHandler) GetAllCategories(c fiber.Ctx) error {
	data, err := h.jobsUsecase.GetAllCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": sentinel.ErrInternalServerError,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}
func (h *JobsHandler) GetAllJob(c fiber.Ctx) error {
	data, err := h.jobsUsecase.GetAllJob(c.Context())
	if err != nil {
		slog.Warn("cannot get all job", "warn", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": sentinel.ErrInternalServerError,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (h *JobsHandler) GetJobByID(c fiber.Ctx) error {
	jobID := c.Params("id")
	job, err := h.jobsUsecase.GetJobByID(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": job})
}

func (h *JobsHandler) CreateJob(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	profileID := c.Locals("profile_id").(string) // Ini adalah CompanyID

	var req DTOJobRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format JSON tidak valid"})
	}

	req.CompanyID = profileID

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.jobsUsecase.CreateJob(c.Context(), accountID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Lowongan berhasil dibuat"})
}

func (h *JobsHandler) UpdateJob(c fiber.Ctx) error {
	jobID := c.Params("id")
	accountID := c.Locals("account_id").(string)
	profileID := c.Locals("profile_id").(string)

	var req DTOUpdateJobRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format JSON tidak valid"})
	}

	req.CompanyID = profileID

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.jobsUsecase.UpdateJob(c.Context(), accountID, jobID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Lowongan berhasil diperbarui"})
}
