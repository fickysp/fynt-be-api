package middleware

import (
	"log/slog"
	"strings"

	"github.com/fickysp/fynt-be-api/config"
	myjwt "github.com/fickysp/fynt-be-api/pkg/jwt"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware interface {
	RequireToken() fiber.Handler
	RequireRole(allowedRoles ...string) fiber.Handler
}

type authMiddleware struct {
	cfg *config.Config
}

func NewAuthMiddleware(cfg *config.Config) AuthMiddleware {
	return &authMiddleware{
		cfg: cfg,
	}
}

func (m *authMiddleware) RequireToken() fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			slog.Info("access_token tidak ditemukan")
			return c.Status(fiber.StatusUnauthorized).JSON(response.SwaggerErrorResponse{
				Success: false,
				Message: "Missing authorization token",
				Data:    nil,
			})
		}

		token, err := jwt.ParseWithClaims(tokenString, &myjwt.Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(m.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			slog.Error("Penyebab error", "error", err)
			slog.Error("Penyebab error", "error", err, "valid", token.Valid)
			return c.Status(fiber.StatusUnauthorized).JSON(response.SwaggerErrorResponse{
				Success: false,
				Message: "Token invalid or expired",
				Data:    nil,
			})
		}

		claims, ok := token.Claims.(*myjwt.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(response.SwaggerErrorResponse{
				Success: false,
				Message: "Failed to parse token claims",
				Data:    nil,
			})
		}

		if claims.Type != "access_token" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.SwaggerErrorResponse{
				Success: false,
				Message: "Invalid token type",
				Data:    nil,
			})
		}

		c.Locals("account_id", claims.ID)
		c.Locals("role", claims.Role)
		c.Locals("profile_id", claims.ProfileID)

		return c.Next()
	}
}

func (m *authMiddleware) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(response.SwaggerErrorResponse{Success: false, Message: "User role not found in session", Data: nil})
		}

		for _, role := range allowedRoles {
			if role == strings.ToLower(userRole) {
				return c.Next()
			}
		}

		slog.Info("users role: ", "user_role", userRole)
		slog.Info("role allowed: ", "user_role", allowedRoles)
		return c.Status(fiber.StatusForbidden).JSON(response.SwaggerErrorResponse{Success: false, Message: "Access denied for this role", Data: nil})
	}
}
