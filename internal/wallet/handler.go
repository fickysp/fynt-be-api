package wallet

import (
	"errors"
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type WalletHandler struct {
	walletUsecase WalletUsecase
	validator     *validation.StructValidator
	cfg           *config.Config
}

func NewWalletHandler(wu WalletUsecase, val *validation.StructValidator, c *config.Config) *WalletHandler {
	return &WalletHandler{
		walletUsecase: wu,
		validator:     val,
		cfg:           c,
	}
}

// @Summary Top Up Wallet
// @Description Membuat permintaan top up wallet
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTOTopUpRequest true "Top-up payload"
// @Success 201 {object} SwaggerTopUpResponse "Top-up request created"
// @Failure 400 {object} response.SwaggerErrorResponse "Invalid request atau minimal top-up tidak terpenuhi"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/wallet/top-up [post]
func (h *WalletHandler) TopUp(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	role := c.Locals("role").(string)

	var req DTOTopUpRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format JSON tidak valid",
			Data:    nil,
		})
	}

	req.AccountID = accountID

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Jumlah topup minimal Rp 10,000",
			Data:    nil,
		})
	}

	resp, err := h.walletUsecase.RequestTopUp(c.Context(), role, req)
	if err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(response.BaseResponse{
				Success: false,
				Message: fiberErr.Message,
				Data:    nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{
		Success: true,
		Message: "Top-up request berhasil dibuat",
		Data:    resp,
	})
}

// @Summary Get Wallet Balance
// @Description Mengambil saldo wallet pengguna
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SwaggerBalanceResponse "Get data success"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/wallet/balance [get]
func (h *WalletHandler) GetBalance(c fiber.Ctx) error {
	accID := c.Locals("account_id").(string)
	balance, err := h.walletUsecase.GetBalance(c.Context(), accID)
	if err != nil {
		slog.Warn("cannot get balance", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: err.Error(), Data: nil})
	}
	return c.Status(fiber.StatusOK).JSON(SwaggerBalanceResponse{Success: true, Message: "Get data success", Data: balance})
}
