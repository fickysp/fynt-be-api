package fileuploads

import (
	"log/slog"
	"strings"

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

func (h *FileUploadsHandler) UploadDocument(c fiber.Ctx) error {
	file, err := c.FormFile("document")
	if err != nil {
		slog.Warn("document upload rejected: file is missing",
			"field", "document",
			"error", err,
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing uploaded document. Please provide a PDF file using the 'document' form field.",
		})
	}

	fileName := strings.ToLower(file.Filename)
	if !strings.HasSuffix(fileName, ".pdf") {
		slog.Warn("document upload rejected: unsupported file type",
			"file_name", file.Filename,
			"content_type", file.Header.Get("Content-Type"),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid document type. Only PDF files are allowed.",
		})
	}

	fileURL, err := h.storageClient.UploadFile(file, "resumes")
	if err != nil {
		slog.Error("document upload failed",
			"file_name", file.Filename,
			"file_size", file.Size,
			"error", err,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload document. Please try again later.",
		})
	}

	slog.Info("document uploaded successfully",
		"file_name", file.Filename,
		"file_size", file.Size,
		"storage_folder", "resumes",
	)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Document uploaded successfully.",
		"data": map[string]interface{}{
			"file_url":  fileURL,
			"file_name": file.Filename,
			"file_size": file.Size,
		},
	})
}
func (h *FileUploadsHandler) UploadPicture(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		slog.Warn("image upload rejected: file is missing",
			"field", "image",
			"error", err,
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing uploaded image. Please provide a PNG, JPG, or JPEG file using the 'image' form field.",
		})
	}

	fileName := strings.ToLower(file.Filename)
	if !strings.HasSuffix(fileName, ".png") && !strings.HasSuffix(fileName, ".jpg") && !strings.HasSuffix(fileName, ".jpeg") {
		slog.Warn("image upload rejected: unsupported file type",
			"file_name", file.Filename,
			"content_type", file.Header.Get("Content-Type"),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid image type. Only PNG, JPG, and JPEG files are allowed.",
		})
	}

	fileURL, err := h.storageClient.UploadFile(file, "images")
	if err != nil {
		slog.Error("image upload failed",
			"file_name", file.Filename,
			"file_size", file.Size,
			"error", err,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload image. Please try again later.",
		})
	}

	slog.Info("image uploaded successfully",
		"file_name", file.Filename,
		"file_size", file.Size,
		"storage_folder", "images",
	)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Image uploaded successfully.",
		"data": map[string]interface{}{
			"file_url":  fileURL,
			"file_name": file.Filename,
			"file_size": file.Size,
		},
	})
}
