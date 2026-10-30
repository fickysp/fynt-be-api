package company

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type CompanyHandler struct {
	companyUsecase CompanyUsecase
	validator      *validation.StructValidator
	cfg            *config.Config
}

func NewCompanyHandler(cu CompanyUsecase, val *validation.StructValidator, cfg *config.Config) *CompanyHandler {
	return &CompanyHandler{
		companyUsecase: cu,
		validator:      val,
		cfg:            cfg,
	}
}

// @Summary Create Company Profile
// @Description Membuat profile perusahaan recruiter
// @Tags Recruiter
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOCompanyRegisterRequest true "Payload untuk membuat profile company baru"
// @Success 201 {object} response.BaseResponse "Company registered successfully"
// @Failure 400 {object} response.SwaggerErrorResponse "Body request tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/recruiter/on-boarding [post]
func (h *CompanyHandler) CreateCompanyProfile(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	var req DTOCompanyRegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Body request tidak valid",
			Data:    nil,
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Input invalid, check again your data",
			Data:    nil,
		})
	}

	if err := h.companyUsecase.RegisterCompany(c.Context(), accountID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{
		Success: true,
		Message: "Company registered successfully",
		Data:    nil,
	})
}

// @Summary Update Company Profile Data
// @Description Memperbarui data teks profil perusahaan
// @Tags Company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOCompanyUpdateRequest true "Data update profil perusahaan"
// @Success 200 {object} response.BaseResponse "Profil perusahaan berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "Validasi input tidak valid"
// @Failure 403 {object} response.SwaggerErrorResponse "Invalid user id"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/company/profile [patch]
func (h *CompanyHandler) UpdateProfile(c fiber.Ctx) error {
	profID, ok := c.Locals("profile_id").(string)
	if !ok || profID == "" {
		return c.Status(fiber.StatusForbidden).JSON(response.BaseResponse{
			Success: false,
			Message: "Invalid user id",
			Data:    nil,
		})
	}

	var req DTOCompanyUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Invalid form request",
			Data:    nil,
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Input invalid, check again your data",
			Data:    nil,
		})
	}

	if err := h.companyUsecase.UpdateProfile(c.Context(), profID, req); err != nil {
		slog.Error("gagal update company profile", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Profil perusahaan berhasil diperbarui",
		Data:    nil,
	})
}

// @Summary Update Company Avatar / Logo
// @Description Mengunggah atau memperbarui logo/avatar perusahaan
// @Tags Company
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "File logo/avatar perusahaan (JPG, PNG, WEBP)"
// @Success 200 {object} response.BaseResponse "Logo perusahaan berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "File logo/avatar wajib dilampirkan"
// @Failure 403 {object} response.SwaggerErrorResponse "Invalid user id"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/company/profile/avatar [patch]
func (h *CompanyHandler) UpdateAvatar(c fiber.Ctx) error {
	profID, ok := c.Locals("profile_id").(string)
	if !ok || profID == "" {
		return c.Status(fiber.StatusForbidden).JSON(response.BaseResponse{
			Success: false,
			Message: "Invalid user id",
			Data:    nil,
		})
	}

	avatar, err := c.FormFile("avatar")
	if err != nil || avatar == nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "File logo/avatar wajib dilampirkan",
			Data:    nil,
		})
	}

	url, err := h.companyUsecase.UpdateAvatar(c.Context(), profID, avatar)
	if err != nil {
		slog.Error("gagal update company avatar", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Logo perusahaan berhasil diperbarui",
		Data:    fiber.Map{"avatar_url": url},
	})
}
