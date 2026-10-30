package submissions

import (
	"log/slog"
	"strings"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
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

// @Summary Apply Job
// @Description Mengirim lamaran pekerjaan
// @Tags Submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Param request body ApplyJobRequest true "Application payload"
// @Success 201 {object} response.BaseResponse "Berhasil melamar pekerjaan"
// @Failure 400 {object} response.SwaggerErrorResponse "Format JSON atau data tidak valid"
// @Failure 402 {object} response.SwaggerErrorResponse "Koin tidak cukup"
// @Failure 404 {object} response.SwaggerErrorResponse "Lowongan atau user tidak ditemukan"
// @Failure 409 {object} response.SwaggerErrorResponse "Lowongan sudah ditutup"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/job/{id}/apply [post]
func (h *SubmissionsHandler) ApplyJob(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	applicantID := c.Locals("profile_id").(string)

	var req ApplyJobRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind apply job request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format JSON invalid",
			Data:    nil,
		})
	}

	pathJobID := c.Params("id")
	if req.JobID != "" && req.JobID != pathJobID {
		slog.Warn("job id mismatch between path and body", "path_job_id", pathJobID, "body_job_id", req.JobID)
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Job ID pada path dan body tidak sama",
			Data:    nil,
		})
	}
	req.JobID = pathJobID

	if err := h.validator.Validate(req); err != nil {
		slog.Error("Validation failed for apply job request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrValidate,
			Data:    nil,
		})
	}

	if err := h.submissionUsecase.ApplyJob(c.Context(), accountID, applicantID, req); err != nil {
		slog.Error("gagal melamar pekerjaan", "error", err, "job_id", req.JobID, "applicant_id", applicantID)
		status := fiber.StatusInternalServerError
		message := sentinel.ErrInternalServerError.Error()

		if err.Error() == "Koin tidak cukup, silahkan isi ulang koin terlebih dahulu" {
			status = fiber.StatusPaymentRequired
			message = err.Error()
		} else if err.Error() == "Lowongan tidak ditemukan" || err.Error() == "User tidak diketahui, silahkan login kembali" {
			status = fiber.StatusNotFound
			message = err.Error()
		} else if err.Error() == "Lowongan sudah ditutup" {
			status = fiber.StatusConflict
			message = err.Error()
		} else if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			status = fiber.StatusConflict
			message = "Anda sudah melamar pekerjaan ini"
		}

		return c.Status(status).JSON(response.SwaggerErrorResponse{Success: false, Message: message, Data: nil})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Berhasil melamar pekerjaan", Data: nil})
}

// @Summary Get My Submissions
// @Description Mengambil daftar lamaran seeker
// @Tags Submissions
// @Produce json
// @Security BearerAuth
// @Param status query string false "Submission status"
// @Param q query string false "Search keyword"
// @Success 200 {object} SwaggerSeekerSubmissionsResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal mengambil lamaran"
// @Router /api/v1/seeker/submissions [get]
func (h *SubmissionsHandler) GetMySubmissions(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	statusFilter := c.Query("status")
	keyword := c.Query("q")

	subs, err := h.submissionUsecase.GetMySubmissions(c.Context(), profID, statusFilter, keyword)
	if err != nil {
		slog.Error("gagal mengambil daftar lamaran seeker", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: "Gagal mengambil lamaran", Data: nil})
	} else if len(subs) == 0 {
		return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: false, Message: "Belum ada riwayat lamaran", Data: []any{}})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerSeekerSubmissionsResponse{Success: true, Message: "Sukses mengambil data", Data: subs})
}

