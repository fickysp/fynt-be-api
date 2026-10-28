package auth

import (
	"errors"
	"log/slog"
	"time"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUsecase AuthUsecase
	validator   *validation.StructValidator
	cfg         *config.Config
}

func NewAuthHandler(au AuthUsecase, val *validation.StructValidator, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authUsecase: au,
		validator:   val,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req DTOLoginRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind login request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid JSON format",
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for login request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid data format",
		})
	}

	resp, err := h.authUsecase.Login(
		c.Context(),
		h.cfg.JWTSecret,
		h.cfg.RefreshTokenSecret,
		req,
	)

	if err != nil {
		slog.Error("Failed to login", "error", err.Error())
		if errors.Is(err, sentinel.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrInvalidCredentials,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrInternalServerError,
			})
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    resp.AccessToken,
		HTTPOnly: true,
		Secure:   true,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login success",
		"data":    resp,
	})
}

func (h *AuthHandler) RegisterSeeker(c fiber.Ctx) error {
	var req DTOApplicantRegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid JSON format",
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid data format",
		})
	}

	if err := h.authUsecase.RegisterApplicant(c.Context(), req); err != nil {
		slog.Error("Failed to register", "error", err.Error())
		if errors.Is(err, sentinel.ErrEmailIsUsed) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrEmailIsUsed,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrInternalServerError,
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Register Success",
	})
}

func (h *AuthHandler) RegisterRecruiter(c fiber.Ctx) error {
	var req DTOCompanyRegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid JSON format",
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid data format",
		})
	}

	if err := h.authUsecase.RegisterCompany(c.Context(), req); err != nil {
		slog.Error("Failed to register", "error", err.Error())
		if errors.Is(err, sentinel.ErrEmailIsUsed) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrEmailIsUsed,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": sentinel.ErrInternalServerError,
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Register Success",
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	accountID, ok := c.Locals("account_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}
	if err := h.authUsecase.Logout(c.Context(), accountID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to process logout",
		})
	}

	// Hancurkan cookie di peramban klien dengan mengatur Expires ke masa lalu (-1 jam)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}
