package transaction

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type TransactionHandler struct {
	trxUsecase TransactionUsecase
	cfg        *config.Config
}

func NewTransactionHandler(tu TransactionUsecase, cfg *config.Config) *TransactionHandler {
	return &TransactionHandler{trxUsecase: tu, cfg: cfg}
}

// @Summary Midtrans Notification
// @Description Menerima notifikasi transaksi dari Midtrans
// @Tags Transaction
// @Accept json
// @Produce json
// @Param request body DTOMidtransNotification true "Midtrans notification payload"
// @Success 200 {object} response.BaseResponse "Sukses memproses notifikasi diproses"
// @Failure 400 {object} response.SwaggerErrorResponse "Payload notifikasi tidak valid"
// @Router /api/v1/transaction/notification [post]
func (h *TransactionHandler) MidtransNotification(c fiber.Ctx) error {
	var payload DTOMidtransNotification
	slog.Info("webhook midtrans masuk", "body", string(c.Body()))
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
			Success: false,
			Message: "Payload notifikasi tidak valid",
			Data:    nil,
		})
	}

	if err := h.trxUsecase.HandleNotification(c.Context(), h.cfg.MidtransServerKey, payload); err != nil {
		slog.Error("gagal memproses notifikasi midtrans", "error", err, "order_id", payload.OrderID)
		// Selalu return 200 agar Midtrans tidak retry
		return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Notifikasi berhasil diproses",
		Data:    nil,
	})
}
