package documents

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/gofiber/fiber/v3"
)

type DocumentHandler struct {
	usecase DocumentUsecase
}

func NewDocumentHandler(u DocumentUsecase) *DocumentHandler {
	return &DocumentHandler{usecase: u}
}

// @Summary Upload Document
// @Description Mengunggah dokumen seeker
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param documents_type formData string true "Document type" Enums(cv,portfolio)
// @Param file formData file true "Document file"
// @Success 201 {object} response.BaseResponse "Document uploaded"
// @Failure 400 {object} response.SwaggerErrorResponse "Invalid document request"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/documents [post]
func (h *DocumentHandler) UploadDocument(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	docType := c.FormValue("documents_type")
	if docType != "cv" && docType != "portfolio" {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Tipe dokumen harus 'cv' atau 'portfolio'",
			Data:    nil,
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Gagal membaca berkas", Data: nil})
	}

	if err := h.usecase.UploadDocument(c.Context(), profID, docType, fileHeader); err != nil {
		slog.Error("gagal mengupload dokumen", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Dokumen berhasil diunggah", Data: nil})
}

// @Summary Get Documents
// @Description Mengambil seluruh dokumen seeker
// @Tags Documents
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerDocumentsResponse "Get data success"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/documents [get]
func (h *DocumentHandler) GetAllDocuments(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	docs, err := h.usecase.GetDocuments(c.Context(), profID)
	if err != nil {
		slog.Error("gagal mengambil dokumen", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerDocumentsResponse{Success: true, Message: "Get data success", Data: docs})
}

// @Summary Set Active Document
// @Description Menetapkan dokumen aktif seeker
// @Tags Documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Param request body DTOSetActiveDocumentRequest true "Payload dokumen aktif"
// @Success 200 {object} response.BaseResponse "Document activated"
// @Failure 400 {object} response.SwaggerErrorResponse "Format data tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/documents/{id}/set-active [patch]
func (h *DocumentHandler) SetActiveDocument(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	docID := c.Params("id")

	var req DTOSetActiveDocumentRequest
	if err := c.Bind().Body(&req); err != nil || (req.DocumentsType != "cv" && req.DocumentsType != "portfolio") {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format data tidak valid", Data: nil})
	}

	if err := h.usecase.SetActiveDocument(c.Context(), profID, docID, req.DocumentsType); err != nil {
		slog.Error("gagal mengubah status dokumen", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Dokumen berhasil diaktifkan", Data: nil})
}

// @Summary Delete Document
// @Description Menghapus dokumen seeker
// @Tags Documents
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Success 200 {object} response.BaseResponse "Document deleted"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/seeker/documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c fiber.Ctx) error {
	profID := c.Locals("profile_id").(string)
	docID := c.Params("id")

	if err := h.usecase.DeleteDocument(c.Context(), profID, docID); err != nil {
		slog.Error("gagal menghapus dokumen", "error", err, "profile_id", profID)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Dokumen berhasil dihapus", Data: nil})
}
