package jobs

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/response"
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

// @Summary Get All Categories
// @Description Mengambil seluruh kategori pekerjaan
// @Tags Jobs
// @Produce json
// @Success 200 {object} SwaggerJobCategoriesResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/categories [get]
func (h *JobsHandler) GetAllCategories(c fiber.Ctx) error {
	data, err := h.jobsUsecase.GetAllCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}
	return c.Status(fiber.StatusOK).JSON(SwaggerJobCategoriesResponse{Success: true, Message: "Sukses mengambil data", Data: data})
}

// @Summary Get All Jobs
// @Description Mengambil seluruh lowongan pekerjaan
// @Tags Jobs
// @Produce json
// @Success 200 {object} SwaggerJobsResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/jobs [get]
func (h *JobsHandler) GetAllJob(c fiber.Ctx) error {
	data, err := h.jobsUsecase.GetAllJob(c.Context())
	if err != nil {
		slog.Warn("cannot get all job", "warn", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.SwaggerErrorResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
	}
	return c.Status(fiber.StatusOK).JSON(SwaggerJobsResponse{Success: true, Message: "Sukses mengambil data", Data: data})
}

// @Summary Get Job By ID
// @Description Mengambil detail lowongan berdasarkan ID
// @Tags Jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} SwaggerJobResponse "Sukses mengambil data"
// @Failure 404 {object} response.SwaggerErrorResponse "Job not found"
// @Router /api/v1/job/{id} [get]
func (h *JobsHandler) GetJobByID(c fiber.Ctx) error {
	jobID := c.Params("id")
	job, err := h.jobsUsecase.GetJobByID(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.SwaggerErrorResponse{Success: false, Message: err.Error(), Data: nil})
	}
	return c.Status(fiber.StatusOK).JSON(SwaggerJobResponse{Success: true, Message: "Sukses mengambil data", Data: job})
}

// @Summary Create Job
// @Description Membuat lowongan pekerjaan
// @Tags Jobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOJobRequest true "Job payload"
// @Success 201 {object} response.BaseResponse "Lowongan berhasil dibuat"
// @Failure 400 {object} response.SwaggerErrorResponse "Format JSON atau data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/job [post]
func (h *JobsHandler) CreateJob(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	profileID := c.Locals("profile_id").(string)

	var req DTOJobRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.SwaggerErrorResponse{Success: false, Message: "Format JSON tidak valid", Data: nil})
	}

	req.CompanyID = profileID

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.SwaggerErrorResponse{Success: false, Message: err.Error(), Data: nil})
	}

	if err := h.jobsUsecase.CreateJob(c.Context(), accountID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.SwaggerErrorResponse{Success: false, Message: err.Error(), Data: nil})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Lowongan berhasil dibuat", Data: nil})
}

// @Summary Update Job
// @Description Mengubah lowongan pekerjaan
// @Tags Jobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Param request body DTOUpdateJobRequest true "Job update payload"
// @Success 200 {object} response.BaseResponse "Lowongan berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "Format JSON atau data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/job/{id} [patch]
func (h *JobsHandler) UpdateJob(c fiber.Ctx) error {
	jobID := c.Params("id")

	accountID, ok := c.Locals("account_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.SwaggerErrorResponse{
				Success: false,
				Message: "Account ID tidak valid",
			},
		)
	}

	profileID, ok := c.Locals("profile_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.SwaggerErrorResponse{
				Success: false,
				Message: "Profile ID tidak valid",
			},
		)
	}

	var req DTOJobUpdate

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.SwaggerErrorResponse{
				Success: false,
				Message: "Format JSON tidak valid",
			},
		)
	}

	req.CompanyID = &profileID

	if err := h.validator.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.SwaggerErrorResponse{
				Success: false,
				Message: err.Error(),
			},
		)
	}

	if err := h.jobsUsecase.UpdateJob(
		c.Context(),
		jobID,
		accountID,
		&req,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.SwaggerErrorResponse{
				Success: false,
				Message: err.Error(),
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		response.BaseResponse{
			Success: true,
			Message: "Lowongan berhasil diperbarui",
			Data:    nil,
		},
	)
}

// @Summary Search Jobs
// @Description Mencari lowongan pekerjaan
// @Tags Jobs
// @Produce json
// @Param q query string false "Search keyword"
// @Success 200 {object} SwaggerJobSearchResponse "Sukses mengambil data"
// @Failure 400 {object} response.SwaggerErrorResponse "Kata kunci pencarian minimal 3 karakter"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal mencari lowongan"
// @Router /api/v1/job/search [get]
func (h *JobsHandler) SearchJobs(c fiber.Ctx) error {
	keyword := c.Query("q")

	if len(keyword) > 0 && len(keyword) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(response.SwaggerErrorResponse{Success: false, Message: "Kata kunci pencarian minimal 3 karakter", Data: nil})
	}

	jobs, err := h.jobsUsecase.SearchJobs(c.Context(), keyword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.SwaggerErrorResponse{Success: false, Message: "Gagal mencari lowongan", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerJobSearchResponse{Success: true, Message: "Sukses mengambil data", Data: jobs})
}

// @Summary Get Job Summary
// @Description Mengambil statistik dashboard recruiter
// @Tags Recruiter
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerJobSummaryResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal mengambil statistik dasbor"
// @Router /api/v1/recruiter/dashboard/stats [get]
func (h *JobsHandler) GetJobSummary(c fiber.Ctx) error {
	compID := c.Locals("profile_id").(string)

	stats, err := h.jobsUsecase.GetJobSummary(c.Context(), compID)
	if err != nil {
		slog.Warn("ERROR", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.SwaggerErrorResponse{Success: false, Message: "Gagal mengambil statistik dasbor", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerJobSummaryResponse{Success: true, Message: "Sukses mengambil data", Data: stats})
}