// @Summary Get Submission Detail
// @Description Mengambil detail lamaran seeker
// @Tags Submissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Submission ID"
// @Success 200 {object} SwaggerSubmissionDetailResponse "Sukses mengambil data"
// @Failure 404 {object} response.SwaggerErrorResponse "Detail lamaran tidak ditemukan"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/submissions/{id} [get]
func (h *SubmissionsHandler) GetSubmissionDetail(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	submissionID := c.Params("id")

	detail, err := h.submissionUsecase.GetSubmissionDetail(c.Context(), profID, submissionID)
	if err != nil {
		slog.Error("gagal mengambil detail lamaran", "error", err, "profile_id", profID, "submission_id", submissionID)
		return c.Status(fiber.StatusNotFound).JSON(response.BaseResponse{Success: false, Message: "Detail lamaran tidak ditemukan", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerSubmissionDetailResponse{Success: true, Message: "Sukses mengambil data", Data: detail})
}

// @Summary Get Submissions By Job
// @Description Mengambil daftar pelamar pada lowongan recruiter
// @Tags Submissions
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "Job ID"
// @Param status query string false "Submission status"
// @Success 200 {object} SwaggerRecruiterSubmissionsResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal mengambil daftar pelamar"
// @Router /api/v1/submissions/job/{job_id} [get]
func (h *SubmissionsHandler) GetSubmissionsByJobID(c fiber.Ctx) error {
	compID := c.Locals("profile_id").(string)
	jobID := c.Params("job_id")

	statusFilter := c.Query("status")

	subs, err := h.submissionUsecase.GetSubmissionsByJobID(c.Context(), compID, jobID, statusFilter)
	if err != nil {
		slog.Error("gagal mengambil daftar pelamar", "error", err, "job_id", jobID, "company_id", compID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: "Gagal mengambil daftar pelamar", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerRecruiterSubmissionsResponse{Success: true, Message: "Sukses mengambil data", Data: subs})
}

// @Summary Update Submission Status
// @Description Mengubah status lamaran
// @Tags Submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Submission ID"
// @Param request body DTOUpdateStatusRequest true "Status update payload"
// @Success 200 {object} response.BaseResponse "Status lamaran berhasil diperbarui"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 403 {object} response.SwaggerErrorResponse "Lamaran tidak ditemukan, atau tidak punya akses"
// @Router /api/v1/submissions/{id}/status [patch]
func (h *SubmissionsHandler) UpdateSubmissionStatus(c fiber.Ctx) error {
	compID := c.Locals("profile_id").(string)
	submissionID := c.Params("id")

	var req DTOUpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind update status request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format data tidak valid", Data: nil})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for update status request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrValidate, Data: nil})
	}

	if err := h.submissionUsecase.UpdateSubmissionStatus(c.Context(), compID, submissionID, req.Status); err != nil {
		slog.Error("gagal update status lamaran", "error", err, "submission_id", submissionID, "company_id", compID)
		return c.Status(fiber.StatusForbidden).JSON(response.BaseResponse{Success: false, Message: "Lamaran tidak ditemukan, atau tidak punya akses", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Status lamaran berhasil diperbarui", Data: nil})
}

// @Summary Update Internal Notes
// @Description Menyimpan catatan internal lamaran
// @Tags Submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Submission ID"
// @Param request body DTOUpdateNotesRequest true "Internal notes payload"
// @Success 200 {object} response.BaseResponse "Catatan internal berhasil disimpan"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 403 {object} response.SwaggerErrorResponse "Lamaran tidak ditemukan, atau tidak memiliki akses"
// @Router /api/v1/submissions/{id}/notes [patch]
func (h *SubmissionsHandler) UpdateInternalNotes(c fiber.Ctx) error {
	compID := c.Locals("profile_id").(string)
	submissionID := c.Params("id")

	var req DTOUpdateNotesRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind update internal notes request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format data tidak valid", Data: nil})
	}

	if err := h.validator.Validate(req); err != nil {
		slog.Error("Validation failed for update internal notes request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrValidate, Data: nil})
	}

	if err := h.submissionUsecase.UpdateInternalNotes(c.Context(), compID, submissionID, req.InternalNotes); err != nil {
		slog.Error("gagal update catatan internal lamaran", "error", err, "submission_id", submissionID, "company_id", compID)
		return c.Status(fiber.StatusForbidden).JSON(response.BaseResponse{Success: false, Message: "Lamaran tidak ditemukan, atau tidak memiliki akses", Data: nil})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Catatan internal berhasil disimpan", Data: nil})
}

func (h *SubmissionsHandler) GetApplicantEssaysByJob(c fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.SwaggerErrorResponse{
			Success: false,
			Message: "ID Lowongan wajib ada",
			Data:    nil,
		})
	}

	data, err := h.submissionUsecase.GetApplicantEssaysByJob(c.Context(), jobID)
	if err != nil {
		slog.Error("Gagal mengambil data", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.SwaggerErrorResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerSubmissionEssay{
		Success: true,
		Message: "Berhasil mengambil daftar esai pelamar",
		Data:    data,
	})
}
