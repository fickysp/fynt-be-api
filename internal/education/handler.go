package education

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/gofiber/fiber/v3"
)

type EducationHandler struct {
	usecase EducationUsecase
}

func NewEducationHandler(u EducationUsecase) *EducationHandler {
	return &EducationHandler{usecase: u}
}

// @Summary Add Education
// @Description Menambahkan pendidikan seeker
// @Tags Education
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOEducationRequest true "Education payload"
// @Success 201 {object} response.BaseResponse "Education added"
// @Failure 400 {object} response.SwaggerErrorResponse "Invalid JSON format"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/educations [post]
func (h *EducationHandler) AddEducation(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)

	var req DTOEducationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Invalid JSON format",
			Data:    nil,
		})
	}

	if err := h.usecase.AddEducation(c.Context(), profID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Pendidikan berhasil ditambahkan", Data: nil})
}

// @Summary Delete Education
// @Description Menghapus pendidikan seeker
// @Tags Education
// @Produce json
// @Security BearerAuth
// @Param id path string true "Education ID"
// @Success 200 {object} response.BaseResponse "Education deleted"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/educations/{id} [delete]
func (h *EducationHandler) DeleteEducation(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	eduID := c.Params("id")

	if err := h.usecase.RemoveEducation(c.Context(), profID, eduID); err != nil {
		slog.Error("gagal menghapus pendidikan", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Pendidikan berhasil dihapus", Data: nil})
}

// @Summary Update Education
// @Description Mengubah pendidikan seeker
// @Tags Education
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Education ID"
// @Param request body DTOUpdateEducationRequest true "Education update payload"
// @Success 200 {object} response.BaseResponse "Education updated"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/educations/{id} [patch]
func (h *EducationHandler) UpdateEducation(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	eduID := c.Params("id")

	var req DTOUpdateEducationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format data tidak valid", Data: nil})
	}

	if err := h.usecase.UpdateEducation(c.Context(), profID, eduID, req); err != nil {
		slog.Error("gagal mengupdate pendidikan", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Pendidikan berhasil diperbarui", Data: nil})
}

// @Summary Get Educations
// @Description Mengambil daftar pendidikan seeker
// @Tags Education
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerEducationsResponse "Get data success"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/educations [get]
func (h *EducationHandler) GetEducations(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	edus, err := h.usecase.GetEducations(c.Context(), profID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerEducationsResponse{Success: true, Message: "Get data success", Data: edus})
}
