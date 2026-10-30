package fileuploads

import (
	"log/slog"
	"strings"

	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/gofiber/fiber/v3"
)

type FileUploadsHandler struct {
	storageClient storage.SupabaseStorage
}

func NewFileUploadsHandler(sc storage.SupabaseStorage) *FileUploadsHandler {
	return &FileUploadsHandler{
		storageClient: sc,
	}
}

// @Summary Upload PDF Document
// @Description Mengunggah dokumen PDF
// @Tags File Uploads
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param document formData file true "PDF document"
// @Success 200 {object} SwaggerUploadResponse "Document uploaded successfully"
// @Failure 400 {object} response.BaseResponse "Invalid document"
// @Failure 500 {object} response.BaseResponse "Upload failed"
// @Router /api/v1/upload/document [post]
func (h *FileUploadsHandler) UploadDocument(c fiber.Ctx) error {
	file, err := c.FormFile("document")
	if err != nil {
		slog.Warn("document upload rejected: file is missing",
			"field", "document",
			"error", err,
		)
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Missing uploaded document. Please provide a PDF file using the 'document' form field.",
			Data:    nil,
		})
	}

	fileName := strings.ToLower(file.Filename)
	if !strings.HasSuffix(fileName, ".pdf") {
		slog.Warn("document upload rejected: unsupported file type",
			"file_name", file.Filename,
			"content_type", file.Header.Get("Content-Type"),
		)
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Invalid document type. Only PDF files are allowed.", Data: nil})
	}

	fileURL, err := h.storageClient.UploadFile(file, "resumes")
	if err != nil {
		slog.Error("document upload failed",
			"file_name", file.Filename,
			"file_size", file.Size,
			"error", err,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: "Failed to upload document. Please try again later.", Data: nil})
	}

	slog.Info("document uploaded successfully",
		"file_name", file.Filename,
		"file_size", file.Size,
		"storage_folder", "resumes",
	)
	return c.Status(fiber.StatusOK).JSON(SwaggerUploadResponse{
		Success: true,
		Message: "Document uploaded successfully.",
		Data:    UploadResponseData{FileURL: fileURL, FileName: file.Filename, FileSize: file.Size},
	})
}

// @Summary Upload Picture
// @Description Mengunggah gambar profil
// @Tags File Uploads
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "Image file"
// @Success 200 {object} SwaggerUploadResponse "Image uploaded successfully"
// @Failure 400 {object} response.BaseResponse "Invalid image"
// @Failure 500 {object} response.BaseResponse "Upload failed"
// @Router /api/v1/upload/picture [post]
func (h *FileUploadsHandler) UploadPicture(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		slog.Warn("image upload rejected: file is missing",
			"field", "image",
			"error", err,
		)
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Missing uploaded image. Please provide a PNG, JPG, or JPEG file using the 'image' form field.", Data: nil})
	}

	fileName := strings.ToLower(file.Filename)
	if !strings.HasSuffix(fileName, ".png") && !strings.HasSuffix(fileName, ".jpg") && !strings.HasSuffix(fileName, ".jpeg") {
		slog.Warn("image upload rejected: unsupported file type",
			"file_name", file.Filename,
			"content_type", file.Header.Get("Content-Type"),
		)
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Invalid image type. Only PNG, JPG, and JPEG files are allowed.", Data: nil})
	}

	fileURL, err := h.storageClient.UploadFile(file, "images")
	if err != nil {
		slog.Error("image upload failed",
			"file_name", file.Filename,
			"file_size", file.Size,
			"error", err,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: "Failed to upload image. Please try again later.", Data: nil})
	}

	slog.Info("image uploaded successfully",
		"file_name", file.Filename,
		"file_size", file.Size,
		"storage_folder", "images",
	)
	return c.Status(fiber.StatusOK).JSON(SwaggerUploadResponse{
		Success: true,
		Message: "Image uploaded successfully.",
		Data:    UploadResponseData{FileURL: fileURL, FileName: file.Filename, FileSize: file.Size},
	})
}
