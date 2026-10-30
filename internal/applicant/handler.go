package applicant

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type ApplicantHandler struct {
	applicantUsecase ApplicantUsecase
	val              *validation.StructValidator
}

func NewApplicantHandler(aU ApplicantUsecase, val *validation.StructValidator) *ApplicantHandler {
	return &ApplicantHandler{
		applicantUsecase: aU,
		val:              val,
	}
}

// @Summary Update Applicant Profile Data
// @Description Memperbarui data teks profil seeker
// @Tags Applicant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOApplicantUpdateRequest true "Data update profile"
// @Success 200 {object} response.BaseResponse "Sukses update data profile"
// @Failure 400 {object} response.SwaggerErrorResponse "Validasi input tidak sesuai"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/profile [patch]
func (h *ApplicantHandler) UpdateProfile(c fiber.Ctx) error {
	profID, ok := c.Locals("profile_id").(string)
	if !ok || profID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{
			Success: false,
			Message: "Unauthorized: profile_id tidak ditemukan",
			Data:    nil,
		})
	}

	var req DTOApplicantUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format request body tidak valid",
			Data:    nil,
		})
	}

	if err := h.applicantUsecase.UpdateProfile(c.Context(), profID, req); err != nil {
		slog.Error("gagal update profile", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Sukses update data profile",
		Data:    nil,
	})
}

// @Summary Update Applicant Avatar
// @Description Mengunggah atau memperbarui foto avatar seeker
// @Tags Applicant
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "File gambar avatar (JPG, PNG, WEBP)"
// @Success 200 {object} response.BaseResponse "Avatar berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "File tidak valid"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/profile/avatar [patch]
func (h *ApplicantHandler) UpdateAvatar(c fiber.Ctx) error {
	profID, ok := c.Locals("profile_id").(string)
	if !ok || profID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{
			Success: false,
			Message: "Unauthorized: profile_id tidak ditemukan",
			Data:    nil,
		})
	}

	avatar, err := c.FormFile("avatar")
	if err != nil || avatar == nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "File avatar wajib dilampirkan",
			Data:    nil,
		})
	}

	url, err := h.applicantUsecase.UpdateAvatar(c.Context(), profID, avatar)
	if err != nil {
		slog.Error("gagal update avatar", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Avatar berhasil diperbarui",
		Data:    fiber.Map{"avatar_url": url},
	})
}

// @Summary Get Profile Applicant/Job Seeker
// @Description Mengambil data detail profile pengguna yang sedang login
// @Tags Seeker
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerGetApplicantResponse "Data user berhasil diambil"
// @Failure 403 {object} response.SwaggerErrorResponse "ID user tidak valid"
// @Failure 404 {object} response.SwaggerErrorResponse "User tidak ditemukan"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/profile [get]
func (h *ApplicantHandler) GetApplicant(c fiber.Ctx) error {
	accID := c.Locals("account_id").(string)
	if accID == "" {
		return c.Status(fiber.StatusForbidden).JSON(response.BaseResponse{
			Success: false,
			Message: "ID user tidak valid",
			Data:    nil,
		})
	}

	app, err := h.applicantUsecase.GetApplicant(c.Context(), accID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(response.BaseResponse{
				Success: false,
				Message: "User tidak ditemukan",
				Data:    nil,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
				Success: false,
				Message: sentinel.ErrInternalServerError.Error(),
				Data:    nil,
			})
		}
	}

	resp := DTOApplicantProfileResponse{
		ID: app.ID.String(),
	}
	if app.FirstName != nil {
		resp.FirstName = *app.FirstName
	}
	if app.LastName != nil {
		resp.LastName = *app.LastName
	}
	if app.Gender != nil {
		resp.Gender = *app.Gender
	}
	if app.BirthDate != nil {
		resp.BirthDate = app.BirthDate.Format("02/01/2006")
	}
	if app.About != nil {
		resp.About = *app.About
	}
	if app.Domicile != nil {
		resp.Domicile = *app.Domicile
	}

	if app.AvatarURL != nil {
		resp.AvatarURL = *app.AvatarURL
	}

	if app.WebPortfolioURL != nil {
		resp.WebPortfolioURL = *app.WebPortfolioURL
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerGetApplicantResponse{
		Success: true,
		Message: "Data user berhasil diambil",
		Data:    resp,
	})
}
