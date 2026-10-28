package transaction

import (
	"github.com/gofiber/fiber/v3"
)

type TransactionHandler struct {
	trxUsecase TransactionUsecase
}

func NewTransactionHandler(tu TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{trxUsecase: tu}
}

func (h *TransactionHandler) MidtransNotification(c fiber.Ctx) error {
	var payload DTOMidtransNotification
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid notification payload",
		})
	}

	if err := h.trxUsecase.HandleNotification(c.Context(), payload); err != nil {
		// always return code 200 OK
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Notification processed successfully",
	})
}
