package wallet

import (
	"log/slog"

	"github.com/fickysp/fynt-be-api/config"
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

func (h *WalletHandler) TopUp(c fiber.Ctx) error {
	accountID := c.Locals("account_id").(string)
	role := c.Locals("role").(string)

	var req DTOTopUpRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON payload",
		})
	}

	req.AccountID = accountID

	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	resp, err := h.walletUsecase.RequestTopUp(c.Context(), role, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "minimum top-up amount is 10000" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Top-up request created",
		"data":    resp,
	})
}

func (h *WalletHandler) GetBalance(c fiber.Ctx) error {
	accID := c.Locals("account_id").(string)
	balance, err := h.walletUsecase.GetBalance(c.Context(), accID)
	if err != nil {
		slog.Warn("cannot get balance", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": sentinel.ErrInternalServerError,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    balance,
	})
}
