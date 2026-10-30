package experiences

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/gofiber/fiber/v3"
)

type ExperienceHandler struct {
	usecase ExperienceUsecase
}

func NewExperienceHandler(u ExperienceUsecase) *ExperienceHandler {
	return &ExperienceHandler{usecase: u}
}

// @Summary Add Experience
// @Description Menambahkan pengalaman kerja seeker
// @Tags Experience
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOExperienceRequest true "Experience payload"
// @Success 201 {object} response.BaseResponse "Experience added"
// @Failure 400 {object} response.SwaggerErrorResponse "Invalid JSON format"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/experiences [post]
func (h *ExperienceHandler) AddExperience(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	var req DTOExperienceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Invalid JSON format",
			Data:    nil,
		})
	}

	if err := h.usecase.AddExperience(c.Context(), profID, req); err != nil {
		slog.Error("gagal menambah pengalaman", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Experience added successfully", Data: nil})
}

// @Summary Delete Experience
// @Description Menghapus pengalaman kerja seeker
// @Tags Experience
// @Produce json
// @Security BearerAuth
// @Param id path string true "Experience ID"
// @Success 200 {object} response.BaseResponse "Experience removed"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/experiences/{id} [delete]
func (h *ExperienceHandler) DeleteExperience(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	expID := c.Params("id")

	if err := h.usecase.RemoveExperience(c.Context(), profID, expID); err != nil {
		slog.Error("gagal menghapus pengalaman", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Experience removed successfully", Data: nil})
}

// @Summary Update Experience
// @Description Mengubah pengalaman kerja seeker
// @Tags Experience
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Experience ID"
// @Param request body DTOUpdateExperienceRequest true "Experience update payload"
// @Success 200 {object} response.BaseResponse "Experience updated"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/experiences/{id} [patch]
func (h *ExperienceHandler) UpdateExperience(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	expID := c.Params("id")

	var req DTOUpdateExperienceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format data tidak valid", Data: nil})
	}

	if err := h.usecase.UpdateExperience(c.Context(), profID, expID, req); err != nil {
		slog.Error("gagal mengubah pengalaman", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Pengalaman berhasil diperbarui", Data: nil})
}

// @Summary Get Experiences
// @Description Mengambil daftar pengalaman kerja seeker
// @Tags Experience
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerExperiencesResponse "Get data success"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/experiences [get]
func (h *ExperienceHandler) GetExperiences(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	exps, err := h.usecase.GetExperiences(c.Context(), profID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerExperiencesResponse{Success: true, Message: "Get data success", Data: exps})
}
